package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/netsec-ethz/2SMS/common"
	"github.com/netsec-ethz/scion-apps/lib/scionutil"
	"github.com/netsec-ethz/scion-apps/lib/shttp"
	"github.com/scionproto/scion/go/lib/snet"
)

type scraperProxyHandler struct {
	ipClient    *http.Client
	scionClient *http.Client
	dnsMap      map[string]*snet.Addr
	enableQUIC  bool
}

func CreateScraperProxyHandler(scraperCACertsDir, scraperCert, scraperPrivKey string, localAddress *snet.Addr, enableQUIC bool) *scraperProxyHandler {
	ipClient := common.CreateHttpsClient(scraperCACertsDir, scraperCert, scraperPrivKey)
	dnsMap := make(map[string]*snet.Addr)
	scionClient := &http.Client{
		Transport: &shttp.Transport{
			LAddr: localAddress,
		},
	}
	return &scraperProxyHandler{ipClient: ipClient, scionClient: scionClient, enableQUIC: enableQUIC, dnsMap: dnsMap}
}

// When receiving an HTTP request try to forward it to its destination using HTTPS over SCION. Would an error occur
// try using HTTPS over IP.
func (sph *scraperProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var resp *http.Response
	var err error
	ip, port, err := net.SplitHostPort(r.URL.Host)
	if err != nil {
		log.Printf("Could not split request's host and port. Error is: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// To include the scion address in the Prometheus configuration file we added it as prefix to the target path
	// E.g. if we target path for 17-ffaa:1:11 is /metrics, then the path in the config file is /17-ffaa:1:11/metrics
	// Here we receive therefore a request to some IP address with that path and prior to forward it we need to remove the IA
	ia := strings.Split(r.URL.Path, "/")[1]
	// Remove IA from the target path to leave only the effective path
	// SplitN will split the path into 3 parts: "", "IA" and "<path>", we then take only the last one and prepend the slash
	r.URL.Path = "/" + strings.SplitN(r.URL.Path, "/", 3)[2]

	// If SQUIC is enabled try first with it
	if enableSQUIC {
		remoteAddr := fmt.Sprintf("%s,[%s]", ia, ip)
		// Hack to have shttp working (remove once scion addresses can be used directly in the url)
		// Replacing all colons in `ia` is necessary otherwise net.splitHostPort will return an incorrect result
		remoteAddrID := strings.Join([]string{strings.Replace(ia, ":", "_", -1), ip}, "_")
		scionutil.AddHost(remoteAddrID, remoteAddr)

		// This URL doesn't have to be safe because it is not used for DNS but only for scionutil.GetHostByName in scion-apps/lib/shttp/transport.go
		requestURL := remoteAddrID + ":" + port + r.URL.Path

		// Perform HTTP request using SCION client
		resp, err = sph.forwardRequest(true, w, requestURL, r)

		if err != nil {
			log.Printf("Failed: SCION/HTTPS request to %s. Error is: %v", requestURL, err)
		} else if resp.StatusCode == http.StatusNotFound {
			// If we couldn't find the path then we don't need to try with IP because it will lead to the same result.
			log.Printf("Failed: SCION/HTTPS request to %s. Path was not found (404).", requestURL)
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}
	// If SQUIC reported an error or it isn't enabled try with IP
	if err != nil || !enableSQUIC {
		if err != nil {
			log.Println("Target unreachable via SCION, falling back to IP.")
		}
		httpPort, err := strconv.Atoi(port)
		if err != nil {
			log.Printf("Cannot compute HTTPS port from SCION one. Error is %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		host := fmt.Sprintf("%s:%d", ip, httpPort+1)

		// Perform HTTP request using IP client
		resp, err = sph.forwardRequest(false, w, host+r.URL.Path, r)

		if err != nil {
			log.Printf("Failed: IP/HTTPS request to %s. Error is: %v", host+r.URL.Path, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else if resp.StatusCode == http.StatusNotFound {
			log.Printf("Failed: IP/HTTPS request to %s. Path was not found (404).", host+r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}

	// Copy response body
	io.Copy(w, resp.Body)
	err = resp.Body.Close()
	if err != nil {
		log.Printf("serveHTTP: could not close response's body after copying it for redirection. Error is: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Copy headers
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
}

func (sph *scraperProxyHandler) forwardRequest(overSCION bool, w http.ResponseWriter, url string, r *http.Request) (resp *http.Response, err error) {
	client := sph.scionClient
	if !overSCION {
		client = sph.ipClient
	}
	if r.Method == http.MethodGet {
		resp, err = client.Get(fmt.Sprintf("https://%s", url))
	} else if r.Method == http.MethodPost {
		resp, err = client.Post(fmt.Sprintf("https://%s", url), "application/x-protobuf", r.Body)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return nil, errors.New(fmt.Sprintf("Scrape proxy handler doesn't support method %s", r.Method))
	}
	return resp, err
}

package main

import (
	"net/http"
	"log"
	"crypto/tls"
	"os/exec"
	"flag"
	"encoding/gob"
	"encoding/json"
	"os"
	"time"

	"github.com/gorilla/mux"

	"github.com/netsec-ethz/2SMS/common"
	"github.com/netsec-ethz/2SMS/scraper/prometheus"

	"github.com/scionproto/scion/go/lib/snet"
	sd "github.com/scionproto/scion/go/lib/sciond"
	"bytes"
	"github.com/netsec-ethz/2SMS/common/types"
	"crypto/ecdsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"net"
	"io/ioutil"
	"fmt"
	"strings"
)

var (
	localhostManagementPort	string
	authDir			string = "auth"
	prometheusOutFile	string
	configManager		prometheus.ConfigManager
	prometheusExec     string
	prometheusConfig   string
	scraperCert        string
	scraperCSR			string
	scraperPrivKey     string
	scraperIP			string
	scraperDNS			string
	caCertsDir   string
	internalScrapePort string
	internalWritePort  string
	managementAPIPort  string
	managerIP		string
	managerUnverifPort			string
	managerVerifPort			string
	prometheusListenAddress	string
	prometheusRetention	string
	local       snet.Addr
	sciond      = flag.String("sciond", "", "Path to sciond socket")
	dispatcher  = flag.String("dispatcher", "/run/shm/dispatcher/default.sock",
		"Path to dispatcher socket")
	isdCoverage string
)

func init() {
	flag.StringVar(&prometheusOutFile, "prometheus.out", "prometheus/out", "file where prometheus output is redirected")
	flag.StringVar(&prometheusExec, "scraper.prometheus.exec", "prometheus/prometheus", "prometheus executable")
	flag.StringVar(&prometheusConfig, "scraper.prometheus.config", "prometheus/prometheus.yml", "prometheus configuration file")
	flag.StringVar(&scraperCert, "scraper.cert", "auth/scraper.crt", "full chain scraper's certificate file")
	flag.StringVar(&scraperPrivKey, "scraper.key", "auth/scraper.key", "scraper's private key file")
	flag.StringVar(&scraperCSR, "scraper.csr", "auth/scraper.csr", "csr for the key")
	flag.StringVar(&scraperDNS, "scraper.DNS", "localhost", "DNS name of scraper machine")
	flag.StringVar(&scraperIP, "scraper.IP", "127.0.0.1", "IP of scraper machine")

	flag.StringVar(&internalScrapePort, "scraper.ports.interal_scrape", "9901", "port the scraping proxy listens on localhost")
	flag.StringVar(&internalWritePort, "scraper.ports.interal_write", "9902", "port the writing proxy listens on localhost")
	flag.StringVar(&managementAPIPort, "scraper.ports.management", "9900", "port where the management API is exposed")
	flag.Var((*snet.Addr)(&local), "local", "(Mandatory) address to listen on")
	flag.StringVar(&prometheusListenAddress, "scraper.prometheus.address", "127.0.0.1:9090", "web.listen-address parameter for prometheus")
	flag.StringVar(&prometheusRetention, "scraper.prometheus.retention", "15d", "retention policy for prometheus server")
	flag.StringVar(&localhostManagementPort, "scraper.ports.local", "9999", "port where the local management API is exposed")

	flag.StringVar(&caCertsDir, "ca.certs", "ca_certs", "directory with trusted ca certificates")

	flag.StringVar(&managerIP, "manager.IP", "", "ip address of the managers")
	flag.StringVar(&managerUnverifPort, "manager.unverif-port", "10000", "port where manager listens for certificate request")
	flag.StringVar(&managerVerifPort, "manager.verif-port", "10001", "port where manager listens for authenticated operations")
	flag.StringVar(&isdCoverage, "scraper.coverage", "", "comma separated list of ISD numbers for which the scraper should accept targets")


	flag.Parse()

	gob.Register(types.Request{})
	gob.Register(types.Response{})

	// Create directory to store auth data
	if !common.FileExists(authDir) {
		os.Mkdir(authDir, 0700)  // The private key is stored here, so permissions are restrictive
	}
	if !common.FileExists(caCertsDir) {
		os.Mkdir(caCertsDir, 0700)
	}
	if *sciond == "" {
		*sciond = sd.GetDefaultSCIONDPath(nil)
	}
	common.InitNetwork(local, sciond, dispatcher)

	// Bootstrap PKI
	err := common.Bootstrap(caCertsDir + "/ca.crt", caCertsDir + "/bootstrap.json", local.IA)
	if err != nil {
		log.Fatal("Verification of ca certificate failed:", err)
	} else {
		log.Println("Successfully verified ca certificate.")
	}

	var privKey *ecdsa.PrivateKey
	if !common.FileExists(scraperPrivKey) {
		// Generate a new key and write it to file
		privKey, _ = common.GenECDSAKey("P256")
		bts, _ := x509.MarshalECPrivateKey(privKey)
		common.WriteToPEMFile(scraperPrivKey, "ECDSA PRIVATE KEY", bts)
		name := pkix.Name{
			Organization:  []string{"SCIONLab"},
			OrganizationalUnit: []string{"Scraper"},
			Country:       []string{"CH"},
			Province:      []string{"Zurich"},
			Locality:      []string{"Zurich"},
		}
		bts, _ = common.GenCertSignRequest(name, privKey, []string{scraperDNS}, []net.IP{net.ParseIP(scraperIP)})
		common.WriteToPEMFile(scraperCSR, "CERTIFICATE REQUEST", bts)
	}
	if !common.FileExists(scraperCert) {
		if managerIP != "" {
			common.RequestAndObtainCert(caCertsDir, managerIP, managerUnverifPort, scraperCert, scraperCSR, "Scraper", scraperIP)
		} else {
			log.Fatal("No certificate found and no connection with manager. Please manually generate and upload a certificate for the csr.")
		}
	}

	configManager = prometheus.ConfigManager{ConfigFile: prometheusConfig, ProxyURL: "http://127.0.0.1:" + internalScrapePort, ListenAddress: prometheusListenAddress}

	if isdCoverage == "" {
		isdCoverage = fmt.Sprint(local.IA.I)
	}

	// Register at manager
	if managerIP != "" {
		client := common.CreateHttpsClient(caCertsDir, scraperCert, scraperPrivKey)
		data, err := json.Marshal(types.Scraper{
			IA:         local.IA.String(),
			IP:         local.Host.IP().String(),
			ManagePort: managementAPIPort,
			ISDs: strings.Split(isdCoverage, ","),
		})
		if err != nil {
			log.Fatal("Failed marshalling Scraper struct:", err)
		}
		resp, err := client.Post("https://" + managerIP + ":" + managerVerifPort + "/scrapers/register", "application/json", bytes.NewBuffer(data))
		if err != nil {
			log.Fatal("Failed sending registration request:", err)
		}
		if resp.StatusCode != 204 {
			message, _ := ioutil.ReadAll(resp.Body)
			log.Fatal("Registration failed. Status code:", resp.StatusCode, "Message:", message)
		}
	}
}

func main() {
	// Spawn and manage prometheus server
	var proc *os.Process
	go func() {
		cmd := exec.Command("bash", "-c", prometheusExec + " --config.file=" + prometheusConfig + " --web.enable-lifecycle --storage.tsdb.retention=" + prometheusRetention + " --web.listen-address=" + prometheusListenAddress + " &> " + prometheusOutFile)
		err := cmd.Start()
		if err != nil {
			log.Fatal("Failed starting Prometheus command:", err)
		}
		proc = cmd.Process
		// Check if Prometheus process terminates within one seconds (usually meaning there is a problem in the config file)
		time.Sleep(1 * time.Second)
		out, err := exec.Command("bash", "-c", "ps -p " + fmt.Sprint(proc.Pid)).Output()
		if strings.Contains(string(out), "defunct") {
			log.Fatal("Failed starting Prometheus server.")
		}
		log.Println("Started Prometheus server as process ", cmd.Process.Pid)
	}()
	defer proc.Kill()

	// Proxy for scraping
	go func() {
		// Create HTTPS client
		client := common.CreateHttpsClient(caCertsDir, scraperCert, scraperPrivKey)
		// Create SCION client
		sclient := SCIONClient{local}

		// Start server listening on localhost only
		s := &http.Server{
			Addr:           "127.0.0.1:" + internalScrapePort,
			Handler:        &scraperProxyHandler{httpsClient: client, scionClient: &sclient},
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}
		log.Println("Starting scraping proxy server")
		log.Fatal("Scraping proxy server listening error: ", s.ListenAndServe())
	}()

	// Proxy for remote writing
	go func() {
		// Create HTTPS client
		client := common.CreateHttpsClient(caCertsDir, scraperCert, scraperPrivKey)
		// Create SCION client
		sclient := SCIONClient{local}

		// Start server listening on localhost only
		s := &http.Server{
			Addr:           "127.0.0.1:" + internalWritePort,
			Handler:        &scraperProxyHandler{httpsClient: client, scionClient: &sclient},
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}
		log.Println("Starting remote writing proxy server")
		log.Fatal("Remote writing proxy server listening error: ",s.ListenAndServe())
	}()

	// Management Server
	router := mux.NewRouter()
	router.HandleFunc("/targets", ListTargets).Methods("GET")
	router.HandleFunc("/targets", AddTarget).Methods("POST")
	router.HandleFunc("/targets", RemoveTarget).Methods("DELETE")
	router.HandleFunc("/storages", ListStorages).Methods("GET")
	router.HandleFunc("/storages", AddStorage).Methods("POST")
	router.HandleFunc("/storages", RemoveStorage).Methods("DELETE")

	go func() {
		srv := &http.Server{
			Addr:      "127.0.0.1:" + localhostManagementPort,
			Handler:   router,
		}
		log.Fatal("Localhost HTTP server listening error: ", srv.ListenAndServe())
	}()

	srv := common.CreateHttpsServer(caCertsDir, "", managementAPIPort, router, tls.RequireAndVerifyClientCert)
	log.Println("Starting management server")
	log.Fatal("Management server listening error: ", srv.ListenAndServeTLS(scraperCert, scraperPrivKey))
}
package main

import (
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/netsec-ethz/2SMS/common"
	"github.com/netsec-ethz/2SMS/common/types"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// Requires a CSR, verifies it's validity and, if it is allowed, generates and returns a certificate.
func requestCert(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received certificate request from %s\n", r.Host)
	if refuseSigning {
		log.Printf("Rejected certificate request from %s: signing is blocked\n", r.Host)
		w.WriteHeader(401)
		w.Write([]byte("Certificate request is blocked. Contact the administrator."))
		return
	}
	// Process csr in the request's body
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(400)
		log.Println("Failed reading body", err)
		return
	}
	csrBytes := make([]byte, len(data))
	dec, err := base64.StdEncoding.Decode(csrBytes, data)
	if err != nil {
		w.WriteHeader(400)
		log.Println("Failed decoding body", err)
		return
	}
	csrBytes = csrBytes[:dec]
	pemBlock, _ := pem.Decode(csrBytes)
	if pemBlock == nil {
		log.Println("Failed decoding csr:", err)
		w.WriteHeader(400)
		return
	}
	csr, err := x509.ParseCertificateRequest(pemBlock.Bytes)
	if err != nil {
		log.Println("Failed parsing csr:", err)
		w.WriteHeader(400)
		return
	}

	ip := csr.IPAddresses[0].String()
	OU := csr.Subject.OrganizationalUnit
	crtFile := approvedCertsDir + "/" + OU[0] + "_" + ip + ".crt"
	// If certificate for csr already exists, just return it
	if common.FileExists(crtFile) {
		log.Printf("Certificate for %s already exists\n", r.Host)
		byts, _ := ioutil.ReadFile(crtFile)
		// Encode it to base64 and write it to the response buffer
		data := make([]byte, base64.StdEncoding.EncodedLen(len(byts)))
		base64.StdEncoding.Encode(data, byts)
		w.Write(data)
		return
	}

	// TODO: add some sort of verification (e.g. registration token)
	// Create new certificate
	certBytes, err := ca.GenCertFromCSR(csr, &common.Duration{1, 0, 0})
	if err != nil {
		log.Println("Failed generating certificate:", err)
		w.WriteHeader(400)
		return
	}
	common.WriteToPEMFile(crtFile, "CERTIFICATE", certBytes)
	log.Printf("Successfully generated new certificate for %s\n", r.Host)
	byts, _ := ioutil.ReadFile(crtFile)
	// Encode it to base64 and write it to the response buffer
	data = make([]byte, base64.StdEncoding.EncodedLen(len(byts)))
	base64.StdEncoding.Encode(data, byts)
	w.Write(data)
}

// Returns the certificate for the requesting entity (if it exists).
func getCert(w http.ResponseWriter, r *http.Request) {
	log.Println("Certificate get received")
	vars := mux.Vars(r)
	crtFile := approvedCertsDir + "/" + vars["type"] + "_" + vars["ip"] + ".crt"
	if !common.FileExists(crtFile) {
		w.WriteHeader(404)
		w.Write([]byte("Certificate doesn't exists"))
	} else {
		byts, err := ioutil.ReadFile(crtFile)
		if err != nil {
			log.Println("Certificate should be present but is not:", err)
			w.WriteHeader(500)
			return
		}
		crt := make([]byte, base64.StdEncoding.EncodedLen(len(byts)))
		base64.StdEncoding.Encode(crt, byts)
		w.Write(crt)
	}
}

// Returns the list of all registered endpoints.
func listEndpoints(w http.ResponseWriter, r *http.Request) {
	log.Println("List endpoints received")
	jsonEnds, err := ioutil.ReadFile("endpoints.json")
	if err != nil {
		log.Println("Error while reading file:", err)
		w.WriteHeader(500)
		return
	}
	w.Write(jsonEnds)
}

// Returns the list of all registered scrapers.
func listScrapers(w http.ResponseWriter, r *http.Request) {
	log.Println("List scrapers received")
	jsonScrs, err := ioutil.ReadFile("scrapers.json")
	if err != nil {
		log.Println("Error while reading file:", err)
		w.WriteHeader(500)
		return
	}
	w.Write(jsonScrs)
}

// Receives a new path for some endpoint and adds it as a target at all scrapers
func notifyAddedMapping(w http.ResponseWriter, r *http.Request) {
	log.Println("Notify added mapping received")
	rBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	var target types.Target
	err = json.Unmarshal(rBytes, &target)
	if err != nil {
		log.Println("Failed marshalling json:", err)
		w.WriteHeader(400)
		return
	}

	// Add to scrapers and return addresses of scrapers for authorization purposes
	jsonScrapers, err := json.Marshal(addTargetToScrapers(&target, rBytes))
	if err != nil {
		log.Println("Failed marshaling json:", err)
		w.WriteHeader(500)
		return
	}
	w.Write(jsonScrapers)
}

// byts is the json binary encoding of target, used just to avoid encoding/decoding multiple times
func addTargetToScrapers(target *types.Target, byts []byte) []types.Scraper {
	addedTo := []types.Scraper{}
	// Add new target to each scraper
	for _, scr := range getScrapers() {
		if scr.Covers(target.ISD) {
			addTargetToScraper(byts, &scr)
			addedTo = append(addedTo, scr)
		}
	}
	return addedTo
}

// byts is the json binary encoding of the target, used just to avoid encoding/decoding multiple times
func addTargetToScraper(byts []byte, scraper *types.Scraper) {
	resp, err := httpsClient.Post("https://"+scraper.IP+":"+scraper.ManagePort+"/targets", "application/json", bytes.NewReader(byts))
	if err != nil {
		log.Println("Error in adding scraper target:", err)
	}
	err = resp.Body.Close()
	if err != nil {
		log.Printf("addTargetToScraper: failed to close response's body getting scraper's response: %v", err)
	}
}

// Receives a removed path for some endpoint and removes it from the targets of all scrapers
func notifyRemovedMapping(w http.ResponseWriter, r *http.Request) {
	log.Println("Notify removed mapping received")
	// Remove target from each scraper
	for _, scr := range getScrapers() {
		req, err := http.NewRequest("DELETE", "https://"+scr.IP+":"+scr.ManagePort+"/targets", r.Body)
		resp, err := httpsClient.Do(req)
		if err != nil {
			log.Println("Error in removing scraper target:", err)
		}
		io.Copy(w, resp.Body)
		err = resp.Body.Close()
		if err != nil {
			log.Printf("notifyRemovedMapping: failed to close response's body getting scraper's response: %v", err)
		}
	}
}

// Registers a new endpoint by writing it in the endpoints' list and creating targets for each mapping at responsible scrapers
func registerEndpoint(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading request body:", err)
		w.WriteHeader(400)
		return
	}
	var end types.Endpoint
	if err := json.Unmarshal(data, &end); err != nil {
		log.Println("Error unmarshalling json:", err)
		w.WriteHeader(400)
		return
	}
	err = addEndpoint(&end)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	scrapersToAuthorize := []types.Scraper{}
	for _, path := range end.Paths {
		// Build target endpoint's path
		target := types.Target{}
		target.AS = strings.Split(end.IA, "-")[1]
		target.ISD = strings.Split(end.IA, "-")[0]
		target.IP = end.IP
		target.Path = path
		target.Port = end.ScrapePort
		target.Labels = make(map[string]string)
		target.Labels["AS"] = target.AS
		target.Labels["ISD"] = target.ISD
		target.Labels["service"] = target.Path[1:] // Assumes path is of the form `/<service-name>`
		target.Name = target.Path[1:]
		jsonBytes, _ := json.Marshal(target) // TODO: handle error

		scrapersToAuthorize = addTargetToScrapers(&target, jsonBytes)
	}
	// Return addresses of scrapers for authorization purposes
	jsonScrapers, err := json.Marshal(scrapersToAuthorize)
	if err != nil {
		log.Println("Failed marshaling json:", err)
		w.WriteHeader(500)
		return
	}
	w.Write(jsonScrapers)
}

// TODO: test
func removeEndpoint(w http.ResponseWriter, r *http.Request) {
	//// Parse request body
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading request body:", err)
		w.WriteHeader(400)
		return
	}
	var end types.Endpoint
	if err := json.Unmarshal(data, &end); err != nil {
		log.Println("Error unmarshalling json:", err)
		w.WriteHeader(400)
		return
	}
	err = RemoveEndpoint(&end)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	// Get all Enpoint's targets
	resp, err := httpsClient.Get("https://" + end.IP + ":" + end.ManagePort + "/mappings")
	if err != nil {
		w.WriteHeader(500)
		return
	}
	data, err = ioutil.ReadAll(resp.Body)
	err = resp.Body.Close()
	if err != nil {
		log.Printf("removeEndpoint: failed to close response's body getting mappings: %v", err)
	}
	var mappings map[string]string
	err = json.Unmarshal(data, &mappings)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	// Remove endpoint targets from each scrapers like in notify removed mapping
	for _, scr := range getScrapers() {
		for path := range mappings {
			target := types.Target{}
			target.AS = local.IA.A.String()
			target.ISD = fmt.Sprint(local.IA.I)
			target.IP = end.IP
			target.Path = path
			target.Port = fmt.Sprint(local.L4Port)
			target.Labels = make(map[string]string)
			target.Name = target.Path[1:]
			targetJson, err := json.Marshal(target)
			req, err := http.NewRequest("DELETE", "https://"+scr.IP+":"+scr.ManagePort+"/targets", bytes.NewReader(targetJson))
			if err != nil {
				log.Println("Error in creating DELETE target request")
				w.WriteHeader(500)
				return
			}
			resp, err = httpsClient.Do(req)
			if err != nil {
				log.Println("Error in removing scraper target:", err)
			}
			err = resp.Body.Close()
			if err != nil {
				log.Printf("removeEndpoint: failed to close response's body deleting endpont %v. Error is: %v", target, err)
			}
		}
	}
	w.WriteHeader(204)
}

// Registers a new scraper by writing it in the scrapers' list.
func registerScraper(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading request body:", err)
		w.WriteHeader(400)
		return
	}
	var scr types.Scraper
	if err := json.Unmarshal(data, &scr); err != nil {
		log.Println("Error unmarshalling json:", err)
		w.WriteHeader(400)
		return
	}
	err = addScraper(&scr)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(204)
}

func removeScraper(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading request body:", err)
		w.WriteHeader(400)
		return
	}
	var scr types.Scraper
	if err := json.Unmarshal(data, &scr); err != nil {
		log.Println("Error unmarshalling json:", err)
		w.WriteHeader(400)
		return
	}
	err = RemoveScraper(&scr)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	// Get scraper targets
	resp, err := httpsClient.Get("https://" + scr.IP + ":" + scr.ManagePort + "/targets")
	if err != nil {
		log.Printf("removeScraper: error getting targets: %v", err)
		w.WriteHeader(500)
		return
	}
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("removeScraper: error reading body: %v", err)
		w.WriteHeader(500)
		return
	}
	err = resp.Body.Close()
	if err != nil {
		log.Printf("removeScraper: error closing request's body: %v", err)
		w.WriteHeader(500)
		return
	}
	var targets []types.Target
	json.Unmarshal(data, &targets)
	// Compute list of endpoints (since mapping->endpoint is many-to-one)
	endpoints := make(map[string]string)
	for _, target := range targets {
		endpoints[target.IP] = getEndpointByIP(target.IP).ManagePort
	}
	// Remove all permissions for the removed scraper on each endpoint
	w.WriteHeader(204)
	for ip, port := range endpoints {
		req, _ := http.NewRequest("DELETE", "https://"+ip+":"+port+"/"+scr.IA+":"+scr.IP+"/permissions", nil)
		resp, err = httpsClient.Do(req)
		if err != nil {
			log.Printf("removeScraper: error deleting scraper on %v: %v", req.URL, err)
			w.WriteHeader(500)
			continue
		}
		err = resp.Body.Close()
		if resp != nil {
			log.Printf("removeScraper: could not close response's body after deleting scraper on %v. Error is: %v", req.URL, err)
		}
	}
}

// Returns the list of all registered scrapers.
func listStorages(w http.ResponseWriter, r *http.Request) {
	log.Println("List storages received")
	jsonScrs, err := ioutil.ReadFile("storages.json")
	if err != nil {
		log.Println("Error while reading file:", err)
		w.WriteHeader(400)
		return
	}
	w.Write(jsonScrs)
}

func registerStorage(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading request body:", err)
		w.WriteHeader(400)
		return
	}
	var str types.Storage
	if err := json.Unmarshal(data, &str); err != nil {
		log.Println("Error unmarshalling json:", err)
		w.WriteHeader(400)
		return
	}
	err = addStorage(&str)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	w.WriteHeader(204)
}

func removeStorage(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading request body:", err)
		w.WriteHeader(400)
		return
	}
	var str types.Storage
	if err := json.Unmarshal(data, &str); err != nil {
		log.Println("Error unmarshalling json:", err)
		w.WriteHeader(400)
		return
	}
	err = RemoveStorage(&str)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	w.WriteHeader(204)
}

func blockSigning(w http.ResponseWriter, r *http.Request) {
	refuseSigning = true
	w.WriteHeader(204)
}

func enableSigning(w http.ResponseWriter, r *http.Request) {
	refuseSigning = false
	go func() {
		time.Sleep(1 * time.Hour)
		refuseSigning = true
	}()
	w.Write([]byte("Signing enabled for 1h"))
}

func addScraperTarget(w http.ResponseWriter, r *http.Request) {
	data, _ := ioutil.ReadAll(r.Body)
	r.Body = ioutil.NopCloser(bytes.NewReader(data))
	redirect(w, r)
	// Get scraper's SCION address from file using path's IP
	host := strings.Split(r.URL.Path, "/")[2]
	ip := strings.Split(host, ":")[0]
	scraper := getScraperByIP(ip)
	scraperAddr := scraper.IA + ":" + scraper.IP
	// Parse request to get target
	var target types.Target
	err := json.Unmarshal(data, &target)
	if err != nil {
		log.Println("Failed unmarshalling data:", err)
		w.WriteHeader(500)
		return
	}
	// Add authorization for scraper to target
	jsonRole, err := json.Marshal(target.Path[1:] + "_" + common.OwnerRole)
	if err != nil {
		log.Println("Failed marshalling data:", err)
		w.WriteHeader(500)
		return
	}
	endpoint := getEndpointByIP(target.IP)
	// Add scraper to mapping owner role
	resp, err := httpsClient.Post("https://"+target.IP+":"+endpoint.ManagePort+"/"+scraperAddr+"/roles", "application/json", bytes.NewReader(jsonRole))
	if err != nil || resp.Status != "201" {
		log.Println("Failed adding scraper to owner role:", resp.Status, err)
	}
	err = resp.Body.Close()
	if resp != nil {
		log.Printf("addScraperTarget: could not close response's body after adding scraper to owner role. Error is: %v", err)
	}
	// Enable scraping
	resp, err = httpsClient.Get("https://" + target.IP + ":" + endpoint.ManagePort + "/" + scraperAddr + target.Path + "/enable")
	if err != nil || resp.Status != "201" {
		log.Println("Failed enabling scraping:", resp.Status, err)
	}
	err = resp.Body.Close()
	if resp != nil {
		log.Printf("addScraperTarget: could not close response's body after enabling scraper for scraping. Error is: %v", err)
	}
}

func removeScraperTarget(w http.ResponseWriter, r *http.Request) {
	data, _ := ioutil.ReadAll(r.Body)
	r.Body = ioutil.NopCloser(bytes.NewReader(data))
	redirect(w, r)
	// Get scraper's SCION address from file using path's IP
	host := strings.Split(r.URL.Path, "/")[2]
	ip := strings.Split(host, ":")[0]
	scraper := getScraperByIP(ip)
	scraperAddr := scraper.IA + ":" + scraper.IP
	// Parse request to get target
	var target types.Target
	err := json.Unmarshal(data, &target)
	if err != nil {
		log.Println("Failed unmarshalling data:", err)
		w.WriteHeader(500)
		return
	}
	// Remove mapping owner role for scraper
	jsonRole, err := json.Marshal(target.Path[1:] + "_" + common.OwnerRole)
	if err != nil {
		log.Println("Failed marshalling data:", err)
		w.WriteHeader(500)
		return
	}
	endpoint := getEndpointByIP(target.IP)
	req, err := http.NewRequest("DELETE", "https://"+target.IP+":"+endpoint.ManagePort+"/"+scraperAddr+"/roles", bytes.NewReader(jsonRole))
	resp, err := httpsClient.Do(req)
	if err != nil || resp.Status != "204" {
		log.Println("Failed removing owner role from scraper:", resp.Status, err)
	}
	err = resp.Body.Close()
	if resp != nil {
		log.Printf("removeScraperTarget: could not close response's body after removing scraper's owner role. Error is: %v", err)
	}
	// Block scraping
	resp, err = httpsClient.Get("https://" + target.IP + ":" + endpoint.ManagePort + "/" + scraperAddr + target.Path + "/block")
	if err != nil || resp.Status != "204" {
		log.Println("Failed blocking scraping:", resp.Status, err)
	}
	err = resp.Body.Close()
	if resp != nil {
		log.Printf("removeScraperTarget: could not close response's body after blocking scraper for scraping. Error is: %v", err)
	}

}

func syncScraperTargets(w http.ResponseWriter, r *http.Request) {
	// Get scraper by ip address in path
	scraperIP := mux.Vars(r)["addr"]
	scraper := getScraperByIP(scraperIP)

	// Get all registered endpoints
	endpoints := getEndpoints()

	// Try adding targets for each endpoint to the scraper
	for _, end := range endpoints {
		targetISD := strings.Split(end.IA, "-")[0]
		targetAS := strings.Split(end.IA, "-")[1]
		labels := make(map[string]string)
		labels["ISD"] = targetISD
		labels["AS"] = targetAS
		if scraper.Covers(targetISD) {
			// Build generic target for all target on `end`
			target := types.Target{
				"",
				targetISD,
				targetAS,
				end.IP,
				end.ScrapePort,
				"",
				labels,
			}
			for _, path := range end.Paths {
				// Add specific path and name infos
				target.Name = path[1:]
				target.Path = path
				labels["service"] = target.Name
				// Encode target to json
				jsonTarget, _ := json.Marshal(target) // TODO: handle error
				// Add target to scraper
				addTargetToScraper(jsonTarget, scraper)
				// Assign owner role for the mapping to the scraper
				jsonRole, _ := json.Marshal(target.Name + "_owner")
				source := scraper.IA + ":" + scraper.IP
				endpointAddress := end.IP + ":" + end.ManagePort
				resp, err := httpsClient.Post("https://"+endpointAddress+"/"+source+"/roles", "application/json", bytes.NewReader(jsonRole))
				if err != nil {
					log.Printf("Post request for adding role failed with error: %v", err)
				} else if resp.StatusCode != http.StatusNoContent {
					log.Printf("Add owner role for %s to scraper %s at endpoint %s failed. Response was: %v", target.Path, source, end.IP, resp)
				} else {
					log.Printf("Added owner role for %s to scraper %s at endpoint %s.", target.Path, source, end.IP)
				}
				err = resp.Body.Close()
				if resp != nil {
					log.Printf("syncScraperTargets: could not close response's body after adding owner role to scraper. Error is: %v", err)
				}
				// Assign a scrape permission for the mapping to the scraper
				resp, err = httpsClient.Get("https://" + endpointAddress + "/" + source + target.Path + "/enable")
				if err != nil {
					log.Printf("Post request for adding scrape permission failed with error: %v", err)
				} else if resp.StatusCode != http.StatusNoContent {
					log.Printf("Add scrape permission for %s to scraper %s at endpoint %s failed. Response was: %v", target.Path, source, end.IP, resp)
				} else {
					log.Printf("Added scrape permission for %s to scraper %s at endpoint %s.", target.Path, source, end.IP)
				}
				err = resp.Body.Close()
				if resp != nil {
					log.Printf("syncScraperTargets: could not close response's body after enabling scraper for scraping. Error is: %v", err)
				}
			}
		}
	}
}

func redirect(w http.ResponseWriter, r *http.Request) {
	// Redirection call path are defined to have /component/address as prefix
	redirAddr := "https://" + strings.SplitN(r.URL.Path, "/", 3)[2]
	var resp *http.Response
	defer func() {
		if resp == nil {
			log.Printf("DEBUG redirect not redirecting to %s, resp is nul", redirAddr)
			return
		}
		log.Printf("DEBUG redirect redirecting to %s", redirAddr)
		err := resp.Body.Close()
		if err != nil {
			log.Printf("redirect to %s: could not close response's body: %v", redirAddr, err)
		}
	}()
	var err error
	switch r.Method {
	case "GET":
		resp, err = httpsClient.Get(redirAddr)
	case "POST":
		resp, err = httpsClient.Post(redirAddr, r.Header.Get("Content-Type"), r.Body)
	case "DELETE":
		req, err := http.NewRequest("DELETE", redirAddr, r.Body)
		if err != nil {
			log.Println("Error in creating new target addition request:", err)
		}
		resp, err = httpsClient.Do(req)
	}
	if err != nil {
		log.Println("Error in getting scraper targets:", err)
		w.WriteHeader(500)
		return
	}
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("redirect to %s:error copying body: %v", redirAddr, err)
	}

}

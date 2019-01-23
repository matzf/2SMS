package main

import (
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/netsec-ethz/2SMS/common"
	"github.com/netsec-ethz/2SMS/common/types"
	"github.com/scionproto/scion/go/lib/snet"
)

var (
	managerCert       string
	managerPrivKey    string
	managerIP         string
	managerDNS        string
	noClientVerifPort string
	clientVerifPort   string
	managementPort    string
	approvedCertsDir  string
	waitingCSRDir     string
	caKey             string = "ca/ca.key"
	caCertFile        string = "ca/ca.crt"
	caSerialFile      string = "ca/serial"
	caDir             string = "ca"
	authDir           string = "auth"
	ca                *common.CA
	local             snet.Addr
	refuseSigning     = true
	httpsClient       *http.Client
)

func initManager() {
	flag.StringVar(&managerCert, "manager.cert", "auth/manager.crt", "full chain manager's certificate file")
	flag.StringVar(&managerPrivKey, "manager.key", "auth/manager.key", "manager's private key file")
	flag.StringVar(&managerIP, "manager.IP", "127.0.0.1", "IP address of the manager machine")
	flag.StringVar(&managerDNS, "manager.DNS", "localhost", "DNS name of the manager machine")
	flag.StringVar(&noClientVerifPort, "ports.no-client-verif", "10000", "port where the client API is exposed (no client side verification)")
	flag.StringVar(&clientVerifPort, "ports.client-verif", "10001", "port where the client API is exposed (with client side verification)")
	flag.StringVar(&managementPort, "ports.management", "10002", "port where the management api is exposed")
	flag.StringVar(&approvedCertsDir, "manager.approved-certs", "approved_certs", "directory where approved certificate are stored")
	flag.StringVar(&waitingCSRDir, "manager.waiting-csrs", "waiting_csrs", "directory where still non approved csr are stored")
	flag.Var((*snet.Addr)(&local), "local", "(Mandatory) local SCION information (port is not needed)")

	flag.Parse()

	var err error
	// Create directory to store auth data
	if !common.FileExists(authDir) {
		os.Mkdir(authDir, 0700) // The private key is stored here, so permissions are restrictive
	}
	if !common.FileExists(caDir) {
		os.Mkdir(caDir, 0700)
	}
	if !common.FileExists(approvedCertsDir) {
		os.Mkdir(approvedCertsDir, 0700)
	}
	if !common.FileExists(waitingCSRDir) {
		os.Mkdir(waitingCSRDir, 0700)
	}
	if !common.FileExists(caKey) {
		name := &pkix.Name{
			Organization:       []string{"SCIONLab"},
			OrganizationalUnit: []string{"CA"},
			Country:            []string{"CH"},
			Province:           []string{"Zurich"},
			Locality:           []string{"Zurich"},
		}
		duration := &common.Duration{1, 0, 0}
		ca, err = common.NewCA(name, duration, caKey, caSerialFile, caCertFile, local.IA)
		if err != nil {
			log.Fatal("Failed creating a new CA:", err)
		}
	} else {
		ca, err = common.LoadCA(caKey, caSerialFile, caCertFile)
		if err != nil {
			log.Fatal("Failed loading CA:", err)
		}
	}
	var privKey *ecdsa.PrivateKey
	if !common.FileExists(managerPrivKey) {
		// Generate a new key and write it to file
		privKey, _ = common.GenECDSAKey("P256")
		bytes, _ := x509.MarshalECPrivateKey(privKey)
		common.WriteToPEMFile(managerPrivKey, "ECDSA PRIVATE KEY", bytes)
	}
	if !common.FileExists(managerCert) {
		if privKey == nil {
			privKey, err = common.ReadECPrivKeyFromPEMFile(managerPrivKey)
		}
		name := pkix.Name{
			Organization:       []string{"SCIONLab"},
			OrganizationalUnit: []string{"Manager"},
			Country:            []string{"CH"},
			Province:           []string{"Zurich"},
			Locality:           []string{"Zurich"},
		}
		duration := &common.Duration{1, 0, 0}
		certBytes, err := ca.GenCert(name, privKey, duration, []string{managerDNS}, []net.IP{net.ParseIP(managerIP)})
		if err != nil {
			log.Fatal("Error while generating manager certificate:", err)
		}
		common.WriteToPEMFile(managerCert, "CERTIFICATE", certBytes)
	}
	if !common.FileExists("endpoints.json") {
		bts, _ := json.Marshal([]types.Endpoint{})
		err = ioutil.WriteFile("endpoints.json", bts, 0644)
		if err != nil {
			log.Fatal("Failed initializing endpoints file:", err)
		}
	}
	if !common.FileExists("scrapers.json") {
		bts, _ := json.Marshal([]types.Scraper{})
		err = ioutil.WriteFile("scrapers.json", bts, 0644)
		if err != nil {
			log.Fatal("Failed initializing scrapers file:", err)
		}
	}
	if !common.FileExists("storages.json") {
		bts, _ := json.Marshal([]types.Scraper{})
		err = ioutil.WriteFile("storages.json", bts, 0644)
		if err != nil {
			log.Fatal("Failed initializing storages file:", err)
		}
	}

	// Bootstrap PKI
	err = common.Bootstrap(caCertFile, "ca/bootstrap.json") // TODO: use param for caData
	if err != nil {
		log.Fatal("Verification of ca certificate failed:", err)
	} else {
		log.Println("Successfully verified ca certificate.")
	}

	httpsClient = common.CreateHttpsClient(caDir, managerCert, managerPrivKey)
}

func main() {
	initManager()
	log.Println("Started Manager Application")

	// HTTPS Server for PKI operations without client side verification
	go func() {
		router := mux.NewRouter()
		// Send csr and ask to sign it
		router.HandleFunc("/certificate/request", requestCert).Methods("POST")
		// Retrieve certificate
		router.HandleFunc("/certificates/{type}/{ip}/get", getCert).Methods("GET")

		srv := common.CreateHttpsServer(caDir, managerCert, managerPrivKey, "", noClientVerifPort, router, tls.NoClientCert)
		log.Println("Starting server without client verification")
		log.Fatal("Server without client verification listening error:", srv.ListenAndServeTLS(managerCert, managerPrivKey))
	}()

	// HTTPS Server for operations with client side verification
	go func() {
		router := mux.NewRouter()

		router.HandleFunc("/endpoint/mappings/notify", notifyAddedMapping).Methods("POST")
		router.HandleFunc("/endpoint/mappings/notify", notifyRemovedMapping).Methods("DELETE")
		router.HandleFunc("/endpoints/register", registerEndpoint).Methods("POST")

		router.HandleFunc("/scrapers/register", registerScraper).Methods("POST")

		router.HandleFunc("/storages/register", registerStorage).Methods("POST")

		srv := common.CreateHttpsServer(caDir, managerCert, managerPrivKey, "", clientVerifPort, router, tls.RequireAndVerifyClientCert)
		log.Println("Starting server with client verification")
		log.Fatal("Server with client verification listening error:", srv.ListenAndServeTLS(managerCert, managerPrivKey))
	}()

	// HTTP Management Server (localhost only)
	router := mux.NewRouter()
	// TODO: currently a csr is either signed or refused, but never put in waiting list.
	//router.HandleFunc("/manager/certificate/requests", getRequests).Methods("GET")
	//router.HandleFunc("/manager/certificate/approve", approveRequest).Methods("POST")
	router.HandleFunc("/manager/signing/block", blockSigning).Methods("GET")
	router.HandleFunc("/manager/signing/enable", enableSigning).Methods("GET")
	router.HandleFunc("/manager/endpoints", listEndpoints).Methods("GET")
	router.HandleFunc("/manager/scrapers", listScrapers).Methods("GET")
	router.HandleFunc("/manager/storages", listStorages).Methods("GET")
	router.HandleFunc("/manager/scrapers/remove", removeScraper).Methods("DELETE")
	router.HandleFunc("/manager/endpoints/remove", removeEndpoint).Methods("DELETE")
	router.HandleFunc("/manager/storages/remove", removeStorage).Methods("DELETE")

	router.HandleFunc("/endpoint/{addr}/mappings", redirect).Methods("GET")
	router.HandleFunc("/endpoint/{addr}/mappings", redirect).Methods("POST")
	router.HandleFunc("/endpoint/{addr}/mappings", redirect).Methods("DELETE")
	router.HandleFunc("/endpoint/{addr}/{mapping}/metrics/list", redirect).Methods("GET")
	router.HandleFunc("/endpoint/{addr}/access_control", redirect).Methods("POST")
	router.HandleFunc("/endpoint/{addr}/access_control", redirect).Methods("DELETE")
	router.HandleFunc("/endpoint/{addr}/sources", redirect).Methods("GET")
	router.HandleFunc("/endpoint/{addr}/{source}/roles", redirect).Methods("GET")
	router.HandleFunc("/endpoint/{addr}/{source}/roles", redirect).Methods("POST")
	router.HandleFunc("/endpoint/{addr}/{source}/roles", redirect).Methods("DELETE")
	router.HandleFunc("/endpoint/{addr}/{source}/permissions", redirect).Methods("DELETE")
	router.HandleFunc("/endpoint/{addr}/{source}/permissions", redirect).Methods("GET")
	router.HandleFunc("/endpoint/{addr}/{source}/status", redirect).Methods("GET")
	router.HandleFunc("/endpoint/{addr}/{source}/{mapping}/block", redirect).Methods("GET")
	router.HandleFunc("/endpoint/{addr}/{source}/{mapping}/enable", redirect).Methods("GET")
	router.HandleFunc("/endpoint/{addr}/{source}/{mapping}/frequency", redirect).Methods("DELETE")
	router.HandleFunc("/endpoint/{addr}/{source}/{mapping}/frequency", redirect).Methods("POST")
	router.HandleFunc("/endpoint/{addr}/{source}/{mapping}/window", redirect).Methods("DELETE")
	router.HandleFunc("/endpoint/{addr}/{source}/{mapping}/window", redirect).Methods("POST")
	router.HandleFunc("/endpoint/{addr}/roles", redirect).Methods("GET")
	router.HandleFunc("/endpoint/{addr}/roles", redirect).Methods("POST")
	router.HandleFunc("/endpoint/{addr}/roles/{role}", redirect).Methods("DELETE")
	router.HandleFunc("/endpoint/{addr}/roles/{role}", redirect).Methods("GET")
	router.HandleFunc("/endpoint/{addr}/roles/{role}/permissions/{mapping}", redirect).Methods("POST")
	router.HandleFunc("/endpoint/{addr}/roles/{role}/permissions/{mapping}", redirect).Methods("DELETE")

	router.HandleFunc("/scraper/{addr}/targets", redirect).Methods("GET")
	router.HandleFunc("/scraper/{addr}/targets", addScraperTarget).Methods("POST")
	router.HandleFunc("/scraper/{addr}/targets", removeScraperTarget).Methods("DELETE")
	router.HandleFunc("/scraper/{addr}/targets/sync", syncScraperTargets).Methods("GET")
	router.HandleFunc("/scraper/{addr}/storages", redirect).Methods("GET")
	router.HandleFunc("/scraper/{addr}/storages", redirect).Methods("POST")
	router.HandleFunc("/scraper/{addr}/storages", redirect).Methods("DELETE")

	//router.HandleFunc("/authorization/requests", listPermissionRequests).Methods("GET")
	//router.HandleFunc("/authorization/approve", approvePermissionRequest).Methods("POST")

	srv := &http.Server{
		Addr:    "127.0.0.1:" + managementPort,
		Handler: router,
	}
	log.Fatal("Localhost HTTP server listening error:", srv.ListenAndServe())
}

package main

import (
	"crypto/tls"
	"encoding/gob"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"github.com/netsec-ethz/2SMS/common"

	"github.com/gorilla/mux"

	sd "github.com/scionproto/scion/go/lib/sciond"
	"github.com/scionproto/scion/go/lib/snet"

	"github.com/lucas-clemente/quic-go"
	//"github.com/scionproto/scion/go/lib/snet/squic" // TODO: change to this (gives type problems)
	"github.com/juagargi/temp_squic"

	"bytes"
	"crypto/ecdsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"io/ioutil"
	"net"
	"strings"
	"sync"

	"github.com/netsec-ethz/2SMS/common/types"
	"github.com/prometheus/common/expfmt"
)

var (
	authorizationPort   string
	nodeOutFile         string
	reloadMappingsMutex = &sync.Mutex{}
	internalMapping     types.EndpointMappings
	endpointIP          string
	endpointDNS         string
	caCertsDir          string
	externalPort        string
	endpointCert        string
	endpointPrivKey     string
	endpointCSR         string
	nodeExporterEnabled string
	managementAPIPort   string
	managerIP           string
	managerVerifPort    string
	managerUnverifPort  string
	local               snet.Addr
	sciond              = flag.String("sciond", "", "Path to sciond socket")
	dispatcher          = flag.String("dispatcher", "/run/shm/dispatcher/default.sock",
		"Path to dispatcher socket")
	authDir                 string = "auth"
	nodeExec                string
	nodePath                string
	nodeListenAddress       string
	localhostManagementPort string
	genFolder               string
	doAccessControl         bool
	accessController        *common.AccessController
	httpsClient             *http.Client
	initRolesFile           string
	authPolicyFile          string
	authModelFile           string
)

func initialize_endpoint() {
	flag.StringVar(&nodeExec, "node.exec", "node-exporter/node_exporter", "path to node exporter executable")
	flag.StringVar(&nodeListenAddress, "node.liste-address", "127.0.0.1:9100", "address where node exporter listens")
	flag.StringVar(&nodePath, "node.path", "/metrics", "path where node exporter's metrics are showed")
	flag.StringVar(&nodeOutFile, "node.out", "node-exporter/out", "file where node exporter output is redirected")
	flag.StringVar(&endpointDNS, "endpoint.DNS", "localhost", "DNS name of endpoint machine")
	flag.StringVar(&endpointIP, "endpoint.IP", "127.0.0.1", "IP of endpoint machine")
	flag.StringVar(&externalPort, "endpoint.external.port", "9200", "externally exposed port for scraping")
	flag.StringVar(&authorizationPort, "endpoint.authorization.port", "9500", "externally exposed port for access control calls")
	flag.StringVar(&endpointCert, "endpoint.cert", "auth/endpoint.crt", "full chain endpoint's certificate file")
	flag.StringVar(&endpointCSR, "endpoint.csr", "auth/endpoint.csr", "csr for the key")
	flag.StringVar(&endpointPrivKey, "endpoint.key", "auth/endpoint.key", "endpoint's private key file")
	flag.StringVar(&nodeExporterEnabled, "endpoint.enable-node", "false", "set to true to enable node_exporter and false otherwise")
	flag.StringVar(&managementAPIPort, "endpoint.ports.management", "9900", "port where the management API is exposed")
	flag.StringVar(&localhostManagementPort, "endpoint.ports.local", "9999", "port where the local management API is exposed")

	flag.StringVar(&initRolesFile, "endpoint.roles_file", "init_roles.json", "contains role definitions that are loaded at startup and added to the authorization policy")
	flag.StringVar(&authModelFile, "endpoint.model", "auth/model.conf", "location of the model file defining authorization schema model")
	flag.StringVar(&authPolicyFile, "endpoint.policy", "auth/policy.csv", "location of the policy file defining authorization schema policy")

	flag.StringVar(&caCertsDir, "ca.certs", "ca_certs", "directory with trusted ca certificates")

	flag.StringVar(&managerIP, "manager.IP", "", "ip address of the manager")
	flag.StringVar(&managerUnverifPort, "manager.unverif-port", "10000", "port where manager listens for certificate request")
	flag.StringVar(&managerVerifPort, "manager.verif-port", "10001", "port where manager listens for authenticated operations")

	flag.StringVar(&genFolder, "gen", "", "path to the SCION gen folder")
	flag.Var((*snet.Addr)(&local), "local", "(Mandatory) address to listen on")

	flag.BoolVar(&doAccessControl, "", true, "")
	flag.Parse()

	gob.Register(types.Request{})
	gob.Register(types.Response{})

	if *sciond == "" {
		*sciond = sd.GetDefaultSCIONDPath(nil)
	}
	// Create directory to store auth data
	if !common.FileExists(authDir) {
		os.Mkdir(authDir, 0700) // The private key is stored here, so permissions are restrictive
	}
	if !common.FileExists(caCertsDir) {
		os.Mkdir(caCertsDir, 0700)
	}

	// Initialize scion network
	common.InitNetwork(local, sciond, dispatcher)

	// Bootstrap PKI
	err := common.Bootstrap(caCertsDir+"/ca.crt", caCertsDir+"/bootstrap.json", local.IA)
	if err != nil {
		log.Fatal("Verification of ca certificate failed:", err)
	} else {
		log.Println("Successfully verified ca certificate.")
	}

	var privKey *ecdsa.PrivateKey
	if !common.FileExists(endpointPrivKey) {
		// Generate a new key and write it to file
		privKey, _ = common.GenECDSAKey("P256")
		bts, _ := x509.MarshalECPrivateKey(privKey)
		common.WriteToPEMFile(endpointPrivKey, "ECDSA PRIVATE KEY", bts)
		name := pkix.Name{
			Organization:       []string{"SCIONLab"},
			OrganizationalUnit: []string{"Endpoint"},
			Country:            []string{"CH"},
			Province:           []string{"Zurich"},
			Locality:           []string{"Zurich"},
		}
		bts, _ = common.GenCertSignRequest(name, privKey, []string{endpointDNS}, []net.IP{net.ParseIP(endpointIP)})
		common.WriteToPEMFile(endpointCSR, "CERTIFICATE REQUEST", bts)
	}
	// Request certificate to the manager
	if !common.FileExists(endpointCert) {
		if managerIP != "" {
			common.RequestAndObtainCert(caCertsDir, managerIP, managerUnverifPort, endpointCert, endpointCSR, "Endpoint", endpointIP)
		} else {
			log.Fatal("No certificate found and no connection with manager. Please manually generate and upload a certificate for the csr.")
		}
	}
	// Init mappings
	if !common.FileExists("mappings.json") {
		var nodeListenPort string
		if enabled, err := strconv.ParseBool(nodeExporterEnabled); err == nil && enabled {
			nodeListenPort = strings.Split(nodeListenAddress, ":")[1]
		}
		mappings := InitInternalMappings(nodeListenPort, genFolder)
		SaveMappings(mappings)
	}

	// Load mapping from file
	internalMapping, err = LoadMappings()
	if err != nil {
		log.Fatal("Failed initializing internal mappings:", err)
	}

	httpsClient = common.CreateHttpsClient(caCertsDir, endpointCert, endpointPrivKey)

	// Initialize Access Controller
	if !common.FileExists(authModelFile) {
		log.Fatal("Casbin authorization model file (" + authModelFile + ") doesn't exist.")
	}
	if !common.FileExists(authPolicyFile) {
		file, _ := os.Create(authPolicyFile)
		file.Close()
	}
	accessController = common.NewAccessController(authModelFile, authPolicyFile, doAccessControl, &local.IA)
}

func main() {
	initialize_endpoint()
	log.Println("Started Endpoint Application")

	// If enabled, run node exporter
	nodeEnabled, err := strconv.ParseBool(nodeExporterEnabled)
	if err != nil {
		log.Fatal("Error in parsing 'activated':", err)
	}
	if nodeEnabled {
		log.Println("Node exporter activated:", nodeEnabled)
		var proc *os.Process
		nodeIsRunning := make(chan struct{})
		go func() {
			cmd := exec.Command("bash", "-c", nodeExec+" --web.telemetry-path="+nodePath+" --web.listen-address="+nodeListenAddress+" &> "+nodeOutFile)
			err := cmd.Start()
			if err != nil {
				log.Fatal("Failed starting Node Exporter: ", err)
			}
			// TODO: check that process really started
			proc = cmd.Process
			log.Println("Started Node Exporter as process ", cmd.Process.Pid)
			// time.Sleep(1 * time.Second)
			nodeIsRunning <- struct{}{}
		}()
		defer proc.Kill()
		<-nodeIsRunning
	}
	// Initialize permissions for mappings
	// Load permissions from file for user defined and reserved roles (core and neighbor)
	if initRolesFile != "" {
		err = accessController.LoadPermsFromFile(initRolesFile)
		if err != nil {
			log.Printf("Loading the role file '%s', ignoring error: %v", initRolesFile, err)
		}
	}
	SyncPermissions(internalMapping, types.EndpointMappings{})
	// Register at manager
	err = SyncManager(internalMapping, types.EndpointMappings{})
	if err != nil {
		log.Fatalf("Initial synchronization to manager failed: %v", err)
	}

	// HTTPS server
	go func() {
		log.Println("Starting HTTPS server")

		srv := common.CreateHttpsServer(caCertsDir, endpointIP, externalPort, &handler{http.Client{}}, tls.RequireAndVerifyClientCert)

		log.Fatal("HTTPS server listening error: ", srv.ListenAndServeTLS(endpointCert, endpointPrivKey))
	}()

	// SCION server
	go func() {
		log.Println("Starting SCION server")
		squic.Init(endpointPrivKey, endpointCert)

		// Initialize HTTP client
		tr := &http.Transport{
			DisableCompression: true,
		}
		client := &http.Client{
			Transport: tr,
		}

		// Listen on SCION address
		qsock, err := squic.ListenSCION(nil, &local)
		if err != nil {
			log.Fatal("Unable to listen: ", err)
		}
		log.Println("Listening on: ", qsock.Addr())
		for {
			qsess, err := qsock.Accept()
			if err != nil {
				// Accept failing means the socket is unusable.
				log.Fatal("Unable to accept quic session: ", err)
			}
			log.Println("Quic session accepted from: ", qsess.RemoteAddr())
			go handleQUICSession(qsess, *client)
		}
	}()

	go func() {
		router := mux.NewRouter()

		router.HandleFunc("/{mapping}/metrics/list", listMetrics).Methods("GET")
		//router.HandleFunc("/metrics/authorization", requestPermissions).Methods("POST") // TODO

		srv := common.CreateHttpsServer(caCertsDir, endpointIP, authorizationPort, router, tls.NoClientCert)
		log.Println("Starting server without client verification")
		log.Fatal("Server without client verification listening error:", srv.ListenAndServeTLS(endpointCert, endpointPrivKey))
	}()

	// Management Server
	router := mux.NewRouter()
	router.HandleFunc("/mappings", listMappings).Methods("GET")
	router.HandleFunc("/mappings", addMapping).Methods("POST")
	router.HandleFunc("/mappings", removeMapping).Methods("DELETE")
	router.HandleFunc("/mappings", putMappings).Methods("PUT")

	router.HandleFunc("/{mapping}/metrics/list", listMetrics).Methods("GET")

	router.HandleFunc("/access_control", enableAccessControl).Methods("POST")
	router.HandleFunc("/access_control", disableAccessControl).Methods("DELETE")
	router.HandleFunc("/sources", listSources).Methods("GET")
	router.HandleFunc("/{source}/roles", listSourceRoles).Methods("GET")
	router.HandleFunc("/{source}/roles", addSourceRole).Methods("POST")
	router.HandleFunc("/{source}/roles", removeSourceRole).Methods("DELETE")
	router.HandleFunc("/{source}/permissions", removeAllSourcePermissions).Methods("DELETE")
	router.HandleFunc("/{source}/permissions", listAllSourcePermissions).Methods("GET")
	router.HandleFunc("/{source}/status", sourceStatus).Methods("GET")
	router.HandleFunc("/{source}/{mapping}/block", blockSource).Methods("GET")
	router.HandleFunc("/{source}/{mapping}/enable", enableSource).Methods("GET")
	router.HandleFunc("/{source}/{mapping}/frequency", removeSourceFrequency).Methods("DELETE")
	router.HandleFunc("/{source}/{mapping}/frequency", setSourceFrequency).Methods("POST")
	router.HandleFunc("/{source}/{mapping}/window", removeSourceWindow).Methods("DELETE")
	router.HandleFunc("/{source}/{mapping}/window", setSourceWindow).Methods("POST")

	router.HandleFunc("/roles", listRoles).Methods("GET")
	router.HandleFunc("/roles", createRole).Methods("POST")
	router.HandleFunc("/roles/{role}", deleteRole).Methods("DELETE")
	router.HandleFunc("/roles/{role}", getRoleInfo).Methods("GET")
	router.HandleFunc("/roles/{role}/permissions/{mapping}", addRolePermissions).Methods("POST")
	router.HandleFunc("/roles/{role}/permissions/{mapping}", removeRolePermissions).Methods("DELETE")

	go func() {
		srv := &http.Server{
			Addr:    "127.0.0.1:" + localhostManagementPort,
			Handler: router,
		}
		log.Println("localhost HTTP server listening error: ", srv.ListenAndServe())
	}()

	srv := common.CreateHttpsServer(caCertsDir, endpointIP, managementAPIPort, router, tls.RequireAndVerifyClientCert)
	log.Println("Starting HTTPS management server")
	log.Fatal("HTTPS server listening error: ", srv.ListenAndServeTLS(endpointCert, endpointPrivKey))
}

func handleQUICSession(qsess quic.Session, client http.Client) {
	source, _ := snet.AddrFromString(qsess.RemoteAddr().String())

	// Authenticate source
	if err := common.DRKeyAuthenticate(source); err != nil {
		log.Println("Failed authenticating source:", err)
		return
	}

	// Accept a stream over the session
	qstream, err := qsess.AcceptStream()
	if err != nil {
		log.Println("Unable to accept quic stream: ", err)
		return
	}

	decoder := gob.NewDecoder(qstream)
	encoder := gob.NewEncoder(qstream)

	// Read and decode request from the stream
	var req types.Request
	err = decoder.Decode(&req)
	if err != nil {
		log.Println("Failed decoding request: ", err)
		return
	}

	path := req.URL.Path
	// Check if remote is authorized to scrape
	if err != nil {
		log.Println("Could not parse remote address to snet.Addr:", err)
		return
	}
	sourceID := source.IA.String() + ":" + source.Host.IP().String()
	if err := accessController.Authorized(sourceID, path); err != nil {
		log.Println("Remote ", source, "not authorized to scrape endpoint")
		return
	}
	// Make HTTP GET request to mapped target on localhost
	resp, err := LocalhostGet(path, client)
	if err != nil {
		log.Println("Failed contacting 127.0.0.1:", err)
		return
	}

	// Read response and parse body to []MetricFamily
	metrics := DecodeResponseBody(resp)
	// Filter metrics according to authorization policy
	filteredMetrics := accessController.FilterMetrics(sourceID, path, metrics)
	// Create new response body with filtered metrics
	byts, err := EncodeMetrics(filteredMetrics, expfmt.Negotiate(req.Header))
	if err != nil {
		log.Println("Error while encoding metrics:", err)
		return
	}
	resp.Body = ioutil.NopCloser(bytes.NewReader(byts))

	// read and encode HTTP response
	quicResp := common.CopyResponseToQUIC(*resp)
	err = encoder.Encode(quicResp)
	if err != nil {
		log.Println("Failed encoding response: ", err)
		return
	}
}

type handler struct {
	client http.Client
}

// Redirects to the right port on localhost based on request path and configured mapping
func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Println("HTTPS request accepted from:", req.RemoteAddr)
	// Get path from request
	path := req.URL.Path
	// Get internal port from mapping
	resp, err := LocalhostGet(path, h.client)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	// Copy back response
	io.Copy(w, resp.Body)
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
}

//// TEST PURPOSE ONLY
//func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
//	log.Println("HTTPS request accepted from:", req.RemoteAddr)
//	// Get path from request
//	path := req.URL.Path
//
//	// TODO: check if path exists before authenticating
//
//	// TEST PURPOSE ONLY
//	source := "17-ffaa:1:43:" + strings.Split(req.RemoteAddr, ":")[0]
//	if err := accessController.Authorized(source, path); err != nil {
//		log.Println(err)
//		w.WriteHeader(401)
//		return
//	}
//	// TEST PURPOSE ONLY
//
//	resp, err := LocalhostGet(path, h.client)
//	if err != nil {
//		w.WriteHeader(500)
//		return
//	}
//
//	// TEST PURPOSE ONLY
//	// Read response and parse body to []MetricFamily
//	metrics := DecodeResponseBody(resp)
//	// Filter metrics according to authorization policy
//	filteredMetrics := accessController.FilterMetrics(source, path, metrics)
//	// Create new response body with filtered metrics
//	byts, err := EncodeMetrics(filteredMetrics, expfmt.Negotiate(req.Header))
//	if err != nil {
//		log.Println("Error while encoding metrics:", err)
//		return
//	}
//	resp.Body = ioutil.NopCloser(bytes.NewReader(byts))
//	// TEST PURPOSE ONLY
//
//	// Copy back response
//	io.Copy(w, resp.Body)
//	for k, vv := range resp.Header {
//		for _, v := range vv {
//			w.Header().Add(k, v)
//		}
//	}
//}

func LocalhostGet(path string, client http.Client) (*http.Response, error) {
	// Make HTTP GET request to mapped target on localhost
	reloadMappingsMutex.Lock()
	internalPort := internalMapping[path]
	reloadMappingsMutex.Unlock()
	resp, err := client.Get("http://127.0.0.1:" + internalPort + "/metrics")
	if err != nil {
		log.Println("Error while contacting local target: ", err)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		log.Println("Status code", resp.StatusCode, "instead of 200 from", "http://127.0.0.1:"+internalPort+path)
		return nil, err
	}
	return resp, nil
}

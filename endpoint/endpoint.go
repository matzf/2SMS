package main

import (
	"crypto/tls"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/netsec-ethz/scion-apps/lib/shttp"

	"github.com/netsec-ethz/2SMS/common"

	"github.com/gorilla/mux"

	sd "github.com/scionproto/scion/go/lib/sciond"
	"github.com/scionproto/scion/go/lib/snet"

	"crypto/ecdsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"net"
	"sync"

	"github.com/netsec-ethz/2SMS/common/types"
)

var (
	nodeOutFile         string
	reloadMappingsMutex = &sync.Mutex{}
	internalMapping     types.EndpointMappings
	endpointIP          string
	endpointPublicBind  string
	endpointDNS         string
	caCertsDir          string
	externalPort        string
	endpointCert        string
	endpointPrivKey     string
	endpointCSR         string
	endpointLocalTarget string
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
	localHTTPClient         *http.Client
	initRolesFile           string
	authPolicyFile          string
	authModelFile           string
)

func initialize_endpoint() {
	flag.StringVar(&nodeExec, "node.exec", "node-exporter/node_exporter", "path to node exporter executable")
	flag.StringVar(&nodeListenAddress, "node.listen-address", "localhost:9100", "address where node exporter listens")
	flag.StringVar(&nodePath, "node.path", "/metrics", "path where node exporter's metrics are showed")
	flag.StringVar(&nodeOutFile, "node.out", "node-exporter/out", "file where node exporter output is redirected")
	flag.StringVar(&endpointDNS, "endpoint.DNS", "localhost", "DNS name of endpoint machine")
	flag.StringVar(&endpointPublicBind, "endpoint.external.bind", "0.0.0.0", "IP that the scrape proxy will bind to")
	flag.StringVar(&externalPort, "endpoint.external.port", "9200", "externally exposed port for scraping")
	flag.StringVar(&endpointCert, "endpoint.cert", "auth/endpoint.crt", "full chain endpoint's certificate file")
	flag.StringVar(&endpointCSR, "endpoint.csr", "auth/endpoint.csr", "csr for the key")
	flag.StringVar(&endpointPrivKey, "endpoint.key", "auth/endpoint.key", "endpoint's private key file")
	flag.StringVar(&nodeExporterEnabled, "endpoint.enable-node", "false", "set to true to enable node_exporter and false otherwise")
	flag.StringVar(&managementAPIPort, "endpoint.ports.management", "9900", "port where the management API is exposed")
	flag.StringVar(&localhostManagementPort, "endpoint.ports.local", "9999", "port where the local management API is exposed")
	flag.StringVar(&endpointLocalTarget, "endpoint.local.target", "localhost", "Internal IP address where SCION services expose their Prometheus metrics")

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

	endpointIP = local.Host.L3.IP().String()

	// Bootstrap PKI
	err := common.Bootstrap(caCertsDir+"/ca.crt", caCertsDir+"/bootstrap.json")
	if err != nil {
		log.Fatal("Verification of ca certificate failed:", err)
	} else {
		log.Println("Successfully verified ca certificate.")
	}

	var privKey *ecdsa.PrivateKey
	if !common.FileExists(endpointPrivKey) {
		log.Printf("No private key found. Generating one in %s", endpointPrivKey)
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
			log.Printf("Certificate not found on %s. Requesting one.", endpointCert)
			common.RequestAndObtainCert(caCertsDir, managerIP, managerUnverifPort, endpointCert, endpointCSR, "Endpoint", endpointIP)
		} else {
			log.Fatal("No certificate found and no connection with manager. Please manually generate and upload a certificate for the csr.")
		}
	}
	// Init mappings
	if !common.FileExists("mappings.json") {
		log.Fatal("Mappings mappings.json file not found in endpoint directory. \nMake sure to create such file with a list of types.Mapping objects in json format.")
	}

	// Load mapping from file
	internalMapping, err = LoadMappings()
	if err != nil {
		log.Fatal("Failed initializing internal mappings:", err)
	}
	httpsClient = common.CreateHttpsClient(caCertsDir, endpointCert, endpointPrivKey)
	localHTTPClient = &http.Client{}
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

		srv := common.CreateHttpsServer(caCertsDir, endpointCert, endpointPrivKey, endpointPublicBind, externalPort, &LocalHandler{"HTTPS", localHTTPClient}, tls.RequireAndVerifyClientCert)

		log.Fatal("HTTPS server listening error: ", srv.ListenAndServeTLS(endpointCert, endpointPrivKey))
	}()

	// SCION server
	go func() {
		log.Printf("Starting SCION server")
		err = shttp.ListenAndServeSCION(strings.Replace(local.String(), " (UDP)", "", 1), endpointCert, endpointPrivKey, &LocalHandler{"SCION HTTPS", localHTTPClient})

		if err != nil {
			log.Printf("SCION HTTP server listening error: %v", err)
		}
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
			Addr:    "localhost:" + localhostManagementPort,
			Handler: router,
		}
		log.Println("localhost HTTP server listening error: ", srv.ListenAndServe())
	}()

	srv := common.CreateHttpsServer(caCertsDir, endpointCert, endpointPrivKey, endpointPublicBind, managementAPIPort, router, tls.RequireAndVerifyClientCert)
	log.Println("Starting HTTPS management server")
	log.Fatal("HTTPS server listening error: ", srv.ListenAndServeTLS(endpointCert, endpointPrivKey))
}

type LocalHandler struct {
	clientType string
	client     *http.Client
}

// Redirects to the right port on localhost based on request path and configured mapping
func (h *LocalHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Printf("Received %s request for path %s", h.clientType, req.URL)
	// Get path from request
	path := req.URL.Path
	// Get internal port from mapping
	resp, err := LocalhostGet(path, h.client)
	if err != nil {
		log.Printf("Failed: %s request from %s to %s%s. Error is: %v", h.clientType, req.RemoteAddr, req.Host, req.URL, err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// Copy back response
	io.Copy(w, resp.Body)
	err = resp.Body.Close()
	if err != nil {
		log.Printf("serveHTTP: could not close response's body after copying it for redirection. Error is: %v", err)
	}
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	log.Printf("Succeeded: %s request from %s to %s%s", h.clientType, req.RemoteAddr, req.Host, req.URL)
}

func LocalhostGet(path string, client *http.Client) (*http.Response, error) {
	// Make HTTP GET request to mapped target on localhost
	reloadMappingsMutex.Lock()
	internalPort := internalMapping[path]
	reloadMappingsMutex.Unlock()
	resp, err := client.Get("http://" + endpointLocalTarget + ":" + internalPort + "/metrics")
	if err != nil {
		log.Println("Error while contacting local target: ", err)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		log.Println("Status code", resp.StatusCode, "instead of 200 from", "http://"+endpointLocalTarget+":"+internalPort+path)
		return nil, err
	}
	return resp, nil
}

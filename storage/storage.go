package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/netsec-ethz/2SMS/common"

	"bytes"
	"crypto/ecdsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/gob"
	"errors"
	"github.com/gorilla/mux"
	"github.com/juagargi/temp_squic"
	"github.com/lucas-clemente/quic-go"
	"github.com/netsec-ethz/2SMS/common/types"
	sd "github.com/scionproto/scion/go/lib/sciond"
	"github.com/scionproto/scion/go/lib/snet"
	"io/ioutil"
	"net"
)

var (
	caCertsDir         string
	internalPort       string
	externalPort       string
	storageCert        string
	storagePrivKey     string
	storageCSR         string
	managementAPIPort  string
	managerPort        string
	storageIP          string
	storageDNS         string
	managerIP          string
	managerVerifPort   string
	managerUnverifPort string
	local              snet.Addr
	sciond             = flag.String("sciond", "", "Path to sciond socket")
	dispatcher         = flag.String("dispatcher", "/run/shm/dispatcher/default.sock",
		"Path to dispatcher socket")
	authDir                 string = "auth"
	localhostManagementPort string
	writePath               string
	readPath                string
	dbName                  string
)

func init() {
	flag.StringVar(&internalPort, "storage.ports.internal_db", "8086", "localhost port where database is listening")
	flag.StringVar(&externalPort, "storage.ports.external_write", "8186", "externally exposed port")
	flag.StringVar(&managementAPIPort, "storage.ports.management", "9900", "port wher the management API is exposed")
	flag.StringVar(&storageCert, "storage.cert", "auth/storage.crt", "full chain storage's certificate file")
	flag.StringVar(&storagePrivKey, "storage.key", "auth/storage.key", "storage's private key file")
	flag.StringVar(&storageCSR, "storage.csr", "auth/storage.csr", "csr for the key")
	flag.StringVar(&localhostManagementPort, "storage.ports.local", "9999", "port where the local management API is exposed")

	flag.StringVar(&storageDNS, "storage.DNS", "localhost", "DNS name of storage machine")
	flag.StringVar(&storageIP, "storage.IP", "127.0.0.1", "IP of storage machine")

	flag.StringVar(&managerPort, "manager.port", "10000", "port where manager listens for certificate request")
	flag.StringVar(&caCertsDir, "ca.certs", "ca_certs", "directory with trusted ca certificates")

	flag.StringVar(&managerIP, "manager.IP", "", "ip address of the manager")
	flag.StringVar(&managerUnverifPort, "manager.unverif-port", "10000", "port where manager listens for certificate request")
	flag.StringVar(&managerVerifPort, "manager.verif-port", "10001", "port where manager listens for authenticated operations")

	flag.StringVar(&writePath, "storage.write", "/api/v1/prom/write", "Path for writing to the database")
	flag.StringVar(&readPath, "storage.read", "/api/v1/prom/read", "Path for reading from the database")
	flag.StringVar(&dbName, "storage.database", "prometheus", "Name of the database")

	flag.Var((*snet.Addr)(&local), "local", "(Mandatory) address to listen on")

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

	// Bootstrap PKI
	err := common.Bootstrap(caCertsDir+"/ca.crt", caCertsDir+"/bootstrap.json")
	if err != nil {
		log.Fatal("Verification of ca certificate failed:", err)
	} else {
		log.Println("Successfully verified ca certificate.")
	}

	var privKey *ecdsa.PrivateKey
	if !common.FileExists(storagePrivKey) {
		// Generate a new key and write it to file
		privKey, _ = common.GenECDSAKey("P256")
		bts, _ := x509.MarshalECPrivateKey(privKey)
		common.WriteToPEMFile(storagePrivKey, "ECDSA PRIVATE KEY", bts)
		name := pkix.Name{
			Organization:       []string{"SCIONLab"},
			OrganizationalUnit: []string{"Storage"},
			Country:            []string{"CH"},
			Province:           []string{"Zurich"},
			Locality:           []string{"Zurich"},
		}
		bts, _ = common.GenCertSignRequest(name, privKey, []string{storageDNS}, []net.IP{net.ParseIP(storageIP)})
		common.WriteToPEMFile(storageCSR, "CERTIFICATE REQUEST", bts)
	}
	if !common.FileExists(storageCert) {
		if managerIP != "" {
			common.RequestAndObtainCert(caCertsDir, managerIP, managerUnverifPort, storageCert, storageCSR, "Storage", storageIP)
		} else {
			log.Fatal("No certificate found and no connection with manager. Please manually generate and upload a certificate for the csr.")
		}
	}

	// Register at manager
	if managerIP != "" {
		client := common.CreateHttpsClient(caCertsDir, storageCert, storagePrivKey)
		data, _ := json.Marshal(types.Storage{
			IA:         local.IA.String(),
			IP:         local.Host.IP().String(),
			Port:       fmt.Sprint(local.L4Port),
			ManagePort: managementAPIPort,
		})
		if err != nil {
			log.Fatal("Failed marshalling Storage struct:", err)
		}
		resp, err := client.Post("https://"+managerIP+":"+managerVerifPort+"/storages/register", "application/json", bytes.NewBuffer(data))
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
	log.Println("Started Storage process")

	// HTTPS server
	go func() {
		log.Println("Starting HTTPS server")

		srv := common.CreateHttpsServer(caCertsDir, storageIP, externalPort, &handler{http.Client{}}, tls.RequireAndVerifyClientCert)

		log.Fatal("HTTPS server listening error: ", srv.ListenAndServeTLS(storageCert, storagePrivKey))
	}()

	// SCION server
	go func() {
		log.Println("Starting SCION server")
		squic.Init(storagePrivKey, storageCert)

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

	// Management Server
	router := mux.NewRouter()
	// TODO: add some call here

	go func() {
		srv := &http.Server{
			Addr:    "127.0.0.1:" + localhostManagementPort,
			Handler: router,
		}
		log.Println("localhost HTTP server listening error: ", srv.ListenAndServe())
	}()

	srv := common.CreateHttpsServer(caCertsDir, storageIP, managementAPIPort, router, tls.RequireAndVerifyClientCert)
	log.Println("Starting HTTPS management server")
	log.Fatal("HTTPS server listening error: ", srv.ListenAndServeTLS(storageCert, storagePrivKey))
}

func handleQUICSession(qsess quic.Session, client http.Client) {
	log.Println("Received SCION request")
	// TODO: extension
	// Check if remote is authorized to write/read
	//if !Authorized(qsess.RemoteAddr()) {
	//	log.Println("Remote ", qsess.RemoteAddr(), "not authorized to scrape endpoint")
	//	return
	//}
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

	var path string
	if req.URL.Path == "/write" {
		path = "http://127.0.0.1:" + internalPort + writePath + "?" + "db=" + dbName
	} else if req.URL.Path == "/read" {
		path = "http://127.0.0.1:" + internalPort + readPath + "?" + "db=" + dbName
	}
	resp, err := client.Post(path, "application/x-protobuf", common.NewRequestFromQUIC(req).Body)
	if err != nil {
		log.Println("Failed contacting 127.0.0.1: ", err)
		return
	}

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

func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Println("Received HTTPS request")
	var resp *http.Response
	var err error
	//db_name, _ := req.URL.Query()["db"]
	path := "http://127.0.0.1:" + internalPort // + req.URL.Path + "?" + "db=" + db_name[0]
	if req.URL.Path == "/write" {
		resp, err = h.client.Post(path+writePath+"?"+"db="+dbName, "application/x-protobuf", req.Body)
	} else if req.URL.Path == "/read" {
		resp, err = h.client.Post(path+readPath+"?"+"db="+dbName, "application/x-protobuf", req.Body)
	} else {
		err = errors.New("Unsupported action: " + req.URL.Path)
	}
	if err != nil {
		log.Println("Request error:", err)
		w.WriteHeader(500)
		return
	}
	io.Copy(w, resp.Body)
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
}

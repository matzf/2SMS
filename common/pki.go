package common

import (
	"net/http"
	"log"
	"io/ioutil"
	"encoding/base64"
	"crypto/tls"
	"bytes"
	"time"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/baehless/2SMS/common/types"
	"encoding/json"
	"github.com/scionproto/scion/go/lib/crypto"
)

// TODO: handle trc updates
func Bootstrap(rootCertFile, bootstrapDataFile string, localIA addr.IA) error {
	// Load local IA configuration
	c, err := LoadConfig(localIA)
	if err != nil {
		return err
	}
	// Parse caData, which contains the manager's IA and raw certificate signature (sig) in json format
	jsonBytes, err := ioutil.ReadFile(bootstrapDataFile)
	if err != nil {
		return err
	}
	var bootstrapData types.BootstrapData
	err = json.Unmarshal(jsonBytes, &bootstrapData)
	if err != nil {
		return err
	}
	// Verify Trust Chain to manager's IA
	chain := c.Store.GetNewestChain(bootstrapData.IA)
	maxTRC := c.Store.GetNewestTRC(bootstrapData.IA.I)
	chain.Verify(bootstrapData.IA, maxTRC)

	// Get verifying key from cert
	verifyKey := chain.Leaf.SubjectSignKey

	// Load raw ca certificate
	sigInput, err := ioutil.ReadFile(rootCertFile)
	if err != nil {
		return err
	}

	// Verify the signature
	return crypto.Verify(sigInput, bootstrapData.RawSignature, verifyKey, "ed25519")
}

// Checks if the response contains the certificate and in case stores it in a file named `certFile`
func getCert(resp *http.Response, certFile string) {
	if resp.StatusCode == http.StatusBadRequest {
		log.Fatal("Error while requesting certificate:", 400)
	} else if resp.StatusCode == http.StatusNoContent {
		log.Println("No certificate available")
	} else if resp.StatusCode == http.StatusUnauthorized {
		log.Println("Not authorized to automatically obtain a certificate")
	} else if resp.StatusCode == http.StatusNotFound {
		log.Println("Certificate not ready yet")
	} else {
		log.Println("Certificate received")
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		crt := make([]byte, base64.StdEncoding.DecodedLen(len(data)))
		dec, err := base64.StdEncoding.Decode(crt, data)
		if err != nil {
			log.Fatal(err)
		}

		ioutil.WriteFile(certFile, crt[:dec], 0644)
	}
}

func RequestAndObtainCert(caCertsDir, managerAddress, managerPort, certFile, csrFile string, typ, ip string) {
	// If not present request certificate from manager and poll until provided
	caCertPool, err := NewCertPoolFromDir(caCertsDir)
	if err != nil {
		log.Fatal(err)
	}

	// Create HTTPS client
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      caCertPool,
			},
		},
	}

	// Request certificate and check if immediately provided
	log.Println("Requesting certificate")

	// Read csr from file
	bts, err := ioutil.ReadFile(csrFile)
	if err != nil {
		log.Fatal(err)
	}
	// Encode it to base64
	data := make([]byte, base64.StdEncoding.EncodedLen(len(bts)))
	base64.StdEncoding.Encode(data, bts)
	resp, err := client.Post("https://" + managerAddress + ":" + managerPort + "/certificate/request", "application/base64", bytes.NewBuffer(data))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	getCert(resp, certFile)
	// Repeatedly try to get the certificate
	for !FileExists(certFile) {
		time.Sleep(30 * time.Second)
		log.Println("Trying to get certificate")
		resp, err = client.Get("https://" + managerAddress + ":" + managerPort + "/certificates/"+ typ + "/" + ip + "/get")
		if err != nil {
			log.Fatal(err)
		}
		getCert(resp, certFile)
	}
}
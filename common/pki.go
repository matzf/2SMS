package common

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"github.com/netsec-ethz/2SMS/common/types"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/crypto"
	"io/ioutil"
	"log"
	"net/http"
	"time"
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

func RequestAndObtainCert(caCertsDir, managerAddress, managerPort, certFile, csrFile string, typ, ip string) {
	// If not present request certificate from manager and try until provided
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

	// Read csr from file
	bts, err := ioutil.ReadFile(csrFile)
	if err != nil {
		log.Fatal(err)
	}
	// Encode it to base64
	data := make([]byte, base64.StdEncoding.EncodedLen(len(bts)))
	base64.StdEncoding.Encode(data, bts)
	// Repeatedly try to request the certificate
	for !FileExists(certFile) {
		log.Println("Requesting certificate")
		resp, err := client.Post("https://" + managerAddress + ":" + managerPort + "/certificate/request", "application/base64", bytes.NewBuffer(data))
		if err != nil {
			log.Fatal(err)
		}
		if resp.StatusCode == http.StatusBadRequest {
			log.Fatal("Error while requesting certificate:", 400)
		} else if resp.StatusCode == http.StatusUnauthorized {
			log.Println("Not authorized to obtain a certificate")
			time.Sleep(30 * time.Second)
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
}
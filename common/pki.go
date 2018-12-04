package common

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"github.com/netsec-ethz/2SMS/common/types"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/crypto"
	"github.com/scionproto/scion/go/lib/crypto/cert"
)

func getSigningKey(IA addr.IA) []byte {
	asCertFileNameRegex, err := regexp.Compile(fmt.Sprintf(`^ISD\d+-AS%s-V\d+.crt$`, IA.A.FileFmt()))
	if err != nil {
		log.Fatalf("Internal error building regular expression \"%s\": %v", asCertFileNameRegex.String(), err)
	}
	scionDir := os.Getenv("SC")
	certsDir := scionDir + fmt.Sprintf("/gen/ISD%d/AS%s/endhost/certs/", IA.I, IA.A.FileFmt())
	fileInfos, err := ioutil.ReadDir(certsDir)
	if err != nil {
		log.Fatalf("Could not read endhost certs directory %s. Error is: %v", certsDir, err)
	}
	fileNames := []string{}
	for _, fi := range fileInfos {
		name := fi.Name()
		if asCertFileNameRegex.MatchString(name) {
			fileNames = append(fileNames, name)
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(fileNames)))
	filePath := filepath.Join(certsDir, fileNames[0])
	log.Printf("Loading AS signing key from %s", filePath)
	chain, err := cert.ChainFromFile(filePath, false)
	if err != nil {
		log.Fatalf("Cannot load AS certificate chain from %s: %v", filePath, err)
	}
	return chain.Leaf.SubjectSignKey
}

// Bootstrap boots the PKI validating the passed CA
func Bootstrap(rootCertFile, bootstrapDataFile string, localIA addr.IA) error {
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
	verifyKey := getSigningKey(localIA)
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
				RootCAs: caCertPool,
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
		resp, err := client.Post("https://"+managerAddress+":"+managerPort+"/certificate/request", "application/base64", bytes.NewBuffer(data))
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

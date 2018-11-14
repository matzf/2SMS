package common

import (
	"os"
	"encoding/pem"
	"crypto/x509/pkix"
	"crypto/ecdsa"
	"crypto/x509"
	"crypto/rand"
	"math/big"
	"io/ioutil"
	"strconv"
	"crypto/elliptic"
	"log"
	"github.com/pkg/errors"
	"net"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/crypto"
	"encoding/json"
	"github.com/netsec-ethz/2SMS/common/types"
)

// TODO: check type and handle errors
func WriteToPEMFile(fileName, typ string, bytes []byte) error {
	out, err := os.Create(fileName)
	if err != nil {
		return err
	}
	err = pem.Encode(out, &pem.Block{Type: typ, Bytes: bytes})
	out.Close()
	return nil
}

// Read first block in PEM file. May be a key, csr or certificate
func ReadCertFromPEMFile(fileName string) (*x509.Certificate, error) {
	bytes, err := ioutil.ReadFile(fileName)
	pemBlock, _ := pem.Decode(bytes)
	if pemBlock == nil {
		return nil, err
	}
	if pemBlock.Type == "CERTIFICATE" {
		return x509.ParseCertificate(pemBlock.Bytes)
	}
	return nil, errors.New("Unsupported type: " + pemBlock.Type)
}

func ReadCSRFromPEMFile(fileName string) (*x509.CertificateRequest, error) {
	bytes, err := ioutil.ReadFile(fileName)
	pemBlock, _ := pem.Decode(bytes)
	if pemBlock == nil {
		return nil, err
	}
	if pemBlock.Type == "CERTIFICATE REQUEST" {
		return x509.ParseCertificateRequest(pemBlock.Bytes)
	}
	return nil, errors.New("Unsupported type: " + pemBlock.Type)
}

func ReadECPrivKeyFromPEMFile(fileName string) (*ecdsa.PrivateKey, error) {
	bytes, err := ioutil.ReadFile(fileName)
	pemBlock, _ := pem.Decode(bytes)
	if pemBlock == nil {
		return nil, err
	}
	if pemBlock.Type == "ECDSA PRIVATE KEY" {
		return x509.ParseECPrivateKey(pemBlock.Bytes)
	}
	return nil, errors.New("Unsupported type: " + pemBlock.Type)
}

// Create a csr for the given key
func GenCertSignRequest(name pkix.Name, keys *ecdsa.PrivateKey, DNSNames []string, IPAddresses []net.IP) (csr []byte, err error){
	// step: generate a csr template
	var csrTemplate = x509.CertificateRequest{
		Subject:            name,
		SignatureAlgorithm: x509.ECDSAWithSHA512,
		DNSNames: DNSNames,
		IPAddresses: IPAddresses,
	}
	// step: generate the csr request
	csrCertificate, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, keys)
	if err != nil {
		return nil, err
	}

	return csrCertificate, nil
}

func LoadSerial(serialFile string) (*big.Int, error) {
	bytes, err := ioutil.ReadFile(serialFile)
	if err != nil {
		return big.NewInt(0), err
	}
	i, err := strconv.ParseInt(string(bytes), 0, 64)
	if err != nil {
		return big.NewInt(0), err
	}
	return big.NewInt(i), nil
}

func WriteSerial(serialFile string, value *big.Int) error {
	return ioutil.WriteFile(serialFile, []byte(strconv.FormatUint(value.Uint64(), 10)), 0644)
}

// TODO: error handling
// Generate new private key and self-signed certificate, link to SCION Control Plane PKI
func NewCA(name *pkix.Name, duration *Duration, privKeyFile, serialFile, certFile string, localIA addr.IA) (*CA, error) {
	ca := CA{PrivKeyFile: privKeyFile, SerialFile: serialFile, CertFile: certFile}
	// Generate new key and write it to file
	ca.privKey, _ = GenECDSAKey("P256")
	bytes, _ := x509.MarshalECPrivateKey(ca.privKey)
	WriteToPEMFile(ca.PrivKeyFile, "ECDSA PRIVATE KEY", bytes)
	// Create serial file
	ca.Serial = big.NewInt(1)
	WriteSerial(serialFile, ca.Serial)
	// Generate a self-signed certificate for the key and write it to file
	caCertBytes, err := ca.genCACert(duration, name)
	if err != nil {
		log.Fatal(err)
	}
	ca.cert, _ = x509.ParseCertificate(caCertBytes)
	WriteToPEMFile(ca.CertFile, "CERTIFICATE", caCertBytes)

	// Link to SCION Control Plane PKI
	// Load IA configuration
	c, err := LoadConfig(localIA)
	if err != nil {
		return nil, err
	}
	// Get raw signing key (signKey)
	signKey := c.GetSigningKey()
	// Read ca cert in pem format
	pemBytes, _ := ioutil.ReadFile(ca.CertFile)
	// Create signature of raw certificate
	rawSignature, err := crypto.Sign(pemBytes, signKey, "ed25519") // TODO: handle error
	// TWrite IA and signature to file in json format
	jsonBytes, err := json.Marshal(types.BootstrapData{IA: localIA, RawSignature: rawSignature}) // TODO: handle error
	ioutil.WriteFile("ca/bootstrap.json", jsonBytes, 0644)

	return &ca, nil
}

// Load private key, certificate and serial number
func LoadCA(privKeyFile, serialFile, certFile string) (*CA, error) {
	var err error
	ca := CA{PrivKeyFile: privKeyFile, SerialFile: serialFile, CertFile: certFile}
	ca.Serial, err = LoadSerial(serialFile)
	if err != nil {
		return nil, err
	}
	ca.privKey, err = ReadECPrivKeyFromPEMFile(privKeyFile)
	if err != nil {
		return nil, err
	}
	ca.cert, err = ReadCertFromPEMFile(certFile)
	if err != nil {
		return nil, err
	}
	return &ca, nil
}

func GenECDSAKey(typ string) (*ecdsa.PrivateKey, error) {
	var priv *ecdsa.PrivateKey
	// TODO: other cases + error
	switch typ {
	case "P256": priv, _ = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	}

	return priv, nil
}
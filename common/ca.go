package common

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"time"
)

// Main source: https://fale.io/blog/2017/06/05/create-a-pki-in-golang/

// TODO: move to types
type Duration struct {
	Years  int
	Months int
	Days   int
}

// TODO: move to types
type CA struct {
	Serial      *big.Int
	PrivKeyFile string
	SerialFile  string
	CertFile    string
	privKey     *ecdsa.PrivateKey
	cert        *x509.Certificate
}

// Generate a self-signed certificate for `priv` key
// TODO: adjust data in cert, error handling
func (ca *CA) genCACert(duration *Duration, name *pkix.Name) (cert []byte, err error) {
	ca_a := &x509.Certificate{
		SerialNumber:          ca.Serial,
		Subject:               *name,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(duration.Years, duration.Months, duration.Days),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	pub := &ca.privKey.PublicKey
	ca_b, err := x509.CreateCertificate(rand.Reader, ca_a, ca_a, pub, ca.privKey)
	if err != nil {
		return nil, err
	}
	return ca_b, nil
}

// Create a certificate for the given csr
func (ca *CA) GenCertFromCSR(csr *x509.CertificateRequest, duration *Duration) (cert []byte, err error) {
	// TODO: verify csr
	cert_a := &x509.Certificate{
		SerialNumber: ca.Serial,
		Subject:      csr.Subject,
		PublicKey:    csr.PublicKey,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(duration.Years, duration.Months, duration.Days),
		//SubjectKeyId: []byte{1, 2, 3, 4, 6}, // TODO: what's this for??
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
		DNSNames:    csr.DNSNames,
		IPAddresses: csr.IPAddresses,
	}
	// Sign the certificate
	cert, err = x509.CreateCertificate(rand.Reader, cert_a, ca.cert, csr.PublicKey, ca.privKey)
	if err != nil {
		return nil, err
	}
	ca.Serial = ca.Serial.Add(ca.Serial, big.NewInt(1))
	WriteSerial(ca.SerialFile, ca.Serial)
	return cert, err
}

// TODO: add also IPAddresses argument
func (ca *CA) GenCert(name pkix.Name, keys *ecdsa.PrivateKey, duration *Duration, DNSNames []string, IPAddresses []net.IP) (cert []byte, err error) {
	cert_a := &x509.Certificate{
		SerialNumber: ca.Serial,
		Subject:      name,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(duration.Years, duration.Months, duration.Days),
		//SubjectKeyId: []byte{1, 2, 3, 4, 6}, // TODO: what's this for??
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
		DNSNames:    DNSNames,
		IPAddresses: IPAddresses,
	}
	// Sign the certificate
	cert, err = x509.CreateCertificate(rand.Reader, cert_a, ca.cert, &keys.PublicKey, ca.privKey)
	if err != nil {
		return nil, err
	}
	ca.Serial = ca.Serial.Add(ca.Serial, big.NewInt(1))
	WriteSerial(ca.SerialFile, ca.Serial)
	return cert, err
}

package mtls

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"time"

	"github.com/cbluth/go/pkg/drbg"
)

func NewHTTPServer(secret []byte, serverURL string, server *http.Server) (*http.Server, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}
	cert, ca, err := generateCert(secret, u.Hostname())
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	ok := pool.AppendCertsFromPEM(ca)
	if !ok {
		return nil, fmt.Errorf("error appending ca cert")
	}
	server.TLSConfig = &tls.Config{
		PreferServerCipherSuites: true,
		ClientCAs:                pool,
		MinVersion:               tls.VersionTLS13,
		Certificates:             []tls.Certificate{*cert},
		ClientAuth:               tls.RequireAndVerifyClientCert,
	}
	return server, nil
}

func NewHTTPClient(secret []byte, serverURL string, client *http.Client) (*http.Client, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}
	cert, ca, err := generateCert(secret, u.Hostname())
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	ok := pool.AppendCertsFromPEM(ca)
	if !ok {
		return nil, fmt.Errorf("error appending ca cert")
	}
	// client.Transport
	client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:      pool,
			ServerName:   u.Hostname(),
			Certificates: []tls.Certificate{*cert},
		},
	}
	return client, nil
}

func generateCert(seed []byte, hostname string) (*tls.Certificate, []byte, error) {
	caPrivKey, ca, capem, err := seedCA(seed)
	if err != nil {
		return nil, nil, fmt.Errorf("generateCert: %v", err)
	}
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generateCert: %v", err)
	}
	asnb, err := asn1.Marshal(
		[]asn1.RawValue{
			{
				Tag:   asn1.TagInteger,
				Bytes: []byte(hostname),
				Class: asn1.ClassContextSpecific,
			},
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("generateCert: %v", err)
	}
	now := time.Now().UTC()
	cert := &x509.Certificate{
		Subject: pkix.Name{
			Organization:       []string{fmt.Sprintf("@%s", base64.StdEncoding.EncodeToString(pub))},
			OrganizationalUnit: []string{"x"},
		},
		Issuer: pkix.Name{
			CommonName:         fmt.Sprintf("@%s", base64.StdEncoding.EncodeToString(caPrivKey.Public().(ed25519.PublicKey))),
			OrganizationalUnit: []string{"x"},
		},
		NotBefore:    now,
		NotAfter:     now.AddDate(10, 0, 0),
		SerialNumber: big.NewInt(1),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		ExtraExtensions: []pkix.Extension{
			{ // SANs
				Value:    asnb,
				Critical: true,
				Id:       asn1.ObjectIdentifier{2, 5, 29, 17},
			},
		},
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, pub, caPrivKey)
	if err != nil {
		return nil, nil, fmt.Errorf("generateCert, create cert: %v", err)
	}
	certPEM := &bytes.Buffer{}
	err = pem.Encode(
		certPEM, &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: certBytes,
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("generateCert: %v", err)
	}
	pb, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, nil, fmt.Errorf("generateCert: %v", err)
	}
	certPrivKeyPEM := &bytes.Buffer{}
	err = pem.Encode(
		certPrivKeyPEM,
		&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: pb,
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("generateCert: %v", err)
	}
	serverCert, err := tls.X509KeyPair(certPEM.Bytes(), certPrivKeyPEM.Bytes())
	if err != nil {
		return nil, nil, fmt.Errorf("generateCert:tls.x509keypair: %v", err)
	}
	return &serverCert, capem, nil
}

func seedCA(seed []byte) (ed25519.PrivateKey, *x509.Certificate, []byte, error) {
	now := time.Now().UTC()
	pub, priv, err := ed25519.GenerateKey(drbg.New(seed))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("generateCert: %v", err)
	}
	ca := &x509.Certificate{
		BasicConstraintsValid: true,
		IsCA:                  true,
		Subject: pkix.Name{
			CommonName:         fmt.Sprintf("@%s", base64.StdEncoding.EncodeToString(pub)),
			OrganizationalUnit: []string{"x"},
		},
		NotBefore:    now,
		NotAfter:     now.AddDate(10, 0, 0),
		SerialNumber: big.NewInt(0),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
	}
	caBytes, err := x509.CreateCertificate(drbg.New(seed), ca, ca, pub, priv)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("generateCert: %v", err)
	}
	caPEM := &bytes.Buffer{}
	err = pem.Encode(
		caPEM,
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: caBytes,
		},
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("generateCert: %v", err)
	}
	return priv, ca, caPEM.Bytes(), nil
}

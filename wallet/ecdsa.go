package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"os"
	"time"
)

type Wallet struct {
	//eliptic curve digital algorithm
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

func generateECKey() (key *ecdsa.PrivateKey) {

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate ECDSA key: %s\n", err)
	}

	keyDer, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		log.Fatalf("Failed to serialize ECDSA key: %s\n", err)
	}

	keyBlock := pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyDer,
	}

	keyFile, err := os.Create("ec_key.pem")
	if err != nil {
		log.Fatalf("Failed to open ec_key.pem for writing: %s", err)
	}
	defer func() {
		keyFile.Close()
	}()

	pem.Encode(keyFile, &keyBlock)

	return
}

func generateCert(pub interface{}, priv *ecdsa.PrivateKey, filename string) {
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Docker, Inc."},
		},
		NotBefore: time.Now().Add(-time.Hour * 24 * 365),
		NotAfter:  time.Now().Add(time.Hour * 24 * 365),
	}
	certDer, err := x509.CreateCertificate(
		rand.Reader, &template, &template, pub, priv,
	)
	if err != nil {
		log.Fatalf("Failed to create certificate: %s\n", err)
	}

	certBlock := pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDer,
	}

	certFile, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to open '%s' for writing: %s", filename, err)
	}
	defer func() {
		certFile.Close()
	}()

	pem.Encode(certFile, &certBlock)
}

// Generate new Key Pair using ecdsa
func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()

	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}

	pub := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, pub
}

func MakeWallet() *Wallet {
	private, public := NewKeyPair()
	return &Wallet{private, public}
}

func main() {
	// Generate ECDSA P-256 Key
	log.Println("Generating an ECDSA P-256 Private Key")
	ECKey, _ := NewKeyPair()

	// pub := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	// Generate Self-Signed Certificate using ECDSA P-256 Key
	log.Println("Generating a Self-Signed Certificate from ECDSA P-256 Key")
	generateCert(&ECKey.PublicKey, &ECKey, "ec_cert.pem")

	// // Generate RSA 3072 Key
	// log.Println("Generating an RSA 3072 Private Key")
	// RSAKey := generateRSAKey()

	// // Generate Self-Signed Certificate using RSA 3072 Key
	// log.Println("Generating a Self-Signed Certificate from RSA 3072 Key")
	// generateCert(&RSAKey.PublicKey, RSAKey, "rsa_cert.pem")
}

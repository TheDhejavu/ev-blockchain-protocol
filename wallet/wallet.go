package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	checkSumlength = 1
	version        = byte(0x00) // hexadecimal representation of zero
)

// https://golang.org/pkg/crypto/ecdsa/
type Wallet struct {
	//eliptic curve digital algorithm
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

type WalletGroup struct {
	Main        *Wallet
	View        *Wallet
	Certificate []byte
}

func GenerateCert(pub interface{}, priv *ecdsa.PrivateKey) []byte {
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"DID, Inc."},
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

	var cert bytes.Buffer
	pem.Encode(&cert, &certBlock)
	return cert.Bytes()
}

// Generate new Key Pair using ecdsa
func NewKeyPair() (*ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()

	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}

	pub := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return private, pub
}

func MakeWallet() *Wallet {
	private, public := NewKeyPair()
	return &Wallet{*private, public}
}

func MakeWalletGroup() *WalletGroup {
	private, public := NewKeyPair()
	mainWallet := &Wallet{*private, public}
	viewWallet := MakeWallet()

	return &WalletGroup{
		View:        viewWallet,
		Main:        mainWallet,
		Certificate: GenerateCert(&private.PublicKey, private),
	}
}

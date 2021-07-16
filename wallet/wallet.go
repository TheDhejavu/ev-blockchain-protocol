package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	checkSumlength = 1
	version        = byte(0x00) // hexadecimal representation of zero
)

// https://golang.org/pkg/crypto/ecdsa/
type WalletView struct {
	// RSA algorithm
	PrivateKey rsa.PrivateKey
	PublicKey  []byte
}

type WalletMain struct {
	//eliptic curve digital algorithm
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

type WalletGroup struct {
	Main        *WalletMain
	View        *WalletView
	Certificate []byte
}

func (w *WalletView) Decrypt(ciphertext []byte) ([]byte, error) {
	block, _ := pem.Decode(w.PublicKey)
	if block == nil {
		return nil, errors.New("private key error!")
	}
	return rsa.DecryptPKCS1v15(rand.Reader, &w.PrivateKey, ciphertext)
}

func (w *WalletView) Encrypt(origData []byte) ([]byte, error) {
	block, _ := pem.Decode(w.PublicKey)
	if block == nil {
		return nil, errors.New("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub := pubInterface.(*rsa.PublicKey)
	return rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
}

func (w *WalletView) CanDecrypt(payload []byte) bool {
	if _, err := w.Decrypt(payload); err != nil {
		return false
	}
	return true
}

// Helper function for displaying wallet data in the console
func (w *WalletGroup) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("Public View Key: %x\n", w.View.PublicKey))
	// lines = append(lines, fmt.Sprintf("Private View Key: %x", w.View.PrivateKey))
	lines = append(lines, fmt.Sprintf("Public Main Key: %x:\n", w.Main.PublicKey))
	// lines = append(lines, fmt.Sprintf("Private Main Key: %x", w.Main.PrivateKey))

	return strings.Join(lines, "\n")
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

func NewRSAKeyPair() (*rsa.PrivateKey, []byte) {
	// Generate RSA Keys
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)

	if err != nil {
		log.Panic(err)
	}

	publicKey := &privateKey.PublicKey
	pubASN1, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		log.Error(err)
	}

	pubBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubASN1,
	})

	return privateKey, pubBytes
}

func MakeMainWallet() *WalletMain {
	private, public := NewKeyPair()
	return &WalletMain{*private, public}
}

func MakeViewWallet() *WalletView {
	private, public := NewRSAKeyPair()
	return &WalletView{*private, public}
}

func MakeWalletGroup() *WalletGroup {
	// private, public := NewKeyPair()
	mainWallet := MakeMainWallet()
	viewWallet := MakeViewWallet()

	return &WalletGroup{
		View: viewWallet,
		Main: mainWallet,
		Certificate: GenerateCert(
			&mainWallet.PrivateKey.PublicKey,
			&mainWallet.PrivateKey,
		),
	}
}

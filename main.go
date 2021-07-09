package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"
	"log"
	"math/big"

	"github.com/workspace/evoting/ev-blockchain-protocol/pkg/crypto/multisig"
	"github.com/workspace/evoting/ev-blockchain-protocol/pkg/crypto/ringsig"
	"github.com/workspace/evoting/ev-blockchain-protocol/wallet"
)

const numOfKeys = 3

var (
	DefaultCurve = elliptic.P256()
	keyring      *ringsig.PublicKeyRing
	privKey      *ecdsa.PrivateKey
	signature    *ringsig.RingSign
	keyRingByte  [][]byte
)

func main() {

	// Generate Decoy keys
	keyring = ringsig.NewPublicKeyRing(numOfKeys)
	for i := 0; i < numOfKeys; i++ {
		w := wallet.MakeWalletGroup()
		// add the public key part to the ring
		keyring.Add(w.Main.PrivateKey.PublicKey)
		keyRingByte = append(keyRingByte, w.Main.PublicKey)
	}
	// Main Key
	w := wallet.MakeWalletGroup()
	keyring.Add(w.Main.PrivateKey.PublicKey)
	keyRingByte = append(keyRingByte, w.Main.PublicKey)

	message := []byte("Big Brother Is Watching")
	// Sign message
	signature, err := ringsig.Sign(&w.Main.PrivateKey, keyring, message)
	if err != nil {
		log.Panic(err)
	}
	// message = []byte("Big Is Watching")
	fmt.Println(ringsig.Verify(keyring, message, signature))

	keyring = ringsig.NewPublicKeyRing(numOfKeys)
	for i := 0; i < len(keyRingByte); i++ {
		pub := keyRingByte[i]
		x := big.Int{}
		y := big.Int{}
		keyLen := len(pub)
		x.SetBytes(pub[:(keyLen / 2)])
		y.SetBytes(pub[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{Curve: DefaultCurve, X: &x, Y: &y}
		keyring.Add(rawPubKey)
	}

	byteSig := signature.ToByte()
	xSig := new(ringsig.RingSign)
	xSig.FromByte(byteSig)
	// message = []byte("Big Is Watching")
	fmt.Println(ringsig.Verify(keyring, message, xSig))

	mu := multisig.NewMultisig(2)
	mu.AddSignature([]byte("hello-word"), w.View.PublicKey, w.View.PrivateKey)
	mu.AddSignature([]byte("hello-word"), w.Main.PublicKey, w.Main.PrivateKey)
	mu.AddSignature([]byte("hello-word"), w.Main.PublicKey, w.Main.PrivateKey)
	r, _ := mu.Verify([]byte("hello-word"))
	fmt.Println(r)
}

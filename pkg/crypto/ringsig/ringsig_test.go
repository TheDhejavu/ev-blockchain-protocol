package ringsig

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"
	"runtime"
	"testing"

	"github.com/workspace/evoting/ev-blockchain-protocol/wallet"
)

var (
	DefaultCurve = elliptic.P256()
	keyring      *PublicKeyRing
	privKey      *ecdsa.PrivateKey
	signature    *RingSign
	keyRingByte  [][]byte
)

func BenchmarkRingsig(b *testing.B) {
	for _, size := range []int{1, 200, 400, 800, 1200, 1600, 2000} {
		benchmarkSign(b, size)
		benchmarkVerify(b, size)
	}
}

func benchmarkSign(b *testing.B, size int) {
	var err error
	runtime.GOMAXPROCS(8)
	b.ResetTimer()

	b.Run(fmt.Sprintf("NumOfKeys_%d", size), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Generate Decoy keys
			numOfKeys := uint(size)
			keyring = NewPublicKeyRing(numOfKeys)
			for i := 0; uint(i) < numOfKeys; i++ {
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
			signature, err = Sign(&w.Main.PrivateKey, keyring, message)
			if err != nil {
				fmt.Println(err.Error())
				b.FailNow()
			}
		}
	})
}

func benchmarkVerify(b *testing.B, size int) {
	runtime.GOMAXPROCS(8)
	b.ResetTimer()

	b.Run(fmt.Sprintf("NumOfKeys_veify%d", size), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			message := []byte("Big Brother Is Watching")
			Verify(keyring, message, signature)
		}
	})
}

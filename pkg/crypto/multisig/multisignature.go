package multisig

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"math/big"
)

type MultiSig struct {
	PubKeys [][]byte
	Sigs    [][]byte
}

// NewMultisig returns new
func NewMultisig(n int) *MultiSig {
	return &MultiSig{
		PubKeys: make([][]byte, 0, n),
		Sigs:    make([][]byte, 0, n),
	}
}

// AddSignature adds a signature to the multisig
func (sig *MultiSig) AddSignature(dataToSign []byte, PubKey []byte, privKey ecdsa.PrivateKey) {
	r, s, err := ecdsa.Sign(rand.Reader, &privKey, dataToSign)
	if err != nil {
		panic(err)
	}
	signature := append(r.Bytes(), s.Bytes()...)
	sig.Sigs = append(sig.Sigs, signature)
	sig.PubKeys = append(sig.PubKeys, PubKey)
}

// Verify all signatures of the multisig
func (sig *MultiSig) Verify(data []byte) (bool, error) {

	for i := 0; i < len(sig.PubKeys); i++ {
		r := big.Int{}
		s := big.Int{}
		sigLen := len(sig.Sigs[i])
		r.SetBytes(sig.Sigs[i][:(sigLen / 2)])
		s.SetBytes(sig.Sigs[i][(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		pubKey := sig.PubKeys[i]
		keyLen := len(pubKey)
		x.SetBytes(pubKey[:(keyLen / 2)])
		y.SetBytes(pubKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{Curve: elliptic.P256(), X: &x, Y: &y}

		return ecdsa.Verify(&rawPubKey, data, &r, &s), nil
	}

	return false, nil
}

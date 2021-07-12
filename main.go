package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"
	"log"
	"math/big"

	blockchain "github.com/workspace/evoting/ev-blockchain-protocol/core"
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
	signers      [][]byte
	privKeys     []*ecdsa.PrivateKey
	candidates   [][]byte
	sigWitness   [][]byte
	keyHash      = []byte("election_x")
)

func TestCrypto() {
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
	r, _ := mu.Verify([]byte("hello-word"))
	fmt.Println(r)
}
func NewElectionEnd(txId []byte, txOut string) blockchain.Transaction {
	txIn := blockchain.NewElectionTxInput(
		keyHash,
		txId,
		txOut,
		signers,
		sigWitness,
	)
	mu := multisig.NewMultisig(1)
	mu.AddSignature(
		txIn.ElectionTx.ToByte(),
		signers[0],
		*privKeys[0],
	)

	electionTx, _ := blockchain.NewTransaction(
		blockchain.ELECTION_TX_TYPE,
		*txIn,
		blockchain.TxOutput{},
	)

	electionTx.Input.ElectionTx.SigWitness = mu.Sigs
	electionTx.Input.ElectionTx.Signers = mu.PubKeys

	return *electionTx
}
func NewElectionStart() blockchain.Transaction {
	var totalPeople int64
	totalPeople = 100
	for i := 0; i < 2; i++ {
		w := wallet.MakeWalletGroup()
		pubkey := fmt.Sprintf("		PubKey: %x", w.View.PublicKey)
		fmt.Println(pubkey)
		candidates = append(candidates, w.View.PublicKey)
	}

	txOut := blockchain.NewElectionTxOutput(
		"Presidential Election",
		"President",
		keyHash,
		nil,
		nil,
		candidates,
		totalPeople,
	)

	mu := multisig.NewMultisig(1)
	w := wallet.MakeWalletGroup()
	mu.AddSignature(
		txOut.ElectionTx.ToByte(),
		w.Main.PublicKey,
		w.Main.PrivateKey,
	)

	privKeys = append(privKeys, &w.Main.PrivateKey)

	sigWitness = mu.Sigs
	signers = mu.PubKeys

	electionTx, _ := blockchain.NewTransaction(
		blockchain.ELECTION_TX_TYPE,
		blockchain.TxInput{},
		*txOut,
	)

	electionTx.Output.ElectionTx.SigWitness = sigWitness
	electionTx.Output.ElectionTx.Signers = signers

	return *electionTx
}

func main() {
	txElectionStart := NewElectionStart()
	txElectionEnd := NewElectionEnd(
		txElectionStart.ID,
		txElectionStart.Output.ElectionTx.ID,
	)

	fmt.Println(
		"VERIFY_ELECTION_START",
		txElectionStart.Verify(blockchain.Transaction{}),
	)
	fmt.Println(
		"VERIFY_ELECTION_END",
		txElectionEnd.Verify(txElectionStart),
	)
}

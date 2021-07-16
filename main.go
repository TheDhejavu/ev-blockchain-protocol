package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"
	"log"
	"math/big"
	"time"

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
	SigWitnesses [][]byte
	keyHash      = []byte("election_x")
	sysWallet    *wallet.WalletGroup
)

func GenerateMainWallet() {
	// Main Key
	keyring = ringsig.NewPublicKeyRing(numOfKeys)
	sysWallet = wallet.MakeWalletGroup()
	keyring.Add(sysWallet.Main.PrivateKey.PublicKey)
	keyRingByte = append(keyRingByte, sysWallet.Main.PublicKey)
}
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
	mu.AddSignature([]byte("hello-word"), w.Main.PublicKey, w.Main.PrivateKey)
	mu.AddSignature([]byte("hello-word"), w.Main.PublicKey, w.Main.PrivateKey)
	r, _ := mu.Verify([]byte("hello-word"))
	fmt.Println(r)
}
func NewElectionEnd(txOut []byte, utxo *blockchain.UnusedXTOSet) *blockchain.Transaction {
	var electionTx *blockchain.Transaction

	txIn := blockchain.NewElectionTxInput(
		keyHash,
		txOut,
		signers,
		SigWitnesses,
	)
	mu := multisig.NewMultisig(1)
	mu.AddSignature(
		txIn.ElectionTx.ToByte(),
		signers[0],
		*privKeys[0],
	)

	electionTx, _ = blockchain.NewTransaction(
		blockchain.ELECTION_TX_TYPE,
		keyHash,
		*txIn,
		blockchain.TxOutput{},
		utxo,
	)

	electionTx.Input.ElectionTx.SigWitnesses = mu.Sigs
	electionTx.Input.ElectionTx.Signers = mu.PubKeys

	return electionTx
}
func NewElectionStart() *blockchain.Transaction {
	var eTx *blockchain.Transaction

	var totalPeople int64
	totalPeople = 100
	for i := 0; i < 2; i++ {
		w := wallet.MakeWalletGroup()
		candidates = append(candidates, w.Main.PublicKey)
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

	SigWitnesses = mu.Sigs
	signers = mu.PubKeys

	eTx, _ = blockchain.NewTransaction(
		blockchain.ELECTION_TX_TYPE,
		keyHash,
		blockchain.TxInput{},
		*txOut,
		&blockchain.UnusedXTOSet{},
	)

	eTx.Output.ElectionTx.SigWitnesses = SigWitnesses
	eTx.Output.ElectionTx.Signers = signers

	return eTx
}

func NewAccreditationEnd(txId []byte, txOut []byte, count int64, utxo *blockchain.UnusedXTOSet) *blockchain.Transaction {
	var acTx *blockchain.Transaction

	txAcIn := blockchain.NewAccreditationTxInput(
		keyHash,
		txId,
		txOut,
		nil,
		nil,
		count,
		time.Now().Unix(),
	)
	mu := multisig.NewMultisig(1)
	mu.AddSignature(
		txAcIn.AccreditationTx.ToByte(),
		signers[0],
		*privKeys[0],
	)

	acTx, _ = blockchain.NewTransaction(
		blockchain.ACCREDITATION_TX_TYPE,
		keyHash,
		*txAcIn,
		blockchain.TxOutput{},
		utxo,
	)

	acTx.Input.AccreditationTx.SigWitnesses = mu.Sigs
	acTx.Input.AccreditationTx.Signers = mu.PubKeys

	return acTx
}

func NewAccreditationStart(txID []byte) *blockchain.Transaction {
	var eaTx *blockchain.Transaction
	txAccreditationOut := blockchain.NewAccreditationTxOutput(
		keyHash,
		txID,
		nil,
		nil,
		time.Now().Unix(),
	)

	mu := multisig.NewMultisig(1)
	mu.AddSignature(
		txAccreditationOut.AccreditationTx.ToByte(),
		signers[0],
		*privKeys[0],
	)

	eaTx, _ = blockchain.NewTransaction(
		blockchain.ACCREDITATION_TX_TYPE,
		keyHash,
		blockchain.TxInput{},
		*txAccreditationOut,
		&blockchain.UnusedXTOSet{},
	)

	eaTx.Output.AccreditationTx.SigWitnesses = mu.Sigs
	eaTx.Output.AccreditationTx.Signers = mu.PubKeys

	return eaTx
}

func NewVotingEnd(txId []byte, txOut []byte, utxo *blockchain.UnusedXTOSet) *blockchain.Transaction {
	var vTx *blockchain.Transaction

	txVotingIn := blockchain.NewVotingTxInput(
		keyHash,
		txId,
		txOut,
		nil,
		nil,
		time.Now().Unix(),
	)
	mu := multisig.NewMultisig(1)
	mu.AddSignature(
		txVotingIn.VotingTx.ToByte(),
		signers[0],
		*privKeys[0],
	)

	vTx, _ = blockchain.NewTransaction(
		blockchain.VOTING_TX_TYPE,
		keyHash,
		*txVotingIn,
		blockchain.TxOutput{},
		utxo,
	)

	vTx.Input.VotingTx.SigWitnesses = mu.Sigs
	vTx.Input.VotingTx.Signers = mu.PubKeys

	return vTx
}

func NewVotingStart(txId []byte) *blockchain.Transaction {
	var vTx *blockchain.Transaction

	txVotingOut := blockchain.NewVotingTxOutput(
		keyHash,
		txId,
		nil,
		nil,
		time.Now().Unix(),
	)
	mu := multisig.NewMultisig(1)
	mu.AddSignature(
		txVotingOut.VotingTx.ToByte(),
		signers[0],
		*privKeys[0],
	)

	vTx, _ = blockchain.NewTransaction(
		blockchain.VOTING_TX_TYPE,
		keyHash,
		blockchain.TxInput{},
		*txVotingOut,
		&blockchain.UnusedXTOSet{},
	)

	vTx.Output.VotingTx.SigWitnesses = mu.Sigs
	vTx.Output.VotingTx.Signers = mu.PubKeys

	return vTx
}

func NewBallot(txId []byte) *blockchain.Transaction {
	var bTx *blockchain.Transaction
	secretMessage := []byte("This is my ballot secret message")
	msg, _ := sysWallet.View.Encrypt(secretMessage)

	bTxOut := blockchain.NewBallotTxOutput(
		keyHash,
		msg,
		txId,
		nil,
		nil,
		nil,
		time.Now().Unix(),
	)
	mu := multisig.NewMultisig(1)
	mu.AddSignature(
		bTxOut.BallotTx.ToByte(),
		signers[0],
		*privKeys[0],
	)

	bTx, _ = blockchain.NewTransaction(
		blockchain.BALLOT_TX_TYPE,
		keyHash,
		blockchain.TxInput{},
		*bTxOut,
		&blockchain.UnusedXTOSet{},
	)

	bTx.Output.BallotTx.SigWitnesses = mu.Sigs
	bTx.Output.BallotTx.Signers = mu.PubKeys

	// Generate Decoy keys
	for i := 0; i < numOfKeys-1; i++ {
		w := wallet.MakeWalletGroup()
		// add the public key part to the ring
		keyring.Add(w.Main.PrivateKey.PublicKey)
		keyRingByte = append(keyRingByte, w.Main.PublicKey)
	}
	bTx.Output.BallotTx.PubKeys = keyRingByte

	return bTx
}

func CastBallot(txId []byte, txOut []byte, utxo *blockchain.UnusedXTOSet) *blockchain.Transaction {
	var bTx *blockchain.Transaction

	bTxIn := blockchain.NewBallotTxInput(
		keyHash,
		candidates[0],
		txId,
		txOut,
		nil,
		nil,
		time.Now().Unix(),
	)

	bTx, _ = blockchain.NewTransaction(
		blockchain.BALLOT_TX_TYPE,
		keyHash,
		*bTxIn,
		blockchain.TxOutput{},
		utxo,
	)

	// Sign message
	signature, err := ringsig.Sign(
		&sysWallet.Main.PrivateKey,
		keyring,
		bTxIn.BallotTx.ToByte(),
	)
	if err != nil {
		log.Panic(err)
	}

	bTx.Input.BallotTx.Signature = signature.ToByte()
	bTx.Input.BallotTx.PubKeys = keyRingByte

	return bTx
}

func main() {
	GenerateMainWallet()
	utxo := blockchain.NewUnusedXTOSet(&blockchain.Blockchain{})
	txElectionStart := NewElectionStart()
	txElectionEnd := NewElectionEnd(txElectionStart.ID, utxo)

	txAccreditationStart := NewAccreditationStart(txElectionStart.ID)
	txAccreditationEnd := NewAccreditationEnd(
		txElectionStart.ID,
		txAccreditationStart.ID,
		100,
		utxo,
	)

	txVotingStart := NewVotingStart(txElectionStart.ID)
	txVotingEnd := NewVotingEnd(txElectionStart.ID, txVotingStart.ID, utxo)

	txNewBallot := NewBallot(txElectionStart.ID)
	txCastBallot := CastBallot(txElectionStart.ID, txNewBallot.ID, utxo)

	fmt.Println(
		"VERIFY_ELECTION_START",
		txElectionStart.Verify(blockchain.Transaction{}),
	)

	fmt.Println(
		"VERIFY_ELECTION_END",
		txElectionEnd.Verify(*txElectionStart),
	)

	fmt.Println(
		"VERIFY_ACCREDITATION_START",
		txAccreditationStart.Verify(*txElectionStart),
	)

	fmt.Println(
		"VERIFY_ACCREDITATION_END",
		txAccreditationEnd.Verify(*txAccreditationStart),
	)

	fmt.Println(
		"VERIFY_VOTING_START",
		txVotingStart.Verify(*txElectionStart),
	)

	fmt.Println(
		"VERIFY_VOTING_END",
		txVotingEnd.Verify(*txVotingStart),
	)

	fmt.Println(
		"VERIFY_NEW_BALLOT",
		txNewBallot.Verify(*txElectionStart),
	)

	fmt.Println(
		"VERIFY_CAST_BALLOT",
		txCastBallot.Verify(*txNewBallot),
	)
	fmt.Println(txCastBallot)

	newBlock := blockchain.Genesis(txElectionStart, 1)
	fmt.Println(newBlock)
}

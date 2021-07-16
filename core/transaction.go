package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"math/big"
	"math/rand"
	"reflect"
	"strings"
	"time"

	logger "github.com/sirupsen/logrus"
	"github.com/workspace/evoting/ev-blockchain-protocol/pkg/crypto/multisig"
	"github.com/workspace/evoting/ev-blockchain-protocol/pkg/crypto/ringsig"
)

const VOTING_TX_TYPE = "voting_tx"
const ACCREDITATION_TX_TYPE = "accreditation_tx"
const BALLOT_TX_TYPE = "ballot_tx"
const ELECTION_TX_TYPE = "election_tx"

var (
	DefaultCurve = elliptic.P256()
	TxTypes      = []string{
		VOTING_TX_TYPE,
		ACCREDITATION_TX_TYPE,
		BALLOT_TX_TYPE,
		ELECTION_TX_TYPE,
	}
)

type Transaction struct {
	ID      []byte
	Input   TxInput
	Output  TxOutput
	Nonce   uint64
	Type    string
	KeyHash []byte
}

// Create new Transaction
func NewTransaction(txType string, keyHash []byte, input TxInput, output TxOutput, utxo *UnusedXTOSet) (*Transaction, error) {
	rand.Seed(time.Now().Unix())
	
	tx := Transaction{
		ID:      nil,
		Input:   input,
		Output:  output,
		Nonce:   rand.Uint64(),
		Type:    txType,
		KeyHash: keyHash,
	}
	tx.ID = tx.Hash()
	return &tx, nil
}

func (tx *Transaction) inputSet() bool {
	return reflect.DeepEqual(tx.Input, TxInput{}) == false
}

func (tx *Transaction) outputSet() bool {
	return reflect.DeepEqual(tx.Output, TxOutput{}) == false
}
func (tx *Transaction) verifyElectionTx(prevTx Transaction) (verified bool) {
	var err error
	electionOut := tx.Output.ElectionTx
	electionIn := tx.Input.ElectionTx
	// fmt.Println(electionIn.IsSet(), electionOut.IsSet())
	if electionOut.IsSet() {
		ms := multisig.MultiSig{
			PubKeys: electionOut.Signers,
			Sigs:    electionOut.SigWitnesses,
		}

		verified, err := ms.Verify(electionOut.ToByte())
		if err != nil {
			logger.Error(err)
		}
		return verified
	}

	if electionIn.IsSet() {
		// fmt.Println("Election Input Is Set")
		if prevTx.IsSet() == false {
			return false
		}
		ms := multisig.MultiSig{
			PubKeys: prevTx.Output.ElectionTx.Signers,
			Sigs:    electionIn.SigWitnesses,
		}

		verified, err = ms.Verify(electionIn.ToByte())
		if err != nil {
			logger.Error(err)
		}
		return
	}

	return
}

func (tx *Transaction) verifyAccreditationTx(prevTx Transaction) bool {
	accreditationOut := tx.Output.AccreditationTx
	accreditationIn := tx.Input.AccreditationTx

	// fmt.Println(accreditationOut.IsSet(), accreditationIn.IsSet())
	if accreditationOut.IsSet() {
		ms := multisig.MultiSig{
			PubKeys: accreditationOut.Signers,
			Sigs:    accreditationOut.SigWitnesses,
		}
		verified, err := ms.Verify(accreditationOut.ToByte())
		if err != nil {
			logger.Error(err)
		}
		return verified
	}

	// fmt.Println("AC_START", accreditationIn.IsSet())
	if accreditationIn.IsSet() {
		if prevTx.IsSet() == false {
			return false
		}

		ms := multisig.MultiSig{
			PubKeys: prevTx.Output.AccreditationTx.Signers,
			Sigs:    accreditationIn.SigWitnesses,
		}

		txCopy := tx.Input.AccreditationTx.TrimmedCopy()
		txCopy.ElectionKeyHash = prevTx.KeyHash
		// Verify data
		verified, err := ms.Verify(txCopy.ToByte())
		if err != nil {
			logger.Error(err)
		}
		return verified
	}
	return false
}

func (tx *Transaction) verifyVotingTx(prevTx Transaction) bool {
	votingOut := tx.Output.VotingTx
	votingIn := tx.Input.VotingTx

	fmt.Println(votingOut.IsSet(), votingIn.IsSet())
	if votingOut.IsSet() {
		ms := multisig.MultiSig{
			PubKeys: votingOut.Signers,
			Sigs:    votingOut.SigWitnesses,
		}
		// txCopy := votingOut.TrimmedCopy()
		// txCopy.ElectionKeyHash = []byte("sm")
		verified, err := ms.Verify(votingOut.ToByte())
		if err != nil {
			logger.Error(err)
		}
		return verified
	}

	// fmt.Println("AC_START", votingIn.IsSet())
	if votingIn.IsSet() {
		if prevTx.IsSet() == false {
			return false
		}

		ms := multisig.MultiSig{
			PubKeys: prevTx.Output.VotingTx.Signers,
			Sigs:    votingIn.SigWitnesses,
		}

		txCopy := tx.Input.VotingTx.TrimmedCopy()
		txCopy.ElectionKeyHash = prevTx.Output.VotingTx.ElectionKeyHash
		// Verify data
		verified, err := ms.Verify(txCopy.ToByte())
		if err != nil {
			logger.Error(err)
		}
		return verified
	}
	return false
}
func (tx *Transaction) verifyBallotTx(prevTx Transaction) bool {
	ballotOut := tx.Output.BallotTx
	ballotIn := tx.Input.BallotTx

	// fmt.Println(ballotIn.IsSet(), ballotOut.IsSet())
	if ballotOut.IsSet() {
		ms := multisig.MultiSig{
			PubKeys: ballotOut.Signers,
			Sigs:    ballotOut.SigWitnesses,
		}
		verified, err := ms.Verify(ballotOut.ToByte())
		if err != nil {
			logger.Error(err)
		}
		return verified
	}

	if ballotIn.IsSet() {
		if prevTx.IsSet() == false {
			return false
		}

		numOfKeys := uint(len(ballotIn.PubKeys))
		keyring := ringsig.NewPublicKeyRing(numOfKeys)
		keyRingByte := ballotIn.PubKeys
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

		signature := new(ringsig.RingSign)
		signature.FromByte(ballotIn.Signature)
		txCopy := ballotIn.TrimmedCopy()
		txCopy.ElectionKeyHash = prevTx.Output.BallotTx.ElectionKeyHash
		txCopy.PubKeys = prevTx.Output.BallotTx.PubKeys

		verified := ringsig.Verify(keyring, txCopy.ToByte(), signature)

		return verified
	}
	return false
}
func (tx *Transaction) Verify(prevTx Transaction) bool {
	switch tx.Type {
	case ELECTION_TX_TYPE:
		// Verify election Transaction
		return tx.verifyElectionTx(prevTx)
	case ACCREDITATION_TX_TYPE:
		// Verify Accreditation Transaction
		return tx.verifyAccreditationTx(prevTx)
	case VOTING_TX_TYPE:
		// Verify Voting Transaction
		return tx.verifyVotingTx(prevTx)
	case BALLOT_TX_TYPE:
		// Verify ballot Transaction
		return tx.verifyBallotTx(prevTx)
	}

	return false
}

func (tx *Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	encode := gob.NewEncoder(&encoded)
	err := encode.Encode(tx)
	if err != nil {
		logger.Panic(err)
	}

	return encoded.Bytes()
}

func DeserializeTransaction(data []byte) Transaction {
	var transaction Transaction

	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&transaction)
	if err != nil {
		logger.Panic(err)
	}
	return transaction
}

func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())
	return hash[:]
}

func (tx *Transaction) IsSet() bool {
	return reflect.DeepEqual(tx, Transaction{}) == false
}

// Helper function for displaying transaction data in the console
func (tx *Transaction) String() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("-TRANSACTION: \n TX_ID: %x", tx.ID))

	switch tx.Type {
	case ELECTION_TX_TYPE:
		lines = append(lines, tx.Input.ElectionTx.String())
		lines = append(lines, tx.Output.ElectionTx.String())
	case ACCREDITATION_TX_TYPE:
		lines = append(lines, tx.Input.AccreditationTx.String())
		lines = append(lines, tx.Output.AccreditationTx.String())
	case VOTING_TX_TYPE:
		lines = append(lines, tx.Input.VotingTx.String())
		lines = append(lines, tx.Output.VotingTx.String())
	case BALLOT_TX_TYPE:
		lines = append(lines, tx.Input.BallotTx.String())
		lines = append(lines, tx.Output.BallotTx.String())
	}

	return strings.Join(lines, "\n")
}

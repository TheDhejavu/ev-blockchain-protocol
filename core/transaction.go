package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"math/rand"
	"reflect"
	"time"

	logger "github.com/sirupsen/logrus"
	"github.com/workspace/evoting/ev-blockchain-protocol/pkg/crypto/multisig"
)

const VOTING_TX_TYPE = "voting_tx"
const ACCREDITATION_TX_TYPE = "accreditation_tx"
const BALLOT_TX_TYPE = "ballot_tx"
const ELECTION_TX_TYPE = "election_tx"

var (
	TxTypes = []string{
		VOTING_TX_TYPE,
		ACCREDITATION_TX_TYPE,
		BALLOT_TX_TYPE,
		ELECTION_TX_TYPE,
	}
)

type Transaction struct {
	ID     []byte
	Input  TxInput
	Output TxOutput
	Nonce  uint64
	Type   string
}

// Create new Transaction
func NewTransaction(txType string, input TxInput, output TxOutput) (*Transaction, error) {
	rand.Seed(time.Now().Unix())

	tx := Transaction{
		ID:     nil,
		Input:  input,
		Output: output,
		Nonce:  rand.Uint64(),
		Type:   txType,
	}
	tx.ID = tx.Hash()
	return &tx, nil
}

func (tx *Transaction) Verify(prevTx Transaction) bool {
	if tx.Type == ELECTION_TX_TYPE {
		electionOut := tx.Output.ElectionTx
		electionIn := tx.Input.ElectionTx

		if electionOut.IsSet() {
			ms := multisig.MultiSig{
				PubKeys: electionOut.Signers,
				Sigs:    electionOut.SigWitness,
			}

			verified, err := ms.Verify(tx.Output.ElectionTx.ToByte())
			if err != nil {
				logger.Error(err)
			}
			return verified
		}

		if electionIn.IsSet() {
			if prevTx.IsSet() == false {
				return false
			}
			ms := multisig.MultiSig{
				PubKeys: prevTx.Output.ElectionTx.Signers,
				Sigs:    electionIn.SigWitness,
			}

			txCopy := tx.Input.ElectionTx.TrimmedCopy()
			txCopy.ElectionKeyHash = prevTx.Output.ElectionTx.ElectionKeyHash
			// Verify data
			verified, err := ms.Verify(txCopy.ToByte())
			if err != nil {
				logger.Error(err)
			}
			return verified
		}
	}
	if tx.Type == ACCREDITATION_TX_TYPE {

		return true
	}

	if tx.Type == VOTING_TX_TYPE {

		return true
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

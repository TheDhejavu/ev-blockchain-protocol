package blockchain

import (
	"fmt"
	"reflect"

	"github.com/google/uuid"
)

// INITIALIZE ELECTION
// Init  election TxOutput
type TxElectionOutput struct {
	ID              string
	Signers         [][]byte
	SigWitness      [][]byte
	ElectionKeyHash []byte
	Title           string
	Despcription    string
	TotalPeople     int64
	Candidates      [][]byte
}

// End Election TxInput
type TxElectionInput struct {
	TxID            []byte
	Signers         [][]byte
	SigWitness      [][]byte
	TxOut           string
	ElectionKeyHash []byte
}

// NewTxAccreditationInput Stops Accreditation  Phase
func NewElectionTxInput(keyHash, txId []byte, txOut string, signers, sigWitness [][]byte) *TxInput {
	tx := &TxInput{
		ElectionTx: TxElectionInput{
			TxID:            txId,
			Signers:         signers,
			SigWitness:      sigWitness,
			ElectionKeyHash: keyHash,
			TxOut:           txOut,
		},
	}
	return tx
}

// NewTxAccreditationTxOutput Starts Accreditation Phase
func NewElectionTxOutput(title, desp string, keyHash []byte, signers, sigWitness, candidates [][]byte, totalPeople int64) *TxOutput {
	tx := &TxOutput{
		ElectionTx: TxElectionOutput{
			ID:              "",
			Signers:         signers,
			SigWitness:      sigWitness,
			ElectionKeyHash: keyHash,
			Title:           title,
			Despcription:    desp,
			TotalPeople:     totalPeople,
			Candidates:      candidates,
		},
	}
	uuid, _ := uuid.NewUUID()
	tx.ElectionTx.ID = uuid.String()
	return tx
}

// Convert Election output to Byte for verification and signing purposes
func (tx TxElectionOutput) TrimmedCopy() TxElectionOutput {
	txCopy := TxElectionOutput{
		tx.ID,
		nil,
		nil,
		tx.ElectionKeyHash,
		tx.Title,
		tx.Despcription,
		tx.TotalPeople,
		tx.Candidates,
	}
	return txCopy
}

// Convert Election output to Byte for verification and signing purposes
func (tx TxElectionOutput) ToByte() []byte {
	txCopy := tx.TrimmedCopy()
	data := fmt.Sprintf("%x\n", txCopy)
	return []byte(data)
}

func (tx TxElectionOutput) IsSet() bool {
	return reflect.DeepEqual(tx, TxElectionOutput{}) == false
}

// Trim election input data
func (tx TxElectionInput) TrimmedCopy() TxElectionInput {
	txCopy := TxElectionInput{
		tx.TxID,
		nil,
		nil,
		tx.TxOut,
		tx.ElectionKeyHash,
	}
	return txCopy
}

// Convert Election output to Byte for verification and signing purposes
func (tx TxElectionInput) ToByte() []byte {
	txCopy := tx.TrimmedCopy()
	data := fmt.Sprintf("%x\n", txCopy)
	return []byte(data)
}

func (tx TxElectionInput) IsSet() bool {
	return reflect.DeepEqual(tx, TxElectionInput{}) == false
}

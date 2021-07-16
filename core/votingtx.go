package blockchain

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"reflect"
	"strings"

	"github.com/google/uuid"
)

// INITIALIZE VOTE
// Start Vote TxTxOutput
type TxVotingOutput struct {
	ID              string
	TxID            []byte
	Signers         [][]byte
	SigWitnesses    [][]byte
	ElectionKeyHash []byte
	Timestamp       int64
}

// End Vote TxInput
type TxVotingInput struct {
	TxID            []byte
	Signers         [][]byte // TransVotingtion Signatures from signers
	SigWitnesses    [][]byte
	ElectionKeyHash []byte
	TxOut           []byte
	Timestamp       int64
}

// NewTxVotingInput  ENDS Voting Phase
func NewVotingTxInput(keyHash, txId []byte, txOut []byte, signers, SigWitnesses [][]byte, timestamp int64) *TxInput {
	tx := &TxInput{
		VotingTx: TxVotingInput{
			TxID:            txId,
			Signers:         signers,
			SigWitnesses:    SigWitnesses,
			TxOut:           txOut,
			ElectionKeyHash: keyHash,
			Timestamp:       timestamp,
		},
	}
	return tx
}

// NewTxVotingOutput BEGINS Voting Phase
func NewVotingTxOutput(keyHash []byte, txId []byte, signers, SigWitnesses [][]byte, timestamp int64) *TxOutput {
	tx := &TxOutput{
		VotingTx: TxVotingOutput{
			TxID:            txId,
			Signers:         signers,
			SigWitnesses:    SigWitnesses,
			ElectionKeyHash: keyHash,
			Timestamp:       timestamp,
		},
	}
	uuid, _ := uuid.NewUUID()
	tx.VotingTx.ID = uuid.String()
	return tx
}

func (TxOut *TxVotingOutput) IsLockWithKey(ElectionKeyHash []byte) bool {
	return bytes.Compare(TxOut.ElectionKeyHash, ElectionKeyHash) == 0
}

func (tx *TxVotingOutput) IsSet() bool {
	return reflect.DeepEqual(tx, &TxVotingOutput{}) == false
}

// Convert Election output to Byte for verification and signing purposes
func (tx *TxVotingOutput) TrimmedCopy() TxVotingOutput {
	txCopy := TxVotingOutput{
		"",
		tx.TxID,
		nil,
		nil,
		tx.ElectionKeyHash,
		tx.Timestamp,
	}
	return txCopy
}

// Convert Election output to Byte for verification and signing purposes
func (tx *TxVotingOutput) ToByte() []byte {
	var hash [32]byte

	txCopy := tx.TrimmedCopy()

	hash = sha256.Sum256([]byte(fmt.Sprintf("%x", txCopy)))
	return hash[:]
}

// Trim election input data
func (tx *TxVotingInput) TrimmedCopy() TxVotingInput {
	txCopy := TxVotingInput{
		tx.TxID,
		nil,
		nil,
		tx.ElectionKeyHash,
		tx.TxOut,
		tx.Timestamp,
	}
	return txCopy
}

// Convert Election output to Byte for verification and signing purposes
func (tx *TxVotingInput) ToByte() []byte {
	var hash [32]byte

	txCopy := tx.TrimmedCopy()

	hash = sha256.Sum256([]byte(fmt.Sprintf("%x", txCopy)))
	return hash[:]
}

func (tx *TxVotingInput) IsSet() bool {
	return reflect.DeepEqual(tx, &TxVotingInput{}) == false
}

// Helper function for displaying transaction data in the console
func (tx *TxVotingInput) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--TX_INPUT: %x", tx.TxID))
	if tx.IsSet() {
		lines = append(lines, fmt.Sprintf("Timestamp: %d", tx.Timestamp))
		for i := 0; i < len(tx.Signers); i++ {
			lines = append(lines, fmt.Sprintf("(Signers) \n --(%d): %x", i, tx.Signers[i]))
		}
		for i := 0; i < len(tx.SigWitnesses); i++ {
			lines = append(lines, fmt.Sprintf("(Signature Witness): \n --(%d): %x", i, tx.SigWitnesses[i]))
		}
		lines = append(lines, fmt.Sprintf("Election Keyhash: %x", tx.ElectionKeyHash))
	}
	return strings.Join(lines, "\n")
}

// Helper function for displaying transaction data in the console
func (tx *TxVotingOutput) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--TX_OUTPUT: %x", tx.ID))
	if tx.IsSet() {
		lines = append(lines, fmt.Sprintf("Timestamp: %d", tx.Timestamp))
		for i := 0; i < len(tx.Signers); i++ {
			lines = append(lines, fmt.Sprintf("(Signers) \n --(%d): %x", i, tx.Signers[i]))
		}
		for i := 0; i < len(tx.SigWitnesses); i++ {
			lines = append(lines, fmt.Sprintf("(Signature Witness): \n --(%d): %x", i, tx.SigWitnesses[i]))
		}
		lines = append(lines, fmt.Sprintf("Election Keyhash: %x", tx.ElectionKeyHash))
	}
	return strings.Join(lines, "\n")
}

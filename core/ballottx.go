package blockchain

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"reflect"
	"strings"

	"github.com/google/uuid"
)

// CAST VOTE (BALLOT)

// Vote TxTxOutput
type TxBallotOutput struct {
	ID              string
	TxID            []byte
	Signers         [][]byte // SIGNATURE BY CONSENSUS GROUP
	SigWitnesses    [][]byte
	SecretMessage   []byte // Signed with Public view key (Decrypted with private view key) 🔑
	PubKeys         [][]byte
	ElectionKeyHash []byte
	Timestamp       int64
}

// Vote TxInput
type TxBallotInput struct {
	TxID            []byte
	Signature       []byte
	PubKeys         [][]byte
	TxOut           []byte
	Candidate       []byte
	ElectionKeyHash []byte
	Timestamp       int64
}

// NewTxBallotInput CASTS Vote using secret ballot
func NewBallotTxInput(keyHash, candidate, txId []byte, txOut []byte, signature []byte, pubKeys [][]byte, timestamp int64) *TxInput {
	tx := &TxInput{
		BallotTx: TxBallotInput{
			TxID:            txId,
			Signature:       signature,
			PubKeys:         pubKeys,
			TxOut:           txOut,
			Candidate:       candidate,
			ElectionKeyHash: keyHash,
			Timestamp:       timestamp,
		},
	}
	return tx
}

// NewTxBallotOutput generates secret Ballot
func NewBallotTxOutput(keyHash, message, txId []byte, pubKeys, signers, SigWitnesses [][]byte, timestamp int64) *TxOutput {
	tx := &TxOutput{
		BallotTx: TxBallotOutput{
			TxID:            txId,
			Signers:         signers,
			SigWitnesses:    SigWitnesses,
			PubKeys:         pubKeys,
			SecretMessage:   message,
			ElectionKeyHash: keyHash,
			Timestamp:       timestamp,
		},
	}
	uuid, _ := uuid.NewUUID()
	tx.BallotTx.ID = uuid.String()
	return tx
}

func (TxOut *TxBallotOutput) IsLockWithKey(ElectionKeyHash []byte) bool {
	return bytes.Compare(TxOut.ElectionKeyHash, ElectionKeyHash) == 0
}

func (tx *TxBallotOutput) IsSet() bool {
	return reflect.DeepEqual(tx, &TxBallotOutput{}) == false
}

// Convert Election output to Byte for verification and signing purposes
func (tx *TxBallotOutput) TrimmedCopy() TxBallotOutput {
	txCopy := TxBallotOutput{
		"",
		tx.TxID,
		nil,
		nil,
		tx.SecretMessage,
		nil,
		tx.ElectionKeyHash,
		tx.Timestamp,
	}
	return txCopy
}

// Convert Election output to Byte for verification and signing purposes
func (tx *TxBallotOutput) ToByte() []byte {
	var hash [32]byte

	txCopy := tx.TrimmedCopy()

	hash = sha256.Sum256([]byte(fmt.Sprintf("%x", txCopy)))
	return hash[:]
}

// Trim election input data
func (tx *TxBallotInput) TrimmedCopy() TxBallotInput {
	txCopy := TxBallotInput{
		tx.TxID,
		nil,
		nil,
		tx.TxOut,
		tx.Candidate,
		tx.ElectionKeyHash,
		tx.Timestamp,
	}
	return txCopy
}

// Convert Election output to Byte for verification and signing purposes
func (tx *TxBallotInput) ToByte() []byte {
	var hash [32]byte

	txCopy := tx.TrimmedCopy()

	hash = sha256.Sum256([]byte(fmt.Sprintf("%x", txCopy)))
	return hash[:]
}

func (tx *TxBallotInput) IsSet() bool {
	return reflect.DeepEqual(tx, &TxBallotInput{}) == false
}

// Helper function for displaying transaction data in the console
func (tx *TxBallotInput) String() string {
	var lines []string
	// lines = append(lines, fmt.Sprintf("TxOut: %x", tx.PubKeys))

	lines = append(lines, fmt.Sprintf("--TX_INPUT: %x", tx.TxID))
	if tx.IsSet() {
		lines = append(lines, fmt.Sprintf("Timestamp: %d", tx.Timestamp))
		lines = append(lines, fmt.Sprintf("Candidate: %x", tx.Candidate))
		lines = append(lines, fmt.Sprintf("Signature: %x", tx.Signature))
		lines = append(lines, fmt.Sprintf("(Election Keyhash): %x", tx.ElectionKeyHash))
	}
	return strings.Join(lines, "\n")
}

// Helper function for displaying transaction data in the console
func (tx *TxBallotOutput) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--TX_OUTPUT: %x", tx.ID))
	if tx.IsSet() {
		lines = append(lines, fmt.Sprintf("Timestamp: %d", tx.Timestamp))
		lines = append(lines, fmt.Sprintf("TxID: %x", tx.TxID))
		for i := 0; i < len(tx.Signers); i++ {
			lines = append(lines, fmt.Sprintf("(Signers) \n --(%d): %x", i, tx.Signers[i]))
		}
		for i := 0; i < len(tx.SigWitnesses); i++ {
			lines = append(lines, fmt.Sprintf("(Signature Witness): \n --(%d): %x", i, tx.SigWitnesses[i]))
		}
		lines = append(lines, fmt.Sprintf("(Secret Message): %x", tx.SecretMessage))
		lines = append(lines, fmt.Sprintf("(Election Keyhash): %s", tx.ElectionKeyHash))
	}
	return strings.Join(lines, "\n")
}

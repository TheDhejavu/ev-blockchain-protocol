package blockchain

import (
	"crypto/sha256"
	"fmt"
	"reflect"
	"strings"

	"github.com/google/uuid"
)

// ACCREDITATION
// Start Vote Accreditation TxTxOutput
type TxAcOutput struct {
	ID             string   `json:"id"`
	TxID           []byte   `json:"tx_id"`
	Signers        [][]byte `json:"signers"`
	SigWitnesses   [][]byte `json:"sig_witnesses"`
	ElectionPubKey []byte   `json:"election_pubkey"`
	Timestamp      int64    `json:"timestamp"`
}

// End Vote Accreditation TxInput
type TxAcInput struct {
	TxID            []byte   `json:"tx_id"`
	Signers         [][]byte `json:"signers"`
	SigWitnesses    [][]byte `json:"sig_witnesses"`
	TxOut           []byte   `json:"tx_out"`
	ElectionPubKey  []byte   `json:"election_pubkey"`
	AccreditedCount int64    `json:"accreditation_count"`
	Timestamp       int64    `json:"timestamp"`
}

// NewTxAccreditationInput Stops Accreditation  Phase
func NewAccreditationTxInput(keyHash, txId []byte, txOut []byte, signers, SigWitnesses [][]byte, count int64, timestamp int64) *TxInput {
	tx := &TxInput{
		AccreditationTx: TxAcInput{
			TxID:            txId,
			Signers:         signers,
			SigWitnesses:    SigWitnesses,
			TxOut:           txOut,
			AccreditedCount: count,
			ElectionPubKey:  keyHash,
			Timestamp:       timestamp,
		},
	}
	return tx
}

// NewTxAccreditationTxOutput Starts Accreditation Phase
func NewAccreditationTxOutput(keyHash []byte, txId []byte, signers, SigWitnesses [][]byte, timestamp int64) *TxOutput {
	tx := &TxOutput{
		AccreditationTx: TxAcOutput{
			TxID:           txId,
			Signers:        signers,
			SigWitnesses:   SigWitnesses,
			ElectionPubKey: keyHash,
			Timestamp:      timestamp,
		},
	}
	uuid, _ := uuid.NewUUID()
	tx.AccreditationTx.ID = uuid.String()
	return tx
}
func (tx *TxAcOutput) IsSet() bool {
	return reflect.DeepEqual(tx, &TxAcOutput{}) == false
}

// Convert Election output to Byte for verification and signing purposes
func (tx *TxAcOutput) TrimmedCopy() TxAcOutput {
	txCopy := TxAcOutput{
		"",
		tx.TxID,
		nil,
		nil,
		tx.ElectionPubKey,
		tx.Timestamp,
	}
	return txCopy
}

// Convert Election output to Byte for verification and signing purposes
func (tx *TxAcOutput) ToByte() []byte {
	var hash [32]byte

	txCopy := tx.TrimmedCopy()

	hash = sha256.Sum256([]byte(fmt.Sprintf("%x", txCopy)))
	return hash[:]
}

// Trim election input data
func (tx *TxAcInput) TrimmedCopy() TxAcInput {
	txCopy := TxAcInput{
		tx.TxID,
		nil,
		nil,
		tx.TxOut,
		tx.ElectionPubKey,
		tx.AccreditedCount,
		tx.Timestamp,
	}
	return txCopy
}

// Convert Election output to Byte for verification and signing purposes
func (tx *TxAcInput) ToByte() []byte {
	var hash [32]byte

	txCopy := tx.TrimmedCopy()

	hash = sha256.Sum256([]byte(fmt.Sprintf("%x", txCopy)))
	return hash[:]
}

func (tx *TxAcInput) IsSet() bool {
	return reflect.DeepEqual(tx, &TxAcInput{}) == false
}

// Helper function for displaying transaction data in the console
func (tx *TxAcInput) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--TX_INPUT: %x", tx.TxID))
	if tx.IsSet() {
		lines = append(lines, fmt.Sprintf("Accreditation Count: %d", tx.AccreditedCount))
		lines = append(lines, fmt.Sprintf("Timestamp: %d", tx.Timestamp))
		for i := 0; i < len(tx.Signers); i++ {
			lines = append(lines, fmt.Sprintf("(Signers) \n --(%d): %x", i, tx.Signers[i]))
		}
		for i := 0; i < len(tx.SigWitnesses); i++ {
			lines = append(lines, fmt.Sprintf("(Signature Witness): \n --(%d): %x", i, tx.SigWitnesses[i]))
		}
		lines = append(lines, fmt.Sprintf("TxOut: %s", tx.TxOut))
		lines = append(lines, fmt.Sprintf("Election Keyhash: %x", tx.ElectionPubKey))
	}

	return strings.Join(lines, "\n")
}

// Helper function for displaying transaction data in the console
func (tx *TxAcOutput) String() string {
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
		lines = append(lines, fmt.Sprintf("Election Keyhash: %x", tx.ElectionPubKey))
	}
	return strings.Join(lines, "\n")
}

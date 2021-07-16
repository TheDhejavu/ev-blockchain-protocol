package blockchain

import (
	"crypto/sha256"
	"fmt"
	"reflect"
	"strings"

	"github.com/google/uuid"
)

// INITIALIZE ELECTION
// Init  election TxOutput
type TxElectionOutput struct {
	ID             string
	Signers        [][]byte
	SigWitnesses   [][]byte
	ElectionPubKey []byte
	Title          string
	Description    string
	TotalPeople    int64
	Candidates     [][]byte
}

// End Election TxInput
type TxElectionInput struct {
	Signers        [][]byte
	SigWitnesses   [][]byte
	TxOut          []byte
	ElectionPubKey []byte
}

// NewTxAccreditationInput Stops Accreditation  Phase
func NewElectionTxInput(keyHash, txOut []byte, signers, SigWitnesses [][]byte) *TxInput {
	tx := &TxInput{
		ElectionTx: TxElectionInput{
			Signers:        signers,
			SigWitnesses:   SigWitnesses,
			ElectionPubKey: keyHash,
			TxOut:          txOut,
		},
	}
	return tx
}

// NewTxAccreditationTxOutput Starts Accreditation Phase
func NewElectionTxOutput(title, desp string, keyHash []byte, signers, SigWitnesses, candidates [][]byte, totalPeople int64) *TxOutput {
	tx := &TxOutput{
		ElectionTx: TxElectionOutput{
			Signers:        signers,
			SigWitnesses:   SigWitnesses,
			ElectionPubKey: keyHash,
			Title:          title,
			Description:    desp,
			TotalPeople:    totalPeople,
			Candidates:     candidates,
		},
	}
	uuid, _ := uuid.NewUUID()
	tx.ElectionTx.ID = uuid.String()
	return tx
}

// Convert Election output to Byte for verification and signing purposes
func (tx TxElectionOutput) TrimmedCopy() *TxElectionOutput {
	txCopy := &TxElectionOutput{
		"",
		nil,
		nil,
		tx.ElectionPubKey,
		tx.Title,
		tx.Description,
		tx.TotalPeople,
		tx.Candidates,
	}
	return txCopy
}

// Convert Election output to Byte for verification and signing purposes
func (tx *TxElectionOutput) ToByte() []byte {
	var hash [32]byte

	txCopy := tx.TrimmedCopy()

	hash = sha256.Sum256([]byte(fmt.Sprintf("%x", txCopy)))
	return hash[:]
}

func (tx *TxElectionOutput) IsSet() bool {
	return reflect.DeepEqual(tx, &TxElectionOutput{}) == false
}

// Trim election input data
func (tx *TxElectionInput) TrimmedCopy() *TxElectionInput {
	txCopy := &TxElectionInput{
		nil,
		nil,
		tx.TxOut,
		tx.ElectionPubKey,
	}
	return txCopy
}

// Convert Election output to Byte for verification and signing purposes
func (tx *TxElectionInput) ToByte() []byte {
	var hash [32]byte

	txCopy := tx.TrimmedCopy()

	hash = sha256.Sum256([]byte(fmt.Sprintf("%x", txCopy)))
	return hash[:]
}

func (tx *TxElectionInput) IsSet() bool {
	return reflect.DeepEqual(tx, &TxElectionInput{}) == false
}

// Helper function for displaying transaction data in the console
func (tx *TxElectionInput) String() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("--TX_INPUT: "))
	if tx.IsSet() {
		lines = append(lines, fmt.Sprintf("	Signers"))
		for i := 0; i < len(tx.Signers); i++ {
			lines = append(lines, fmt.Sprintf("		--(%d): %x", i, tx.Signers[i]))
		}
		lines = append(lines, fmt.Sprintf("	Signature Witnesses:"))
		for i := 0; i < len(tx.SigWitnesses); i++ {
			lines = append(lines, fmt.Sprintf("		--(%d): %x", i, tx.SigWitnesses[i]))
		}
		lines = append(lines, fmt.Sprintf("	Election Keyhash: %s", tx.ElectionPubKey))
	}

	return strings.Join(lines, "\n")
}

// Helper function for displaying transaction data in the console
func (tx *TxElectionOutput) String() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("--TX_OUTPUT \n"))
	if tx.IsSet() {
		lines = append(lines, fmt.Sprintf("	ID: %s", tx.ID))
		lines = append(lines, fmt.Sprintf("	Title: %s", tx.Title))
		lines = append(lines, fmt.Sprintf("	Signers"))
		for i := 0; i < len(tx.Signers); i++ {
			lines = append(lines, fmt.Sprintf("		--(%d): %x", i, tx.Signers[i]))
		}
		lines = append(lines, fmt.Sprintf("	Signature Witnesses:"))
		for i := 0; i < len(tx.SigWitnesses); i++ {
			lines = append(lines, fmt.Sprintf("		--(%d): %x", i, tx.SigWitnesses[i]))
		}
		lines = append(lines, fmt.Sprintf("	Description: %s", tx.Description))
		lines = append(lines, fmt.Sprintf("	People: %d", tx.TotalPeople))
		lines = append(lines, fmt.Sprintf("	Election Keyhash: %s", tx.ElectionPubKey))
	}
	return strings.Join(lines, "\n")
}

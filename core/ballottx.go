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
	ID             string   `json:"id"`
	TxID           []byte   `json:"tx_id"`
	Signers        [][]byte `json:"signers"` // SIGNATURE BY CONSENSUS GROUP
	SigWitnesses   [][]byte `json:"sig_witnesses"`
	SecretMessage  []byte   `json:"secret_message"` // Signed with Public view key (Decrypted with private view key) 🔑
	PubKeys        [][]byte `json:"pub_keys"`
	ElectionPubKey []byte   `json:"election_pubKey"`
	Timestamp      int64    `json:"timestamp"`
}

// Vote TxInput
type TxBallotInput struct {
	TxID           []byte   `json:"tx_id"`
	Signature      []byte   `json:"signature"`
	PubKeys        [][]byte `json:"pub_keys"`
	TxOut          []byte   `json:"tx_out"`
	Candidate      []byte   `json:"candidate"`
	ElectionPubKey []byte   `json:"election_pubkey"`
	Timestamp      int64    `json:"timestamp"`
}

// NewTxBallotInput CASTS Vote using secret ballot
func NewBallotTxInput(pubKey, candidate, txId []byte, txOut []byte, signature []byte, pubKeys [][]byte, timestamp int64) *TxInput {
	tx := &TxInput{
		BallotTx: TxBallotInput{
			TxID:           txId,
			Signature:      signature,
			PubKeys:        pubKeys,
			TxOut:          txOut,
			Candidate:      candidate,
			ElectionPubKey: pubKey,
			Timestamp:      timestamp,
		},
	}
	return tx
}

// NewTxBallotOutput generates secret Ballot
func NewBallotTxOutput(pubKey, message, txId []byte, pubKeys, signers, SigWitnesses [][]byte, timestamp int64) *TxOutput {
	tx := &TxOutput{
		BallotTx: TxBallotOutput{
			TxID:           txId,
			Signers:        signers,
			SigWitnesses:   SigWitnesses,
			PubKeys:        pubKeys,
			SecretMessage:  message,
			ElectionPubKey: pubKey,
			Timestamp:      timestamp,
		},
	}
	uuid, _ := uuid.NewUUID()
	tx.BallotTx.ID = uuid.String()
	return tx
}

func (TxOut *TxBallotOutput) IsLockWithKey(ElectionPubKey []byte) bool {
	return bytes.Compare(TxOut.ElectionPubKey, ElectionPubKey) == 0
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
		tx.ElectionPubKey,
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
		tx.ElectionPubKey,
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
		lines = append(lines, fmt.Sprintf("(Election pubKey): %x", tx.ElectionPubKey))
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
		lines = append(lines, fmt.Sprintf("(Election pubKey): %s", tx.ElectionPubKey))
	}
	return strings.Join(lines, "\n")
}

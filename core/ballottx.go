package blockchain

import (
	"bytes"
	"encoding/gob"

	logger "github.com/sirupsen/logrus"
)

// CAST VOTE (BALLOT)
// Vote TxInput
type TxBallotInput struct {
	ID              string
	Signature       []byte
	PubKeys         [][]byte
	TxOut           string
	Candidate       []byte
	ElectionKeyHash []byte
}

// Vote TxTxOutput
type TxBallotOutput struct {
	ID              string
	Signers         [][]byte // SIGNATURE BY CONSENSUS GROUP
	SigWitness      [][]byte
	TxOut           string
	SecretMessage   []byte // Signed with Public view key (Decrypted with private view key) 🔑
	PubKeys         [][]byte
	ElectionKeyHash []byte
}

type TxBallotOutputs struct {
	BallotTxOutputs []TxBallotOutput
}

// NewTxBallotInput CASTS Vote using secret ballot
func NewBallotTxInput(keyHash, signature, candidate []byte, txOut string, pubKeys [][]byte) *TxInput {
	tx := &TxInput{
		BallotTx: TxBallotInput{
			ID:              "",
			Signature:       signature,
			PubKeys:         pubKeys,
			TxOut:           txOut,
			Candidate:       candidate,
			ElectionKeyHash: keyHash,
		},
	}
	return tx
}

// NewTxBallotOutput generates secret Ballot
func NewBallotTxOutput(keyHash, id, message []byte, txOut string, pubKeys, signers, sigWitness [][]byte) *TxOutput {
	tx := &TxOutput{
		BallotTx: TxBallotOutput{
			ID:              "",
			Signers:         signers,
			SigWitness:      sigWitness,
			PubKeys:         pubKeys,
			TxOut:           txOut,
			SecretMessage:   message,
			ElectionKeyHash: keyHash,
		},
	}
	return tx
}

func (TxOutput *TxBallotOutputs) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)

	err := encoder.Encode(TxOutput)
	if err != nil {
		logger.Panic(err)
	}
	return res.Bytes()
}

func DeSerializeTxOutputs(data []byte) TxBallotOutputs {
	var TxOutputs TxBallotOutputs
	encoder := gob.NewDecoder(bytes.NewReader(data))

	err := encoder.Decode(&TxOutputs)
	if err != nil {
		logger.Panic(err)
	}
	return TxOutputs
}

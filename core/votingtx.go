package blockchain

import "bytes"

// INITIALIZE VOTE
// Start Vote TxTxOutput
type TxVotingOutput struct {
	ID              string
	Signers         [][]byte
	SigWitness      [][]byte
	ElectionKeyHash []byte
	TxOut           string
}

// End Vote TxInput
type TxVotingInput struct {
	ID              string
	Signers         [][]byte // Transaction Signatures from signers
	SigWitness      [][]byte
	ElectionKeyHash []byte
	TxOut           string
}

// NewTxVotingInput  ENDS Voting Phase
func NewVotingTxInput(keyHash []byte, txOut string, pubKeys, signers, sigWitness [][]byte) *TxInput {
	tx := &TxInput{
		VotingTx: TxVotingInput{
			ID:              "",
			Signers:         pubKeys,
			SigWitness:      sigWitness,
			TxOut:           txOut,
			ElectionKeyHash: keyHash,
		},
	}
	return tx
}

// NewTxVotingOutput BEGINS Voting Phase
func NewVotingTxOutput(keyHash []byte, txOut string, pubKeys, signers, sigWitness [][]byte) *TxOutput {
	tx := &TxOutput{
		VotingTx: TxVotingOutput{
			ID:              "",
			Signers:         pubKeys,
			SigWitness:      sigWitness,
			ElectionKeyHash: keyHash,
			TxOut:           txOut,
		},
	}
	return tx
}

func (TxOut *TxVotingOutput) IsLockWithKey(ElectionKeyHash []byte) bool {
	return bytes.Compare(TxOut.ElectionKeyHash, ElectionKeyHash) == 0
}

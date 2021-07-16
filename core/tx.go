package blockchain

import (
	"bytes"
	"encoding/gob"

	logger "github.com/sirupsen/logrus"
)

type TxInput struct {
	ElectionTx      TxElectionInput
	AccreditationTx TxAcInput
	VotingTx        TxVotingInput
	BallotTx        TxBallotInput
}

type TxOutput struct {
	ElectionTx      TxElectionOutput
	AccreditationTx TxAcOutput
	VotingTx        TxVotingOutput
	BallotTx        TxBallotOutput
}

type TxOutputs struct {
	Outputs []TxOutput
}

func (TxOutput *TxOutputs) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)

	err := encoder.Encode(TxOutput)
	if err != nil {
		logger.Panic(err)
	}
	return res.Bytes()
}

func DeSerializeOutputs(data []byte) TxOutputs {
	var TxOutputs TxOutputs
	encoder := gob.NewDecoder(bytes.NewReader(data))

	err := encoder.Decode(&TxOutputs)
	if err != nil {
		logger.Panic(err)
	}
	return TxOutputs
}

func (out *TxOutput) IsLockWithKeyHash(pubKeyHash []byte) bool {
	if out.ElectionTx.IsSet() {
		return bytes.Compare(out.ElectionTx.ElectionKeyHash, pubKeyHash) == 0
	}
	if out.VotingTx.IsSet() {
		return bytes.Compare(out.VotingTx.ElectionKeyHash, pubKeyHash) == 0
	}
	if out.BallotTx.IsSet() {
		return bytes.Compare(out.BallotTx.ElectionKeyHash, pubKeyHash) == 0
	}
	if out.AccreditationTx.IsSet() {
		return bytes.Compare(out.AccreditationTx.ElectionKeyHash, pubKeyHash) == 0
	}

	return false
}

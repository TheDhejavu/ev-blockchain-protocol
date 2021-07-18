package blockchain

import (
	"bytes"
	"encoding/gob"

	logger "github.com/sirupsen/logrus"
)

type TxInput struct {
	ElectionTx      TxElectionInput `json:"election_tx,omitempty"`
	AccreditationTx TxAcInput       `json:"accreditation_tx,omitempty"`
	VotingTx        TxVotingInput   `json:"voting_tx,omitempty"`
	BallotTx        TxBallotInput   `json:"ballot_tx,omitempty"`
}

type TxOutput struct {
	ElectionTx      TxElectionOutput `json:"election_tx,omitempty"`
	AccreditationTx TxAcOutput       `json:"accreditation_tx,omitempty"`
	VotingTx        TxVotingOutput   `json:"voting_tx,omitempty"`
	BallotTx        TxBallotOutput   `json:"ballot_tx,omitempty"`
}

type TxOutputs struct {
	Outputs []TxOutput `json:"outputs"`
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

func (out *TxOutput) IsLockWithKeyHash(pubKey []byte) bool {
	if out.ElectionTx.IsSet() {
		return bytes.Compare(out.ElectionTx.ElectionPubKey, pubKey) == 0
	}
	if out.VotingTx.IsSet() {
		return bytes.Compare(out.VotingTx.ElectionPubKey, pubKey) == 0
	}
	if out.BallotTx.IsSet() {
		return bytes.Compare(out.BallotTx.ElectionPubKey, pubKey) == 0
	}
	if out.AccreditationTx.IsSet() {
		return bytes.Compare(out.AccreditationTx.ElectionPubKey, pubKey) == 0
	}

	return false
}

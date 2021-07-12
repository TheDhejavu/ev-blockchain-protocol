package blockchain

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

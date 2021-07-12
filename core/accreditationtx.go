package blockchain

// ACCREDITATION
// Start Vote Accreditation TxTxOutput
type TxAcOutput struct {
	ID              string
	Signers         [][]byte
	SigWitness      [][]byte
	ElectionKeyHash []byte
	TxOut           string
}

// End Vote Accreditation TxInput
type TxAcInput struct {
	ID              string
	Signers         [][]byte
	SigWitness      [][]byte
	TxOut           string
	ElectionKeyHash []byte
	AccreditedCount int64
}

// NewTxAccreditationInput Stops Accreditation  Phase
func NewAccreditationTxInput(keyHash []byte, txOut string, pubKeys, signers, sigWitness [][]byte, count int64) *TxInput {
	tx := &TxInput{
		AccreditationTx: TxAcInput{
			ID:              "",
			Signers:         pubKeys,
			SigWitness:      sigWitness,
			TxOut:           txOut,
			AccreditedCount: count,
		},
	}
	return tx
}

// NewTxAccreditationTxOutput Starts Accreditation Phase
func NewAccreditationTxOutput(keyHash []byte, txOut string, pubKeys, signers, sigWitness [][]byte) *TxOutput {
	tx := &TxOutput{
		AccreditationTx: TxAcOutput{
			ID:              "",
			Signers:         pubKeys,
			SigWitness:      sigWitness,
			ElectionKeyHash: keyHash,
			TxOut:           txOut,
		},
	}

	return tx
}

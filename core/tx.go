package core

// EvTransaction represent the Transaction entity of the blockchain block
type EvTransaction struct {
	ID     []byte
	Input  []TxVoteInput
	Output []TxVoteOutput
}

// CAST VOTE (BALLOT)

// Vote TxInput
type TxVoteInput struct {
	ID        []byte
	Signature []byte
	PubKeys   [][]byte
	Out       int
	Candidate string
	Action    string
}

// Vote TxOutput
type TxVoteOutput struct {
	ID            []byte
	ProofSigs     [][]byte // SIGNATURE BY CONSENSUS GROUP
	ProofPubKeys  [][]byte
	Out           int
	SecretMessage []byte // Signed with Public view key (Decrypted with private view key) 🔑
	PubKeys       [][]byte
	Action        string
}

// ACCREDITATION

// Start Vote Accreditation TxOutput
type TxAccreditationOutput struct {
	ID           []byte
	ProofSigs    [][]byte
	ProofPubKeys []Pubkey
	Name         string
	Desp         string
	Action       string // "BEGIN_VOTE_ACCREDITATION"
}

// End Vote Accreditation TxInput
type TxAccreditationInput struct {
	ID           []byte
	ProofSigs    [][]byte
	ProofPubKeys []Pubkey
	Out          int
	Action       string // "END_VOTE_ACCREDITATION"
}

// INITIALIZE VOTE

// Start Vote TxOutput
type TxVoteInitOutput struct {
	ID           []byte
	ProofSig     []byte
	ProofPubKeys [][]byte
	Out          int
	Action       string //BEGIN_VOTE
}

// End Vote TxInput
type TxVoteInitInput struct {
	ID           []byte
	ProofSigs    [][]byte
	ProofPubKeys []Pubkey
	Out          int
	Action       string //"END_VOTE"
}

type Pubkey []byte

// We need to make sure that both the EvTxOtput and EvTxInput are valid and not changed
// using ring digital signature this can be used to verify accredited ballot
// and the casted ballot during and after the election without compromising the user identity
// and ballot anonymity

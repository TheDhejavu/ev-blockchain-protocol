package rpc

import (
	"context"
	"encoding/json"
	"fmt"

	jrpc "github.com/gumeniukcom/golang-jsonrpc2"
	logger "github.com/sirupsen/logrus"
	blockchain "github.com/workspace/evoting/ev-blockchain-protocol/core"
)

type Handler struct {
	Blockchain *blockchain.Blockchain
	Serve      *jrpc.JSONRPC
}

type HandlerEntity interface {
	// Query all election results by election pubkey
	QueryResults(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error)

	// Query all un used Ballot transactions
	QueryUnUsedBallotTxs(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error)

	// Query the blockchain data
	QueryBlockchain(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error)

	// Query all block transactions
	QueryTransactions(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error)

	// Query all transactions by pub key
	QueryTransactionsByPubkey(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error)

	// Get transaction by ID
	GetTransaction(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error)

	// Start election  by creating new TxOutput
	StartElectionTx(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error)

	// Stop election  by creating new TxInput
	StopElectionTx(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error)

	// Start accreditation by creating new TxOutput
	StartAccreditationTx(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error)

	// Stop accreditation by creating new TxInput
	StopAccreditationTx(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error)

	// Start voting by creating new TxOutput
	StartVotingTx(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error)

	// Stop voting by creating new TxInput
	StopVotingTx(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error)

	// Create ballot transaction by creating new TxOutput
	CreateBallotTx(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error)

	// Cast Ballot by creating new TxInput
	CastBallotTx(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error)

	//Finf transaction with transaction Output by public key
	FindTransactionWithTxOutput(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error)
}

func NewHandler(bc *blockchain.Blockchain, serve *jrpc.JSONRPC) HandlerEntity {
	bc = bc.ReInit()
	handler := &Handler{bc, serve}
	registerHandlers(handler)
	return handler
}

func registerHandlers(h *Handler) {
	if err := h.Serve.RegisterMethod("QueryResults", h.QueryResults); err != nil {
		logger.Panic(err)
	}
	if err := h.Serve.RegisterMethod("QueryUnUsedBallotTxs", h.QueryUnUsedBallotTxs); err != nil {
		logger.Panic(err)
	}
	if err := h.Serve.RegisterMethod("QueryBlockchain", h.QueryBlockchain); err != nil {
		logger.Panic(err)
	}
	if err := h.Serve.RegisterMethod("QueryTransactions", h.QueryTransactions); err != nil {
		logger.Panic(err)
	}
	if err := h.Serve.RegisterMethod("QueryTransactionsByPubkey", h.QueryTransactionsByPubkey); err != nil {
		logger.Panic(err)
	}
	if err := h.Serve.RegisterMethod("GetTransaction", h.GetTransaction); err != nil {
		logger.Panic(err)
	}

	if err := h.Serve.RegisterMethod("FindTxWithTxOutput", h.FindTransactionWithTxOutput); err != nil {
		logger.Panic(err)
	}

	if err := h.Serve.RegisterMethod("StartElection", h.StartElectionTx); err != nil {
		logger.Panic(err)
	}
	if err := h.Serve.RegisterMethod("StopElection", h.StopElectionTx); err != nil {
		logger.Panic(err)
	}
	if err := h.Serve.RegisterMethod("StartAccreditation", h.StartAccreditationTx); err != nil {
		logger.Panic(err)
	}
	if err := h.Serve.RegisterMethod("StopAccreditation", h.StopAccreditationTx); err != nil {
		logger.Panic(err)
	}
	if err := h.Serve.RegisterMethod("StartVoting", h.StartVotingTx); err != nil {
		logger.Panic(err)
	}
	if err := h.Serve.RegisterMethod("StopVoting", h.StopVotingTx); err != nil {
		logger.Panic(err)
	}
	if err := h.Serve.RegisterMethod("CreateBallot", h.CreateBallotTx); err != nil {
		logger.Panic(err)
	}
	if err := h.Serve.RegisterMethod("CastBallot", h.CastBallotTx); err != nil {
		logger.Panic(err)
	}
}

type QueryResultsRequest struct {
	PubKey []byte `json:"pubkey"`
}

type QueryResultsResponse struct {
	Data map[string]int `json:"data"`
}

func (h *Handler) QueryResults(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	if data == nil {
		return nil, jrpc.InvalidRequestErrorCode, fmt.Errorf("Empty request")
	}
	request := &QueryResultsRequest{}
	err := json.Unmarshal(data, request)
	if err != nil {
		logger.Error("UnMarshal Error: ", err)
		return nil, jrpc.InvalidRequestErrorCode, err
	}

	results, err := h.Blockchain.QueryResult(request.PubKey)
	if err != nil {
		logger.Error(err)
		return nil, jrpc.InvalidRequestErrorCode, err
	}
	response := QueryResultsResponse{
		Data: results,
	}
	mdata, err := json.Marshal(response)
	if err != nil {
		logger.Error("Marshal Error: ", err)
		return nil, jrpc.InternalErrorCode, err
	}

	return mdata, jrpc.OK, nil
}

type QueryUnUsedBallotTxsRequest struct {
	PubKey []byte `json:"pubkey"`
}

type QueryUnUsedBallotTxsResponse struct {
	Data []map[string]blockchain.TxBallotOutput `json:"data"`
}

func (h *Handler) QueryUnUsedBallotTxs(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	if data == nil {
		return nil, jrpc.InvalidRequestErrorCode, fmt.Errorf("Empty request")
	}
	request := &QueryUnUsedBallotTxsRequest{}
	err := json.Unmarshal(data, request)
	if err != nil {
		logger.Error("UnMarshal Error: ", err)
		return nil, jrpc.InvalidRequestErrorCode, err
	}

	results, err := h.Blockchain.GetUnUsedBallotTxOutputs(request.PubKey)
	if err != nil {
		logger.Error(err)
		return nil, jrpc.InvalidRequestErrorCode, err
	}
	response := QueryUnUsedBallotTxsResponse{
		Data: results,
	}

	mdata, err := json.Marshal(response)
	if err != nil {
		logger.Error("Marshal Error: ", err)
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}

type QueryBlockchainResponse struct {
	Data []blockchain.Block `json:"data"`
}

func (h *Handler) QueryBlockchain(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	if data == nil {
		return nil, jrpc.InvalidRequestErrorCode, fmt.Errorf("Empty request")
	}

	results, err := h.Blockchain.GetBlockchain()

	if err != nil {
		logger.Error(err)
		return nil, jrpc.InvalidRequestErrorCode, err
	}
	response := QueryBlockchainResponse{
		Data: results,
	}
	mdata, err := json.Marshal(response)
	if err != nil {
		logger.Error("Marshal Error: ", err)
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}

type QueryTransactionsResponse struct {
	Data []*blockchain.Transaction `json:"data"`
}

func (h *Handler) QueryTransactions(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	if data == nil {
		return nil, jrpc.InvalidRequestErrorCode, fmt.Errorf("Empty request")
	}

	results, err := h.Blockchain.GetTransactions()
	if err != nil {
		logger.Error("Results Error:", err)
		return nil, jrpc.InvalidRequestErrorCode, err
	}
	response := QueryTransactionsResponse{
		Data: results,
	}
	mdata, err := json.Marshal(response)
	if err != nil {
		logger.Error("Marshal Error: ", err)
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}

type GetTransactionRequest struct {
	ID []byte `json:"id"`
}

type GetTransactionResponse struct {
	Data blockchain.Transaction `json:"data"`
}

func (h *Handler) GetTransaction(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	if data == nil {
		return nil, jrpc.InvalidRequestErrorCode, fmt.Errorf("Empty request")
	}
	request := &GetTransactionRequest{}
	err := json.Unmarshal(data, request)
	if err != nil {
		logger.Error("UnMarshal Error: ", err)
		return nil, jrpc.InvalidRequestErrorCode, err
	}

	if err != nil {
		logger.Error("Decode Error:", err)
		return nil, jrpc.InvalidRequestErrorCode, err
	}
	results, err := h.Blockchain.GetTransaction(request.ID)
	if err != nil {
		logger.Error("Results Error:", err)
		return nil, jrpc.InvalidRequestErrorCode, err
	}
	response := GetTransactionResponse{
		Data: results,
	}
	mdata, err := json.Marshal(response)
	if err != nil {
		logger.Error("Marshal Error: ", err)
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}

type QueryTransactionsByPubkeyRequest struct {
	PubKey string `json:"pubkey"`
}

type QueryTransactionsByPubkeyResponse struct {
	Data []blockchain.Transaction `json:"data"`
}

func (h *Handler) QueryTransactionsByPubkey(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	if data == nil {
		return nil, jrpc.InvalidRequestErrorCode, fmt.Errorf("Empty request")
	}
	request := &QueryTransactionsByPubkeyRequest{}
	err := json.Unmarshal(data, request)
	if err != nil {
		logger.Error("UnMarshal Error: ", err)
		return nil, jrpc.InvalidRequestErrorCode, err
	}

	results, err := h.Blockchain.GetTransactionsByPubkey([]byte(request.PubKey))
	if err != nil {
		logger.Error("Results Error:", err)
		return nil, jrpc.InvalidRequestErrorCode, err
	}

	response := QueryTransactionsByPubkeyResponse{
		Data: results,
	}
	mdata, err := json.Marshal(response)
	if err != nil {
		logger.Error("Marshal Error: ", err)
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}

type FindTxWithTxOutputRequest struct {
	PubKey []byte `json:"pubkey"`
	Type   string `json:"type"`
}

type FindTxWithTxOutputResponse struct {
	Data blockchain.Transaction
}

func (h *Handler) FindTransactionWithTxOutput(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	if data == nil {
		return nil, jrpc.InvalidRequestErrorCode, fmt.Errorf("Empty request")
	}
	request := &FindTxWithTxOutputRequest{}
	err := json.Unmarshal(data, request)
	if err != nil {
		logger.Error("UnMarshal Error: ", err)
		return nil, jrpc.InvalidRequestErrorCode, err
	}

	var results blockchain.Transaction
	switch request.Type {
	case blockchain.ELECTION_TX_TYPE:
		results, err = h.Blockchain.FindTxWithElectionOutByPubkey(request.PubKey)
	case blockchain.ACCREDITATION_TX_TYPE:
		results, err = h.Blockchain.FindTxWithAcOutByPubkey(request.PubKey)
	case blockchain.VOTING_TX_TYPE:
		results, err = h.Blockchain.FindTxWithVotingOutByPubkey(request.PubKey)
	}

	if err != nil {
		logger.Error("Results Error:", err)
		return nil, jrpc.InvalidRequestErrorCode, err
	}

	response := FindTxWithTxOutputResponse{
		Data: results,
	}
	mdata, err := json.Marshal(response)
	if err != nil {
		logger.Error("Marshal Error: ", err)
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}

type TxResponse struct {
	Data ResponseData `json:"data"`
}

type ResponseData struct {
	TxID []byte `json:"tx_id"`
}

type StartElectionRequest struct {
	Data blockchain.TxElectionOutput `json:"data"`
}

// Start election  by creating new TxOutput
func (h *Handler) StartElectionTx(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	var eTx *blockchain.Transaction

	if data == nil {
		return nil, jrpc.InvalidRequestErrorCode, fmt.Errorf("Empty request")
	}
	request := &StartElectionRequest{}

	err := json.Unmarshal(data, &request)
	if err != nil {
		logger.Error("UnMarshal Error: ", err)
		return nil, jrpc.InvalidRequestErrorCode, err
	}
	fmt.Println("PUBKKKK", request.Data.ElectionPubKey)
	txOut := blockchain.NewElectionTxOutput(
		request.Data.Title,
		request.Data.Description,
		request.Data.ElectionPubKey,
		request.Data.Signers,
		request.Data.SigWitnesses,
		request.Data.Candidates,
		request.Data.TotalPeople,
	)

	eTx, err = blockchain.NewTransaction(
		blockchain.ELECTION_TX_TYPE,
		request.Data.ElectionPubKey,
		blockchain.TxInput{},
		*txOut,
	)
	if err != nil {
		logger.Error("Block Error:", err)
		return nil, jrpc.InternalErrorCode, err
	}

	_, err = h.Blockchain.AddBlock([]*blockchain.Transaction{eTx})
	if err != nil {
		logger.Error("Block Error:", err)
		return nil, jrpc.InternalErrorCode, err
	}

	response := TxResponse{
		Data: ResponseData{
			TxID: eTx.ID,
		},
	}
	mdata, err := json.Marshal(response)
	if err != nil {
		logger.Error("Marshal Error: ", err)
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}

type StopElectionRequest struct {
	Pubkey []byte                     `json:"pubkey"`
	Data   blockchain.TxElectionInput `json:"data"`
}

// Stop election  by creating new TxInput
func (h *Handler) StopElectionTx(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	var electionTx *blockchain.Transaction

	if data == nil {
		return nil, jrpc.InvalidRequestErrorCode, fmt.Errorf("Empty request")
	}
	request := &StopElectionRequest{}

	err := json.Unmarshal(data, &request)
	if err != nil {
		logger.Error("UnMarshal Error: ", err)
		return nil, jrpc.InvalidRequestErrorCode, err
	}

	txIn := blockchain.NewElectionTxInput(
		request.Pubkey,
		request.Data.TxOut,
		request.Data.Signers,
		request.Data.SigWitnesses,
	)

	electionTx, err = blockchain.NewTransaction(
		blockchain.ELECTION_TX_TYPE,
		request.Pubkey,
		*txIn,
		blockchain.TxOutput{},
	)
	if err != nil {
		logger.Error("Block Error:", err)
		return nil, jrpc.InternalErrorCode, err
	}

	_, err = h.Blockchain.AddBlock([]*blockchain.Transaction{electionTx})
	if err != nil {
		logger.Error("Block Error:", err)
		return nil, jrpc.InternalErrorCode, err
	}

	response := TxResponse{
		Data: ResponseData{
			TxID: electionTx.ID,
		},
	}
	mdata, err := json.Marshal(response)
	if err != nil {
		logger.Error("Marshal Error: ", err)
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}

type StartAccreditationRequest struct {
	Pubkey []byte                `json:"pubkey"`
	Data   blockchain.TxAcOutput `json:"data"`
}

// Start accreditation by creating new TxOutput
func (h *Handler) StartAccreditationTx(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	var eaTx *blockchain.Transaction

	if data == nil {
		return nil, jrpc.InvalidRequestErrorCode, fmt.Errorf("Empty request")
	}
	request := &StartAccreditationRequest{}

	err := json.Unmarshal(data, &request)
	if err != nil {
		logger.Error("UnMarshal Error: ", err)
		return nil, jrpc.InternalErrorCode, err
	}

	txAccreditationOut := blockchain.NewAccreditationTxOutput(
		request.Pubkey,
		request.Data.TxID,
		request.Data.Signers,
		request.Data.SigWitnesses,
		request.Data.Timestamp,
	)

	eaTx, _ = blockchain.NewTransaction(
		blockchain.ACCREDITATION_TX_TYPE,
		request.Pubkey,
		blockchain.TxInput{},
		*txAccreditationOut,
	)
	block, err := h.Blockchain.AddBlock([]*blockchain.Transaction{eaTx})
	if err != nil {
		logger.Error("Block Error:", err)
		return nil, jrpc.InternalErrorCode, err
	}

	fmt.Println("Block added  sucessfully: \n", block)

	response := TxResponse{
		Data: ResponseData{
			TxID: eaTx.ID,
		},
	}
	mdata, err := json.Marshal(response)
	if err != nil {
		logger.Error("Marshal Error: ", err)
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}

type StopAccreditationRequest struct {
	Pubkey []byte               `json:"pubkey"`
	Data   blockchain.TxAcInput `json:"data"`
}

// Stop accreditation by creating new TxInput
func (h *Handler) StopAccreditationTx(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	var acTx *blockchain.Transaction

	if data == nil {
		return nil, jrpc.InvalidRequestErrorCode, fmt.Errorf("Empty request")
	}
	request := &StopAccreditationRequest{}

	err := json.Unmarshal(data, &request)
	if err != nil {
		logger.Error("UnMarshal Error: ", err)
		return nil, jrpc.InternalErrorCode, err
	}

	txAcIn := blockchain.NewAccreditationTxInput(
		request.Pubkey,
		request.Data.TxID,
		request.Data.TxOut,
		request.Data.Signers,
		request.Data.SigWitnesses,
		request.Data.AccreditedCount,
		request.Data.Timestamp,
	)

	acTx, _ = blockchain.NewTransaction(
		blockchain.ACCREDITATION_TX_TYPE,
		request.Pubkey,
		*txAcIn,
		blockchain.TxOutput{},
	)
	block, err := h.Blockchain.AddBlock([]*blockchain.Transaction{acTx})
	if err != nil {
		logger.Error("Block Error:", err)
		return nil, jrpc.InternalErrorCode, err
	}

	fmt.Println("Block added  sucessfully: \n", block)

	response := TxResponse{
		Data: ResponseData{
			TxID: acTx.ID,
		},
	}
	mdata, err := json.Marshal(response)
	if err != nil {
		logger.Error("Marshal Error: ", err)
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}

type StartVotingRequest struct {
	Pubkey []byte                    `json:"pubkey"`
	Data   blockchain.TxVotingOutput `json:"data"`
}

// Start voting by creating new TxOutput
func (h *Handler) StartVotingTx(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	var vTx *blockchain.Transaction

	if data == nil {
		return nil, jrpc.InternalErrorCode, fmt.Errorf("Empty request")
	}
	request := &StartVotingRequest{}

	err := json.Unmarshal(data, &request)
	if err != nil {
		logger.Error("UnMarshal Error: ", err)
		return nil, jrpc.InternalErrorCode, err
	}

	votingOut := blockchain.NewVotingTxOutput(
		request.Pubkey,
		request.Data.TxID,
		request.Data.Signers,
		request.Data.SigWitnesses,
		request.Data.Timestamp,
	)

	vTx, _ = blockchain.NewTransaction(
		blockchain.VOTING_TX_TYPE,
		request.Pubkey,
		blockchain.TxInput{},
		*votingOut,
	)
	block, err := h.Blockchain.AddBlock([]*blockchain.Transaction{vTx})
	if err != nil {
		logger.Error("Block Error:", err)
		return nil, jrpc.InternalErrorCode, err
	}

	fmt.Println("Block added  sucessfully: \n", block)

	response := TxResponse{
		Data: ResponseData{
			TxID: vTx.ID,
		},
	}

	mdata, err := json.Marshal(response)
	if err != nil {
		logger.Error("Marshal Error: ", err)
		return nil, jrpc.InternalErrorCode, err
	}
	return mdata, jrpc.OK, nil
}

type StopVotingRequest struct {
	Pubkey []byte                   `json:"pubkey"`
	Data   blockchain.TxVotingInput `json:"data"`
}

// Stop voting by creating new TxInput
func (h *Handler) StopVotingTx(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	var vTx *blockchain.Transaction

	if data == nil {
		return nil, jrpc.InvalidRequestErrorCode, fmt.Errorf("Empty request")
	}
	request := &StopVotingRequest{}

	err := json.Unmarshal(data, &request)
	if err != nil {
		logger.Error("UnMarshal Error: ", err)
		return nil, jrpc.InternalErrorCode, err
	}

	txVotingIn := blockchain.NewVotingTxInput(
		request.Pubkey,
		request.Data.TxID,
		request.Data.TxOut,
		request.Data.Signers,
		request.Data.SigWitnesses,
		request.Data.Timestamp,
	)

	vTx, _ = blockchain.NewTransaction(
		blockchain.VOTING_TX_TYPE,
		request.Pubkey,
		*txVotingIn,
		blockchain.TxOutput{},
	)
	block, err := h.Blockchain.AddBlock([]*blockchain.Transaction{vTx})
	if err != nil {
		logger.Error("Block Error:", err)
		return nil, jrpc.InternalErrorCode, err
	}

	fmt.Println("Block added  sucessfully: \n", block)

	response := TxResponse{
		Data: ResponseData{
			TxID: vTx.ID,
		},
	}
	mdata, err := json.Marshal(response)
	if err != nil {
		logger.Error("Marshal Error: ", err)
		return nil, jrpc.InternalErrorCode, err
	}

	return mdata, jrpc.OK, nil
}

type CreateBallotRequest struct {
	Pubkey []byte                    `json:"pubkey"`
	Data   blockchain.TxBallotOutput `json:"data"`
}

// Create ballot transaction by creating new TxOutput
func (h *Handler) CreateBallotTx(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {
	var bTx *blockchain.Transaction
	if data == nil {
		return nil, jrpc.InvalidRequestErrorCode, fmt.Errorf("Empty request")
	}
	request := &CreateBallotRequest{}

	err := json.Unmarshal(data, &request)
	if err != nil {
		logger.Error("UnMarshal Error: ", err)
		return nil, jrpc.InvalidRequestErrorCode, err
	}

	bTxOut := blockchain.NewBallotTxOutput(
		request.Pubkey,
		request.Data.SecretMessage,
		request.Data.TxID,
		request.Data.PubKeys,
		request.Data.Signers,
		request.Data.SigWitnesses,
		request.Data.Timestamp,
	)

	bTx, _ = blockchain.NewTransaction(
		blockchain.BALLOT_TX_TYPE,
		request.Pubkey,
		blockchain.TxInput{},
		*bTxOut,
	)
	block, err := h.Blockchain.AddBlock([]*blockchain.Transaction{bTx})
	if err != nil {
		logger.Error("Block Error:", err)
		return nil, jrpc.InternalErrorCode, err
	}

	fmt.Println("Block added  sucessfully: \n", block)

	response := TxResponse{
		Data: ResponseData{
			TxID: bTx.ID,
		},
	}
	mdata, err := json.Marshal(response)
	if err != nil {
		logger.Error("Marshal Error: ", err)
		return nil, jrpc.InternalErrorCode, err
	}

	return mdata, jrpc.OK, nil
}

type CastBallotRequest struct {
	Pubkey []byte                   `json:"pubkey"`
	Data   blockchain.TxBallotInput `json:"data"`
}

// Cast Ballot by creating new TxInput
func (h *Handler) CastBallotTx(ctx context.Context, data json.RawMessage) (json.RawMessage, int, error) {

	var bTx *blockchain.Transaction
	if data == nil {
		return nil, jrpc.InvalidRequestErrorCode, fmt.Errorf("Empty request")
	}
	request := &CastBallotRequest{}

	err := json.Unmarshal(data, &request)
	if err != nil {
		logger.Error("UnMarshal Error: ", err)
		return nil, jrpc.InvalidRequestErrorCode, err
	}

	bTxIn := blockchain.NewBallotTxInput(
		request.Pubkey,
		request.Data.Candidate,
		request.Data.TxID,
		request.Data.TxOut,
		request.Data.Signature,
		request.Data.PubKeys,
		request.Data.Timestamp,
	)

	bTx, _ = blockchain.NewTransaction(
		blockchain.BALLOT_TX_TYPE,
		request.Pubkey,
		*bTxIn,
		blockchain.TxOutput{},
	)
	block, err := h.Blockchain.AddBlock([]*blockchain.Transaction{bTx})
	if err != nil {
		logger.Error("Block Error:", err)
		return nil, jrpc.InternalErrorCode, err
	}

	fmt.Println("Block added  sucessfully: \n", block)

	response := TxResponse{
		Data: ResponseData{
			TxID: bTx.ID,
		},
	}
	mdata, err := json.Marshal(response)
	if err != nil {
		logger.Error("Marshal Error: ", err)
		return nil, jrpc.InternalErrorCode, err
	}

	return mdata, jrpc.OK, nil
}

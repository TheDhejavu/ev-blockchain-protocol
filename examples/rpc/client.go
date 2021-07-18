package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	logger "github.com/sirupsen/logrus"
	blockchain "github.com/workspace/evoting/ev-blockchain-protocol/core"
	"github.com/workspace/evoting/ev-blockchain-protocol/pkg/crypto/multisig"
	"github.com/workspace/evoting/ev-blockchain-protocol/pkg/crypto/ringsig"
	"github.com/workspace/evoting/ev-blockchain-protocol/rpc"
	"github.com/workspace/evoting/ev-blockchain-protocol/wallet"
)

const numOfKeys = 3

var (
	DefaultCurve = elliptic.P256()
	keyring      *ringsig.PublicKeyRing
	signers      [][]byte
	privKeys     []*ecdsa.PrivateKey
	sigCount     = 4
)

type BlockchainRepo struct {
	client *rpc.Client
}

func NewBlockchainRepository(client *rpc.Client) *BlockchainRepo {
	return &BlockchainRepo{client}
}
func (repo *BlockchainRepo) FindTxWithTxOutput(pubkey, ttype string) blockchain.Transaction {
	data := map[string]string{
		"pubkey": pubkey,
		"type":   ttype,
	}

	resp, err := repo.client.Do("FindTxWithTxOutput", data)
	if err != nil {
		logger.Error("UnMarshal Error: ", err)
	}

	var tx blockchain.Transaction
	inrec, err := json.Marshal(resp.Body.Result.Data)
	if err != nil {
		logger.Error("Marshal Error: ", err)
	}

	json.Unmarshal(inrec, &tx)

	return tx
}

func (repo *BlockchainRepo) QueryResults(pubkey string) {
	data := map[string]string{
		"pubkey": pubkey,
	}

	resp, err := repo.client.Do("QueryResults", data)
	if err != nil {
		logger.Error("UnMarshal Error: ", err)
	}
	fmt.Println(resp.Body)
}
func (repo *BlockchainRepo) QueryBlockchain() {
	data := map[string]string{}

	resp, err := repo.client.Do("QueryBlockchain", data)
	if err != nil {
		logger.Error("UnMarshal Error: ", err)
	}
	fmt.Println(resp.Body)
}

func (repo *BlockchainRepo) QueryUnUsedBallotTxs(pubkey string) []map[string]blockchain.TxBallotOutput {
	data := map[string]string{
		"pubkey": pubkey,
	}

	resp, err := repo.client.Do("QueryUnUsedBallotTxs", data)
	if err != nil {
		logger.Error("UnMarshal Error: ", err)
	}
	// fmt.Println(resp.Body)

	var utxbo []map[string]blockchain.TxBallotOutput
	inrec, err := json.Marshal(resp.Body.Result.Data)
	if err != nil {
		logger.Error("Marshal Error: ", err)
	}

	json.Unmarshal(inrec, &utxbo)

	return utxbo
}

func (repo *BlockchainRepo) GetTransaction(id string) {

	data := map[string]string{
		"id": id,
	}

	resp, err := repo.client.Do("GetTransaction", data)
	if err != nil {
		logger.Error("UnMarshal Error: ", err)
	}
	fmt.Println(resp.Body)
}

type ElectionOutput struct {
	Data blockchain.TxElectionOutput `json:"data"`
}

func (repo *BlockchainRepo) StartElection(title string, pubkey string, description string, totalPeople int64, candidates [][]byte) {
	electionPubkey, _ := base64.StdEncoding.DecodeString(pubkey)

	txOut := blockchain.TxElectionOutput{
		ID:             "",
		SigWitnesses:   nil,
		Signers:        nil,
		ElectionPubKey: electionPubkey,
		Title:          title,
		Description:    description,
		TotalPeople:    totalPeople,
		Candidates:     candidates,
	}

	mu := multisig.NewMultisig(sigCount)
	for i := 0; i < sigCount; i++ {
		// Initialize system identity wallet
		wallets, _ := wallet.InitializeWallets()
		// Add new identity to the wallet with the User ID
		userId := wallets.AddWallet(fmt.Sprintf("signers_%d", i))
		wallets.Save()
		w, err := wallets.GetWallet(userId)
		if err != nil {
			logger.Panic(err)
		}
		mu.AddSignature(
			txOut.ToByte(),
			w.Main.PublicKey,
			w.Main.PrivateKey,
		)

		privKeys = append(privKeys, &w.Main.PrivateKey)
	}

	txOut.SigWitnesses = mu.Sigs
	txOut.Signers = mu.PubKeys

	Output := ElectionOutput{txOut}
	var outInterface map[string]interface{}
	inrec, err := json.Marshal(Output)
	if err != nil {
		logger.Error("Marshal Error: ", err)
	}
	json.Unmarshal(inrec, &outInterface)

	resp, err := repo.client.Do("StartElection", outInterface)
	if err != nil {
		logger.Error("UnMarshal Error: ", err)
	}
	fmt.Println(resp.Body)
}

type ElectionInput struct {
	Pubkey []byte                     `json:"pubkey"`
	Data   blockchain.TxElectionInput `json:"data"`
}

func (repo *BlockchainRepo) StopElection(pubkey string) {
	electionPubkey, _ := base64.StdEncoding.DecodeString(pubkey)
	txElectionOut := repo.FindTxWithTxOutput(pubkey, "election_tx")

	txIn := blockchain.TxElectionInput{
		SigWitnesses:   nil,
		Signers:        nil,
		TxOut:          txElectionOut.ID,
		ElectionPubKey: electionPubkey,
	}

	mu := multisig.NewMultisig(sigCount)
	for i := 0; i < sigCount; i++ {
		// Initialize system identity wallet
		wallets, _ := wallet.InitializeWallets()
		userId := fmt.Sprintf("signers_%d", i)
		w, err := wallets.GetWallet(userId)
		if err != nil {
			logger.Panic(err)
		}
		mu.AddSignature(
			txIn.ToByte(),
			w.Main.PublicKey,
			w.Main.PrivateKey,
		)
		privKeys = append(privKeys, &w.Main.PrivateKey)
	}

	txIn.SigWitnesses = mu.Sigs
	txIn.Signers = mu.PubKeys

	Output := ElectionInput{txElectionOut.ElectionPubkey, txIn}

	var outInterface map[string]interface{}
	inrec, err := json.Marshal(Output)
	if err != nil {
		logger.Error("Marshal Error: ", err)
	}
	json.Unmarshal(inrec, &outInterface)

	resp, err := repo.client.Do("StopElection", outInterface)
	if err != nil {
		logger.Error("UnMarshal Error: ", err)
	}
	fmt.Println(resp.Body)
}

type AccreditationOutput struct {
	Pubkey []byte                `json:"pubkey"`
	Data   blockchain.TxAcOutput `json:"data"`
}

func (repo *BlockchainRepo) StartAccreditation(pubkey, txElectionOutId string) {
	electionPubkey, _ := base64.StdEncoding.DecodeString(pubkey)
	txId, _ := base64.StdEncoding.DecodeString(txElectionOutId)

	txOut := blockchain.TxAcOutput{
		SigWitnesses:   nil,
		Signers:        nil,
		TxID:           txId,
		ElectionPubKey: electionPubkey,
		Timestamp:      time.Now().Unix(),
	}

	mu := multisig.NewMultisig(sigCount)
	for i := 0; i < sigCount; i++ {
		// Initialize system identity wallet
		wallets, _ := wallet.InitializeWallets()
		userId := fmt.Sprintf("signers_%d", i)
		w, err := wallets.GetWallet(userId)
		if err != nil {
			logger.Panic(err)
		}
		mu.AddSignature(
			txOut.ToByte(),
			w.Main.PublicKey,
			w.Main.PrivateKey,
		)
		privKeys = append(privKeys, &w.Main.PrivateKey)
	}

	txOut.SigWitnesses = mu.Sigs
	txOut.Signers = mu.PubKeys

	Output := AccreditationOutput{electionPubkey, txOut}
	var outInterface map[string]interface{}
	inrec, err := json.Marshal(Output)
	if err != nil {
		logger.Error("Marshal Error: ", err)
	}
	json.Unmarshal(inrec, &outInterface)

	resp, err := repo.client.Do("StartAccreditation", outInterface)
	if err != nil {
		logger.Error("UnMarshal Error: ", err)
	}
	fmt.Println(resp.Body)
}

type AccreditationInput struct {
	Pubkey []byte               `json:"pubkey"`
	Data   blockchain.TxAcInput `json:"data"`
}

func (repo *BlockchainRepo) StopAccreditation(pubkey, txElectionOutId string, txAcOutId string) {
	txId, _ := base64.StdEncoding.DecodeString(txElectionOutId)
	txOut, _ := base64.StdEncoding.DecodeString(txAcOutId)
	electionPubkey, _ := base64.StdEncoding.DecodeString(pubkey)

	txIn := blockchain.TxAcInput{
		SigWitnesses:    nil,
		Signers:         nil,
		TxID:            txId,
		TxOut:           txOut,
		ElectionPubKey:  electionPubkey,
		Timestamp:       time.Now().Unix(),
		AccreditedCount: 100,
	}

	mu := multisig.NewMultisig(sigCount)
	for i := 0; i < sigCount; i++ {
		// Initialize system identity wallet
		wallets, _ := wallet.InitializeWallets()
		userId := fmt.Sprintf("signers_%d", i)
		w, err := wallets.GetWallet(userId)
		if err != nil {
			logger.Panic(err)
		}
		mu.AddSignature(
			txIn.ToByte(),
			w.Main.PublicKey,
			w.Main.PrivateKey,
		)
		privKeys = append(privKeys, &w.Main.PrivateKey)
	}

	txIn.SigWitnesses = mu.Sigs
	txIn.Signers = mu.PubKeys

	Output := AccreditationInput{electionPubkey, txIn}
	var outInterface map[string]interface{}
	inrec, err := json.Marshal(Output)
	if err != nil {
		logger.Error("Marshal Error: ", err)
	}
	json.Unmarshal(inrec, &outInterface)

	// fmt.Println(/outInterface)
	resp, err := repo.client.Do("StopAccreditation", outInterface)
	if err != nil {
		logger.Error("UnMarshal Error: ", err)
	}
	fmt.Println(resp.Body)
}

type VotingOutput struct {
	Pubkey []byte                    `json:"pubkey"`
	Data   blockchain.TxVotingOutput `json:"data"`
}

func (repo *BlockchainRepo) StartVoting(pubkey string, txElectionOutId string) {
	txId, _ := base64.StdEncoding.DecodeString(txElectionOutId)
	electionPubkey, _ := base64.StdEncoding.DecodeString(pubkey)

	txOut := blockchain.TxVotingOutput{
		SigWitnesses:   nil,
		Signers:        nil,
		TxID:           txId,
		ElectionPubKey: electionPubkey,
		Timestamp:      time.Now().Unix(),
	}

	mu := multisig.NewMultisig(sigCount)
	for i := 0; i < sigCount; i++ {
		// Initialize system identity wallet
		wallets, _ := wallet.InitializeWallets()
		userId := fmt.Sprintf("signers_%d", i)
		w, err := wallets.GetWallet(userId)
		if err != nil {
			logger.Panic(err)
		}
		mu.AddSignature(
			txOut.ToByte(),
			w.Main.PublicKey,
			w.Main.PrivateKey,
		)
		privKeys = append(privKeys, &w.Main.PrivateKey)
	}

	txOut.SigWitnesses = mu.Sigs
	txOut.Signers = mu.PubKeys

	Output := VotingOutput{electionPubkey, txOut}
	var outInterface map[string]interface{}
	inrec, err := json.Marshal(Output)
	if err != nil {
		logger.Error("Marshal Error: ", err)
	}
	json.Unmarshal(inrec, &outInterface)

	fmt.Println(outInterface)
	resp, err := repo.client.Do("StartVoting", outInterface)
	if err != nil {
		logger.Error("UnMarshal Error: ", err)
	}
	fmt.Println(resp.Body)
}

type VotingInput struct {
	Pubkey []byte                   `json:"pubkey"`
	Data   blockchain.TxVotingInput `json:"data"`
}

func (repo *BlockchainRepo) StopVoting(pubkey, txElectionOutId, txVotingOutId string) {
	txId, _ := base64.StdEncoding.DecodeString(txElectionOutId)
	txOut, _ := base64.StdEncoding.DecodeString(txVotingOutId)
	electionPubkey, _ := base64.StdEncoding.DecodeString(pubkey)

	txIn := blockchain.TxVotingInput{
		SigWitnesses:   nil,
		Signers:        nil,
		TxID:           txId,
		TxOut:          txOut,
		ElectionPubKey: electionPubkey,
		Timestamp:      time.Now().Unix(),
	}

	mu := multisig.NewMultisig(sigCount)
	for i := 0; i < sigCount; i++ {
		// Initialize system identity wallet
		wallets, _ := wallet.InitializeWallets()
		userId := fmt.Sprintf("signers_%d", i)
		w, err := wallets.GetWallet(userId)
		if err != nil {
			logger.Panic(err)
		}
		mu.AddSignature(
			txIn.ToByte(),
			w.Main.PublicKey,
			w.Main.PrivateKey,
		)
		privKeys = append(privKeys, &w.Main.PrivateKey)
	}

	txIn.SigWitnesses = mu.Sigs
	txIn.Signers = mu.PubKeys

	Output := VotingInput{electionPubkey, txIn}
	var outInterface map[string]interface{}
	inrec, err := json.Marshal(Output)
	if err != nil {
		logger.Error("Marshal Error: ", err)
	}
	json.Unmarshal(inrec, &outInterface)

	// fmt.Println(/outInterface)
	resp, err := repo.client.Do("StopVoting", outInterface)
	if err != nil {
		logger.Error("UnMarshal Error: ", err)
	}
	fmt.Println(resp.Body)
}

type BallotOutput struct {
	Pubkey []byte                    `json:"pubkey"`
	Data   blockchain.TxBallotOutput `json:"data"`
}

func (repo *BlockchainRepo) CreateBallot(userId, pubkey, txElectionOutId string) {
	var keyRingByte [][]byte

	txId, _ := base64.StdEncoding.DecodeString(txElectionOutId)
	electionPubkey, _ := base64.StdEncoding.DecodeString(pubkey)
	secretMessage := []byte("This is my ballot secret message")
	wallets, _ := wallet.InitializeWallets()
	userWallet, _ := wallets.GetWallet(userId)
	msg, _ := userWallet.View.Encrypt(secretMessage)

	keyRingByte = append(keyRingByte, userWallet.Main.PublicKey)
	// Generate Decoy keys
	for i := 0; i < numOfKeys-1; i++ {

		// Initialize system identity wallet
		wallets, _ := wallet.InitializeWallets()
		userId := wallets.AddWallet(fmt.Sprintf("decoy_%d", i))
		wallets.Save()
		w, _ := wallets.GetWallet(userId)
		// add the public key part to the ring
		keyRingByte = append(keyRingByte, w.Main.PublicKey)

	}

	txOut := blockchain.TxBallotOutput{
		SigWitnesses:   nil,
		Signers:        nil,
		TxID:           txId,
		ElectionPubKey: electionPubkey,
		PubKeys:        keyRingByte,
		SecretMessage:  msg,
		Timestamp:      time.Now().Unix(),
	}

	mu := multisig.NewMultisig(sigCount)
	for i := 0; i < sigCount; i++ {
		// Initialize system identity wallet
		wallets, _ := wallet.InitializeWallets()
		userId := fmt.Sprintf("signers_%d", i)
		w, err := wallets.GetWallet(userId)
		if err != nil {
			logger.Panic(err)
		}
		mu.AddSignature(
			txOut.ToByte(),
			w.Main.PublicKey,
			w.Main.PrivateKey,
		)
		privKeys = append(privKeys, &w.Main.PrivateKey)
	}

	txOut.SigWitnesses = mu.Sigs
	txOut.Signers = mu.PubKeys

	Output := BallotOutput{electionPubkey, txOut}
	var outInterface map[string]interface{}
	inrec, err := json.Marshal(Output)
	if err != nil {
		logger.Error("Marshal Error: ", err)
	}
	json.Unmarshal(inrec, &outInterface)

	resp, err := repo.client.Do("CreateBallot", outInterface)
	if err != nil {
		logger.Error("UnMarshal Error: ", err)
	}
	fmt.Println(resp.Body)
}

type BallotInput struct {
	Pubkey []byte                   `json:"pubkey"`
	Data   blockchain.TxBallotInput `json:"data"`
}

func (repo *BlockchainRepo) CastBallot(userId, pubkey, txElectionOutId, candidatePubkey string) {
	electionPubkey, _ := base64.StdEncoding.DecodeString(pubkey)
	candidate, _ := base64.StdEncoding.DecodeString(candidatePubkey)
	txId, _ := base64.StdEncoding.DecodeString(txElectionOutId)
	wallets, _ := wallet.InitializeWallets()
	// Get user wallet from UserId
	userWallet, _ := wallets.GetWallet(userId)

	var txOut []byte
	var keyRingByte [][]byte

	//Find Ballot
	utxbo := repo.QueryUnUsedBallotTxs(pubkey)
	for _, v := range utxbo {
		for txId, value := range v {
			if userWallet.View.CanDecrypt(value.SecretMessage) {
				txOut, _ = hex.DecodeString(txId)
				keyRingByte = value.PubKeys
				break
			}
		}
	}
	// fmt.Println(keyRingByte)
	// Convert keys from byte to ecdsa.Publickey and add to the ring.
	keyring = ringsig.NewPublicKeyRing(numOfKeys)
	keyring.Add(userWallet.Main.PrivateKey.PublicKey)
	// Generate Decoy keys
	for i := 0; i < numOfKeys-1; i++ {
		// Initialize system identity wallet
		userId := fmt.Sprintf("decoy_%d", i)
		w, _ := wallets.GetWallet(userId)
		// add the public key part to the ring
		keyring.Add(w.Main.PrivateKey.PublicKey)
	}
	// for i := 0; i < len(keyRingByte); i++ {
	// 	pub := keyRingByte[0]
	// 	x := big.Int{}
	// 	y := big.Int{}
	// 	keyLen := len(pub)
	// 	x.SetBytes(pub[:(keyLen / 2)])
	// 	y.SetBytes(pub[(keyLen / 2):])

	// 	rawPubKey := ecdsa.PublicKey{Curve: DefaultCurve, X: &x, Y: &y}
	// 	keyring.Add(rawPubKey)
	// }

	txIn := blockchain.TxBallotInput{
		Signature:      nil,
		PubKeys:        nil,
		TxID:           txId,
		TxOut:          txOut,
		Candidate:      candidate,
		ElectionPubKey: electionPubkey,
		Timestamp:      time.Now().Unix(),
	}

	// Sign message
	signature, err := ringsig.Sign(
		&userWallet.Main.PrivateKey,
		keyring,
		txIn.ToByte(),
	)

	txIn.Signature = signature.ToByte()
	txIn.PubKeys = keyRingByte

	Output := BallotInput{electionPubkey, txIn}
	var outInterface map[string]interface{}
	inrec, err := json.Marshal(Output)
	if err != nil {
		logger.Error("Marshal Error: ", err)
	}
	json.Unmarshal(inrec, &outInterface)

	resp, err := repo.client.Do("CastBallot", outInterface)
	if err != nil {
		logger.Error("UnMarshal Error: ", err)
	}
	fmt.Println(resp.Body)
}

func Candidates() [][]byte {
	var candidates [][]byte

	for i := 0; i < 4; i++ {
		wallets, _ := wallet.InitializeWallets()
		// Add new identity to the wallet with the User ID
		userId := wallets.AddWallet(fmt.Sprintf("candidates_%d", i))
		wallets.Save()
		w, _ := wallets.GetWallet(userId)
		candidates = append(candidates, w.Main.PublicKey)
	}

	return candidates
}
func GetCandidates() [][]byte {
	var candidates [][]byte

	for i := 0; i < 4; i++ {
		wallets, _ := wallet.InitializeWallets()
		userId := fmt.Sprintf("candidates_%d", i)
		w, _ := wallets.GetWallet(userId)
		candidates = append(candidates, w.Main.PublicKey)
	}

	return candidates
}

var (
	_txElectionOutId = "XmJpKlJpuLWSKIgaM1CoNAYsjNTR6FwvHQ888FlMGRE="
	_txAcOutId       = "rd26pDYNqLu7JhUttdR70P+mH3PqEfpAXn1zgPsoBzU="
	_txVotingOutId   = "MGTFKE/oanyX5UtYyff/XZ9EOSnjH5sqSbntFB2aQvg="
)

func main() {
	electionPubkey := []byte("1_election_12345678")
	pubkeyStr := base64.StdEncoding.EncodeToString(electionPubkey)
	userId := fmt.Sprintf("candidates_1")

	client := rpc.NewClient("http://localhost:8088/json-rpc")
	chainRepo := NewBlockchainRepository(client)

	// chainRepo.GetTransaction()
	// chainRepo.StartElection(
	// 	"Presidential election",
	// 	"president  speeeechs",
	// 	pubkeyStr,
	// 	100,
	// 	Candidates(),
	// )
	// chainRepo.StopElection()
	// chainRepo.StartAccreditation(
	// 	pubkeyStr,
	// 	_txElectionOutId,
	// )
	// chainRepo.StopAccreditation(
	// 	pubkeyStr,
	// 	_txElectionOutId,
	// 	_txAcOutId,
	// )
	// chainRepo.StartVoting(
	// 	pubkeyStr,
	// 	_txElectionOutId,
	// )
	// chainRepo.StopVoting()
	// chainRepo.CreateBallot(
	// 	userId,
	// 	pubkeyStr,
	// 	_txElectionOutId,
	// )
	candidates := GetCandidates()
	candidate := base64.StdEncoding.EncodeToString(candidates[0])
	chainRepo.CastBallot(userId, pubkeyStr, _txElectionOutId, candidate)
	// chainRepo.QueryUnUsedBallotTxs()
	// chainRepo.QueryBlockchain()
	// chainRepo.QueryResults()

}

// m7jxErFXq/fy4mP0X/SxclaEDQStJryKRuqBVmI/4zo=

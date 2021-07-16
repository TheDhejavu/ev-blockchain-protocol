package blockchain

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"sync"

	logger "github.com/sirupsen/logrus"
	"github.com/workspace/evoting/ev-blockchain-protocol/database"
	"github.com/workspace/evoting/ev-blockchain-protocol/pkg/config"
)

// Blockchain struct such that lastHash represents the lastblock hash
// on the ledger
type Blockchain struct {
	config   config.Config
	lashHash []byte
	crud     *Crud
}

var (
	mutex = &sync.Mutex{}
)

func NewBlockchain(s database.Store, cfg config.Config) *Blockchain {
	return &Blockchain{
		config: cfg,
		crud:   NewCrud(s),
	}
}
func (bc *Blockchain) ResetBlockchain(name string) error {
	return database.RemoveDatabase(name)
}
func (bc *Blockchain) Init() *Blockchain {
	logger.Info("Initializing blockchain")
	_, err := bc.crud.GetLastHash()

	if err != nil {
		logger.Info("Create genesis block")
		genesis := Genesis(&Transaction{}, Version)
		// add genesis block to blockchain
		err := bc.crud.Save(genesis.Hash, genesis.Serialize())
		if err != nil {
			logger.Panic(err)
		}
		//save genesis hash as lasthash
		err = bc.crud.Save(lastHashKey, genesis.Hash)
		if err != nil {
			logger.Panic(err)
		}
		bc.lashHash = genesis.Hash
	} else {
		logger.Error("Blockchain exist already with a genesis block")
	}
	return bc
}
func (bc *Blockchain) ReInit() *Blockchain {
	logger.Info("Re-Initializing blockchain")
	lastHash, err := bc.crud.GetLastHash()

	if err != nil {
		logger.Error("You cannot Re-initialize a blockchain that has not been initialized before!!")
	} else {
		logger.Info("Get last blockchain hash")
		bc.lashHash = lastHash
	}
	// Compute
	// bc.ComputeUnUsedTXOs()

	return bc
}

func (bc *Blockchain) AddBlock(transactions []*Transaction) (*Block, error) {
	mutex.Lock()

	for _, tx := range transactions {
		if bc.VerifyTx(tx) != true {
			log.Panic("Invalid Transaction")
		}
	}
	// get block from lasthash
	lastBlock, err := bc.crud.GetBlock(bc.lashHash)
	if err != nil {
		return &Block{}, err
	}
	// New block
	block := NewBlock(
		transactions,
		Version,
		bc.lashHash,
		lastBlock.Height+1,
	)
	// Store block
	block, err = bc.crud.StoreBlock(block)
	if err != nil {
		return block, err
	}
	err = bc.crud.Save(lastHashKey, block.Hash)
	if err != nil {
		return &Block{}, err
	}
	bc.lashHash = block.Hash

	mutex.Unlock()
	return block, nil
}

// Get Block from the blockchain
func (bc *Blockchain) GetBlock(hash []byte) (Block, error) {
	var block Block
	block, err := bc.crud.GetBlock(hash)

	if err != nil {
		return block, err
	}

	return block, nil
}

func (bc *Blockchain) GetBlockByHeight(height int) (Block, error) {
	var block *Block

	iter, err := bc.crud.Iterator()
	if err != nil {
		return *block, err
	}

	for {
		block = iter.Next()
		prevHash := block.PrevHash
		if block.Height == height+1 {
			break
		}
		if prevHash == nil {
			break
		}
	}

	return *block, nil
}

// Get Block from the blockchain
func (bc *Blockchain) GetBlockHashes(height int) ([][]byte, error) {
	data, err := bc.crud.GetBlockHashes(height)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Get Best height basically gets the height(Index) of the lastBlock
func (bc *Blockchain) GetBestHeight() int {
	var lastBlock Block
	lastBlockData, err := bc.crud.GetLastHash()
	if err != nil {
		return 0
	}
	lastBlock = *DeSerialize(lastBlockData)

	return lastBlock.Height
}

func (bc *Blockchain) GetTransaction(txId []byte) Transaction {
	tx, err := bc.crud.FindTransaction(txId)
	if err != nil {
		log.Panic("Error: Invalid Transaction Ewwww")
	}

	return tx
}

func (bc *Blockchain) GetTransactions() (txs []*Transaction, err error) {
	iter, err := bc.crud.Iterator()
	if err != nil {
		return
	}
	for {
		block := iter.Next()
		txs = append(txs, block.Transactions...)
		if len(block.PrevHash) == 0 {
			break
		}
	}

	return
}

func (bc *Blockchain) GetTransactionByPubkey(pubKey []byte) (Transaction, error) {
	tx, err := bc.crud.FindTxByElectionPubkey(pubKey)
	if err != nil {
		logger.Error("Error: Transaction with pubkey doesn't exist")
		return tx, err
	}

	return tx, nil
}

func (bc *Blockchain) GetElectionTxByPubkey(pubKey []byte) (tx Transaction, err error) {
	iter, err := bc.crud.Iterator()
	if err != nil {
		return
	}
	for {
		block := iter.Next()
		for i := 0; i < len(block.Transactions); i++ {
			tx = *block.Transactions[i]
			txElection := tx.Output.ElectionTx
			if bytes.Compare(txElection.ElectionPubKey, pubKey) == 0 {
				return
			}
		}
		if len(block.PrevHash) == 0 {
			break
		}
	}

	return
}

func (bc *Blockchain) GetAcTxByPubkey(pubKey []byte) (tx Transaction, err error) {
	iter, err := bc.crud.Iterator()
	if err != nil {
		return
	}
	for {
		block := iter.Next()
		for i := 0; i < len(block.Transactions); i++ {
			tx = *block.Transactions[i]
			txAc := tx.Output.AccreditationTx
			if bytes.Compare(txAc.ElectionPubKey, pubKey) == 0 {
				return
			}
		}
		if len(block.PrevHash) == 0 {
			break
		}
	}
	return
}

func (bc *Blockchain) GetVotingTxByPubkey(pubKey []byte) (tx Transaction, err error) {
	iter, err := bc.crud.Iterator()
	if err != nil {
		return
	}
	for {
		block := iter.Next()
		for i := 0; i < len(block.Transactions); i++ {
			tx = *block.Transactions[i]
			txVtx := tx.Output.VotingTx
			if bytes.Compare(txVtx.ElectionPubKey, pubKey) == 0 {
				return
			}
		}
		if len(block.PrevHash) == 0 {
			break
		}
	}

	return
}

func (bc *Blockchain) GetBallotTxByPubkey(pubKey []byte) (tx Transaction, err error) {
	iter, err := bc.crud.Iterator()
	if err != nil {
		return
	}
	for {
		block := iter.Next()
		for i := 0; i < len(block.Transactions); i++ {
			tx = *block.Transactions[i]
			txBtx := tx.Output.BallotTx
			if bytes.Compare(txBtx.ElectionPubKey, pubKey) == 0 {
				return
			}
		}
		if len(block.PrevHash) == 0 {
			break
		}
	}

	return
}

func (bc *Blockchain) GetUnUsedBallotTxOutputs(pubKey []byte) (tx []TxBallotOutput, err error) {
	utxos := NewUnusedXTOSet(bc)
	// Compute UTXoSet
	utxos.Compute()
	unUsedTxos := utxos.FindUnUsedBallotTxOuputs(pubKey)
	for _, v := range unUsedTxos {
		tx = append(tx, v.BallotTx)
	}
	return
}

func (bc *Blockchain) QueryResult(pubKey []byte) (map[string]int, error) {
	var candidate string
	var results = make(map[string]int)
	txElection, _ := bc.GetElectionTxByPubkey(pubKey)
	for _, v := range txElection.Output.ElectionTx.Candidates {
		candidate := hex.EncodeToString(v)
		results[candidate] = 0
	}

	iter, err := bc.crud.Iterator()
	if err != nil {
		return results, err
	}
	for {
		block := iter.Next()
		for i := 0; i < len(block.Transactions); i++ {
			tx := *block.Transactions[i]
			txBallotIn := tx.Input.BallotTx
			fmt.Println(txBallotIn)
			if bytes.Compare(txBallotIn.ElectionPubKey, pubKey) == 0 {
				candidate = hex.EncodeToString(txBallotIn.Candidate)
				if _, ok := results[candidate]; ok {
					results[candidate] += 1
				}
			}
		}
		if len(block.PrevHash) == 0 {
			break
		}
	}
	return results, nil
}

func (bc *Blockchain) GetPrevTransactionByInput(transaction *Transaction) (Transaction, error) {
	var txId []byte
	var tx Transaction

	switch transaction.Type {
	case ELECTION_TX_TYPE:
		txId = transaction.Input.ElectionTx.TxOut
	case ACCREDITATION_TX_TYPE:
		txId = transaction.Input.AccreditationTx.TxOut
	case VOTING_TX_TYPE:
		txId = transaction.Input.VotingTx.TxOut
	case BALLOT_TX_TYPE:
		txId = transaction.Input.BallotTx.TxOut
	}

	tx, err := bc.crud.FindTransaction(txId)
	if err != nil {
		logger.Error("Error: Transaction with ID does not exist")
		return tx, err
	}

	return tx, nil
}

func (bc *Blockchain) GetPrevTransactionByOutput(transaction *Transaction) (Transaction, error) {
	var txId []byte
	var tx Transaction

	switch transaction.Type {
	case ACCREDITATION_TX_TYPE:
		txId = transaction.Output.AccreditationTx.TxID
	case VOTING_TX_TYPE:
		txId = transaction.Output.VotingTx.TxID
	case BALLOT_TX_TYPE:
		txId = transaction.Output.BallotTx.TxID
	}

	tx, err := bc.crud.FindTransaction(txId)
	if err != nil {
		logger.Info("Error: Transaction does not exist")
		return tx, err
	}

	return tx, nil
}
func (bc *Blockchain) VerifyTx(tx *Transaction) bool {
	var prevTx Transaction
	var err error

	if tx.inputSet() {
		prevTx, err = bc.GetPrevTransactionByInput(tx)
	}

	if tx.outputSet() {
		prevTx, err = bc.GetPrevTransactionByOutput(tx)
	}

	if err != nil {
		logger.Error("Verification error occurred", err)
		return false
	}

	return tx.Verify(prevTx)
}

// Aggregate all Unused Transaction output from the blockchain
func (bc *Blockchain) FindUnUsedTXO() (map[string]TxOutputs, error) {
	UTXOs := make(map[string]TxOutputs)
	usedTXOs := make(map[string]string)

	iter, err := bc.crud.Iterator()
	if err != nil {
		logger.Error("Iteration failed with error:", err)
		return UTXOs, nil
	}
	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

			if tx.outputSet() {
				if _, ok := usedTXOs[txID]; ok {
					continue
				}

				outs := UTXOs[txID]
				outs.Outputs = append(outs.Outputs, tx.Output)
				UTXOs[txID] = outs
			}

			var txInID []byte
			var valueIn []byte

			if tx.inputSet() {

				switch tx.Type {
				case ELECTION_TX_TYPE:
					txInID = tx.Input.ElectionTx.TxOut
					valueIn = tx.Input.ElectionTx.ElectionPubKey
				case ACCREDITATION_TX_TYPE:
					txInID = tx.Input.AccreditationTx.TxOut
					valueIn = tx.Input.AccreditationTx.ElectionPubKey
				case VOTING_TX_TYPE:
					txInID = tx.Input.VotingTx.TxOut
					valueIn = tx.Input.VotingTx.ElectionPubKey
				case BALLOT_TX_TYPE:
					txInID = tx.Input.BallotTx.TxOut
					valueIn = tx.Input.BallotTx.ElectionPubKey
				}
			}

			if txInID != nil {
				id := hex.EncodeToString(txInID)
				usedTXOs[id] = fmt.Sprintf("%s", valueIn)
			}

		}
		if len(block.PrevHash) == 0 {
			break
		}
	}

	return UTXOs, nil
}

func (bc *Blockchain) ComputeUnUsedTXOs() {
	unusedXTOSet := UnusedXTOSet{bc}
	unusedXTOSet.Compute()
}

func (bc *Blockchain) PrintBlockchain() {
	var oldBlock Block
	iter, err := bc.crud.Iterator()

	if err != nil {
		logger.Panic("An error occurred", err)
	}
	for {
		block := iter.Next()
		fmt.Printf("PrevHash: %x\n", block.PrevHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("Height: %d\n", block.Height)
		if block.PrevHash != nil {
			oldBlock, _ = bc.GetBlock(block.PrevHash)
			validate := block.IsBlockValid(oldBlock)
			fmt.Printf("Valid: %s\n", strconv.FormatBool(validate))
		}

		for _, tx := range block.Transactions {
			fmt.Println(tx)
			if block.IsGenesis() == false {
				fmt.Printf("Transaction Valid: %t\n", bc.VerifyTx(tx))
			}
		}
		fmt.Println()

		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (bc *Blockchain) GetBlockchain() ([]*Block, error) {
	var blocks []*Block

	iter, err := bc.crud.Iterator()
	if err != nil {
		return blocks, err
	}
	for {
		block := iter.Next()
		blocks = append(blocks, block)

		if len(block.PrevHash) == 0 {
			break
		}
	}
	return blocks, nil
}

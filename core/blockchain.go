package blockchain

import (
	"encoding/hex"
	"fmt"
	"log"
	"reflect"
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

func (bc *Blockchain) Init() *Blockchain {
	logger.Info("Initializing blockchain")
	_, err := bc.crud.GetLastHash()

	if err != nil {
		logger.Info("Create genesis block")
		genesis := Genesis(&Transaction{}, Version)
		err := bc.crud.Save(lastHashKey, genesis.Hash)
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
	return bc
}
func (bc *Blockchain) AddBlock(block *Block) (*Block, error) {
	mutex.Lock()
	block, err := bc.crud.StoreBlock(block)
	if err != nil {
		return block, err
	}
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

func (bc *Blockchain) GetTransactionByKeyHash(keyHash []byte) Transaction {
	tx, err := bc.crud.FindTransactionByKeyHash(keyHash)
	if err != nil {
		log.Panic("Error: Invalid Transaction Ewwww")
	}

	return tx
}

func (bc *Blockchain) GetPrevTransactionByInput(transaction *Transaction) (Transaction, error) {
	var txId []byte
	var tx Transaction

	switch transaction.Type {
	case ELECTION_TX_TYPE:
		txId = transaction.Input.ElectionTx.TxID
	case ACCREDITATION_TX_TYPE:
		txId = transaction.Input.AccreditationTx.TxID
	case VOTING_TX_TYPE:
		txId = transaction.Input.VotingTx.TxID
	case BALLOT_TX_TYPE:
		txId = transaction.Input.BallotTx.TxID
	}

	tx, err := bc.crud.FindTransaction(txId)
	if err != nil {
		logger.Error("Error: Invalid Transaction Ewwww")
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
		logger.Error("Error: Invalid Transaction Ewwww")
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

	if reflect.DeepEqual(prevTx, Transaction{}) {
		return false
	}

	return tx.Verify(prevTx)
}

// Aggregate all Unused Transaction output from the blockchain
func (bc *Blockchain) FindUnUsedTXO() (map[string]TxOutputs, error) {
	UTXOs := make(map[string]TxOutputs)
	usedTXOs := make(map[string][]byte)

	iter, err := bc.crud.Iterator()
	if err != nil {
		logger.Error("Iteration failed with error:", err)
		return UTXOs, nil
	}
	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)
			var txOutID string

			switch tx.Type {
			case ELECTION_TX_TYPE:
				txOutID = tx.Output.ElectionTx.ID
			case ACCREDITATION_TX_TYPE:
				txOutID = tx.Output.AccreditationTx.ID
			case VOTING_TX_TYPE:
				txOutID = tx.Output.VotingTx.ID
			case BALLOT_TX_TYPE:
				txOutID = tx.Output.VotingTx.ID
			}
			if txOutID != "" {
				if _, ok := usedTXOs[txOutID]; ok {
					continue
				}

				outs := UTXOs[txID]
				outs.Outputs = append(outs.Outputs, tx.Output)
				UTXOs[txID] = outs
			}

			// Add
			var txInID string
			var valueIn []byte
			switch tx.Type {
			case ELECTION_TX_TYPE:
				txInID = tx.Input.ElectionTx.TxOut
				valueIn = tx.Input.ElectionTx.ElectionKeyHash
			case ACCREDITATION_TX_TYPE:
				txInID = tx.Input.AccreditationTx.TxOut
				valueIn = tx.Input.AccreditationTx.ElectionKeyHash
			case VOTING_TX_TYPE:
				txInID = tx.Input.VotingTx.TxOut
				valueIn = tx.Input.VotingTx.ElectionKeyHash
			case BALLOT_TX_TYPE:
				txInID = tx.Input.VotingTx.TxOut
				valueIn = tx.Input.VotingTx.ElectionKeyHash
			}

			if txInID != "" {
				usedTXOs[txInID] = valueIn
			}
		}
		if len(block.PrevHash) == 0 {
			break
		}
	}
	return UTXOs, nil
}

func (bc *Blockchain) ComputeUsedTXOs(keyHash string) {}

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

		validate := block.IsBlockValid(oldBlock)
		fmt.Printf("Valid: %s\n", strconv.FormatBool(validate))
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		fmt.Println()

		if len(block.PrevHash) == 0 {
			break
		}
		oldBlock = *block
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

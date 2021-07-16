package blockchain

import (
	"bytes"
	"errors"
	"log"

	logger "github.com/sirupsen/logrus"
	"github.com/workspace/evoting/ev-blockchain-protocol/database"
)

var (
	ErrAlreadyExists      = errors.New("transaction already exists")
	ErrInvalidTransaction = errors.New("No transaction with id")
)

type (
	Crud struct {
		ps database.Store
	}

	Iterator struct {
		ps          database.Store
		currentHash []byte
	}
)

var (
	lastHashKey = []byte("lh")
)

func NewCrud(store database.Store) *Crud {
	return &Crud{ps: store}
}

func (crud *Crud) GetLastHash() ([]byte, error) {
	return crud.ps.Get(lastHashKey)
}

func (crud *Crud) Save(key, value []byte) error {
	return crud.ps.Put(key, value)
}

func (crud *Crud) StoreBlock(block *Block) (*Block, error) {
	_, err := crud.GetBlock(block.Hash)
	if err == nil {
		return block, err
	}

	blockData := block.Serialize()
	err = crud.Save(block.Hash, blockData)
	if err == nil {
		return block, err
	}

	err = crud.Save(lastHashKey, block.Hash)
	if err == nil {
		logger.Error("Error: Unable to update lastblock Hash")
		return block, err
	}
	return block, nil
}
func (crud *Crud) GetBlock(key []byte) (Block, error) {
	blockData, err := crud.ps.Get(key)
	if err != nil {
		return Block{}, err
	}
	block := *DeSerialize(blockData)
	return block, err
}

//Aggregate and get all block hashes in the blockchain
func (crud *Crud) GetBlockHashes(height int) ([][]byte, error) {
	var blockHashes [][]byte

	iter, err := crud.Iterator()
	if err != nil {
		return blockHashes, err
	}

	for {
		block := iter.Next()
		prevHash := block.PrevHash
		if block.Height == height {
			break
		}
		blockHashes = append([][]byte{block.Hash}, blockHashes...)

		if prevHash == nil {
			break
		}
	}

	return blockHashes, nil
}

func (crud *Crud) DeleteByPrefix(prefix []byte) {
	deleteKeys := func(keysForDelete [][]byte) error {
		for _, key := range keysForDelete {
			if err := crud.ps.Delete(key); err != nil {
				return err
			}
		}
		return nil
	}
	// This is the maximum number of items that badgerDB can delete at once, so we
	// have to aggregate all keys with utxo prefix and delete it in batch
	collectSize := 100000
	keysForDelete := make([][]byte, 0, collectSize)
	keysCollected := 0
	crud.ps.Seek(prefix, func(key, v []byte) {
		keysForDelete = append(keysForDelete, key)
		keysCollected++
		if keysCollected == collectSize {
			if err := deleteKeys(keysForDelete); err != nil {
				log.Panic(err)
			}
			// Reset keys to delete collection size
			keysForDelete = make([][]byte, 0, collectSize)
			keysCollected = 0
		}
	})
	if keysCollected > 0 {
		if err := deleteKeys(keysForDelete); err != nil {
			log.Panic(err)
		}
	}
}

func (crud *Crud) CountByPrefix(prefix []byte) int {
	counter := 0
	crud.ps.Seek(prefix, func(key, v []byte) {
		counter++
	})
	return counter
}

func (crud *Crud) FindTransaction(ID []byte) (Transaction, error) {
	iter, err := crud.Iterator()
	if err != nil {
		return Transaction{}, err
	}

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}
		if len(block.PrevHash) == 0 {
			break
		}
	}
	logger.Error("Error: No Transaction with ID")

	return Transaction{}, ErrInvalidTransaction
}

func (crud *Crud) FindTransactionByKeyHash(keyHash []byte) (Transaction, error) {
	iter, err := crud.Iterator()
	if err != nil {
		return Transaction{}, err
	}
	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.KeyHash, keyHash) == 0 {
				return *tx, nil
			}
		}
		if len(block.PrevHash) == 0 {
			break
		}
	}
	logger.Error("Error: No Transaction with ID")

	return Transaction{}, ErrInvalidTransaction
}

func (crud *Crud) Iterator() (*Iterator, error) {
	lastHash, err := crud.GetLastHash()
	if err != nil {
		return nil, err
	}

	return &Iterator{crud.ps, lastHash}, nil
}

func (iter *Iterator) Next() *Block {
	var block *Block

	blockData, err := iter.ps.Get(iter.currentHash)
	if err != nil {
		logger.Panic(err)
	}
	block = DeSerialize(blockData)

	iter.currentHash = block.PrevHash
	return block
}

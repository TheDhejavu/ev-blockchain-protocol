package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"time"

	logger "github.com/sirupsen/logrus"
)

// Block represent the Block entity of the blockchain
type Block struct {
	Timestamp    int64          `json:"timestamp"`
	Version      int            `json:"version"`
	Hash         []byte         `json:"hash"`
	PrevHash     []byte         `json:"prev_hash"`
	Transactions []*Transaction `json:"transactions"`
	Height       int            `json:"height"`
	MerkleRoot   []byte         `json:"merkle_root"`
	TxCount      int            `json:"tx_count"`
}

var (
	Version = 1
)

func NewBlock(txs []*Transaction, version int, prevHash []byte, height int) *Block {
	block := &Block{
		time.Now().Unix(),
		version,
		[]byte{},
		prevHash,
		txs,
		height,
		[]byte{},
		len(txs),
	}
	block.MerkleRoot = block.HashTransactions()
	block.Hash = block.GetHashData()

	return block
}

// GetHashData  returns the hash of the block
func (block *Block) GetHashData() []byte {
	info := bytes.Join(
		[][]byte{
			block.MerkleRoot,
			block.PrevHash,
		}, []byte{})

	hash := sha256.Sum256(info)
	return hash[:]
}

// HashTransactions Uses Merkle Tree to hash the Transactions
func (block *Block) HashTransactions() []byte {
	var txHashes [][]byte

	for _, tx := range block.Transactions {
		txHashes = append(txHashes, tx.Serialize())
	}

	tree := NewMerkleTree(txHashes)
	return tree.RootNode.Data
}

//Serialize function for serializing blockchain data
func (b *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)

	err := encoder.Encode(b)
	if err != nil {
		logger.Panic(err)
	}
	return res.Bytes()
}

// DeSerialize function for De-serializing blockchain data
func DeSerialize(data []byte) *Block {
	var block Block
	encoder := gob.NewDecoder(bytes.NewReader(data))

	err := encoder.Decode(&block)
	if err != nil {
		logger.Panic(err)
	}
	return &block
}

func (b *Block) IsGenesis() bool {
	return b.PrevHash == nil
}

// IsBlockValid Checks if the block is valid by confirming variety of information in the block
func (b *Block) IsBlockValid(oldBlock Block) bool {
	if oldBlock.Height+1 != b.Height {
		return false
	}
	res := bytes.Compare(oldBlock.Hash, b.PrevHash)
	if res != 0 {
		return false
	}

	return true
}

// func (b *Block) String() string {
// 	var lines []string
// 	lines = append(lines, fmt.Sprintf("BLOCK: \n Hash: %x", b.Hash))
// 	lines = append(lines, fmt.Sprintf("Merkle Root: %x", b.MerkleRoot))
// 	lines = append(lines, fmt.Sprintf("Height: %d", b.Height))
// 	lines = append(lines, fmt.Sprintf("TxCount: %d", len(b.Transactions)))
// 	return strings.Join(lines, "\n")
// }

// Genesis block
func Genesis(firstTx *Transaction, version int) *Block {
	newBlock := NewBlock([]*Transaction{firstTx}, version, []byte{}, 1)
	return newBlock
}

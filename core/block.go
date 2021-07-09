package core

// EvBlock represent the Block entity of the blockchain
type EvBlock struct {
	Timestamp    int64            `json:"Timestamp"`
	Hash         []byte           `json:"Hash"`
	PrevHash     []byte           `json:"PrevHash"`
	Transactions []*EvTransaction `json:"Transactions"`
	Height       int              `json:"Height"`
	MerkleRoot   []byte           `json:"MerkleRoot"`
	Difficulty   int              `json:"Difficulty"`
	TxCount      int              `json:"TxCount"`
}



package block

import (
	"time"

	"github.com/n3055/liquid-link/internal/transaction"
)

// Header contains the metadata for a block.
type Header struct {
	Index        uint64    `json:"index"`         // Height of the block in the blockchain
	PrevHash     string    `json:"prev_hash"`     // SHA-256 hash of the previous block
	MerkleRoot   string    `json:"merkle_root"`   // Merkle root hash of the transactions
	Timestamp    time.Time `json:"timestamp"`     // Block production time
	Proposer     string    `json:"proposer"`      // ID of the node that proposed the block
	Signature    []byte    `json:"signature"`     // Block signature by the proposer
}

// Block represents a container of transactions added to the blockchain.
type Block struct {
	Header       Header                     `json:"header"`
	Transactions []*transaction.Transaction `json:"transactions"`
	Hash         string                     `json:"hash"` // SHA-256 hash of header and transactions
}

// NewBlock constructs a new block with the given parameters and computes its hash placeholder.
func NewBlock(index uint64, prevHash string, txs []*transaction.Transaction, proposer string) *Block {
	return &Block{
		Header: Header{
			Index:      index,
			PrevHash:   prevHash,
			Timestamp:  time.Now(),
			Proposer:   proposer,
			MerkleRoot: "", // Calculated from txs
		},
		Transactions: txs,
		Hash:         "", // Calculated from Header
	}
}

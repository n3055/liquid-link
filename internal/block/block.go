package block

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/n3055/liquid-link/internal/transaction"
)

// Header contains the metadata for a block.
type Header struct {
	Index      uint64    `json:"index"`       // Height of the block in the blockchain
	PrevHash   string    `json:"prev_hash"`   // SHA-256 hash of the previous block
	MerkleRoot string    `json:"merkle_root"` // Merkle root hash of the transactions
	Timestamp  time.Time `json:"timestamp"`   // Block production time
	Proposer   string    `json:"proposer"`    // ID of the node that proposed the block
	Signature  []byte    `json:"signature"`   // Block signature by the proposer
}

// Block represents a container of transactions added to the blockchain.
type Block struct {
	Header       Header                     `json:"header"`
	Transactions []*transaction.Transaction `json:"transactions"`
	Hash         string                     `json:"hash"` // SHA-256 hash of header and transactions
}

// NewBlock constructs a new block with the given parameters and computes its hash placeholder.
func NewBlock(index uint64, prevHash string, txs []*transaction.Transaction, proposer string) *Block {
	b := &Block{
		Header: Header{
			Index:      index,
			PrevHash:   prevHash,
			Timestamp:  time.Now(),
			Proposer:   proposer,
			MerkleRoot: "",
		},
		Transactions: txs,
		Hash:         "",
	}
	b.Header.MerkleRoot = b.CalculateMerkleRoot()
	b.Hash = b.CalculateHash()
	return b
}

// SigningHash computes the deterministic SHA-256 hash of the block header fields excluding the signature.
// This is the payload signed by the block proposer.
func (b *Block) SigningHash() [32]byte {
	data := fmt.Sprintf("%d|%s|%s|%d|%s",
		b.Header.Index,
		b.Header.PrevHash,
		b.Header.MerkleRoot,
		b.Header.Timestamp.UnixNano(),
		b.Header.Proposer,
	)
	return sha256.Sum256([]byte(data))
}

// CalculateHash computes the final block hash, incorporating the proposer's signature.
func (b *Block) CalculateHash() string {
	sigHash := b.SigningHash()
	combined := append(sigHash[:], b.Header.Signature...)
	hash := sha256.Sum256(combined)
	return hex.EncodeToString(hash[:])
}

// CalculateMerkleRoot computes the cryptographic Merkle root hash of the transactions in the block.
func (b *Block) CalculateMerkleRoot() string {
	if len(b.Transactions) == 0 {
		emptyHash := sha256.Sum256([]byte{})
		return hex.EncodeToString(emptyHash[:])
	}

	var hashes [][]byte
	for _, tx := range b.Transactions {
		h, err := hex.DecodeString(tx.ID)
		if err != nil {
			leaf := sha256.Sum256([]byte(tx.ID))
			h = leaf[:]
		}
		hashes = append(hashes, h)
	}

	for len(hashes) > 1 {
		var nextLevel [][]byte
		for i := 0; i < len(hashes); i += 2 {
			if i+1 < len(hashes) {
				combined := append(hashes[i], hashes[i+1]...)
				parentHash := sha256.Sum256(combined)
				nextLevel = append(nextLevel, parentHash[:])
			} else {
				combined := append(hashes[i], hashes[i]...)
				parentHash := sha256.Sum256(combined)
				nextLevel = append(nextLevel, parentHash[:])
			}
		}
		hashes = nextLevel
	}

	return hex.EncodeToString(hashes[0])
}

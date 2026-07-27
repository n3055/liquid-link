package blockchain

import (
	"context"

	"github.com/n3055/liquid-link/internal/block"
)

// Blockchain represents an interface to read, write, and validate the block sequence.
type Blockchain interface {
	// AddBlock validates and appends a new block to the chain.
	AddBlock(ctx context.Context, b *block.Block) error

	// GetBlockByHeight retrieves a block at the given index/height.
	GetBlockByHeight(ctx context.Context, height uint64) (*block.Block, error)

	// GetBlockByHash retrieves a block by its SHA-256 hash.
	GetBlockByHash(ctx context.Context, hash string) (*block.Block, error)

	// CurrentBlock returns the latest block in the chain (the tip).
	CurrentBlock(ctx context.Context) (*block.Block, error)

	// Height returns the current height of the blockchain.
	Height(ctx context.Context) (uint64, error)
}

// Repository defines storage access for storing and querying blocks.
type Repository interface {
	// SaveBlock persists a block and its associated state updates.
	SaveBlock(ctx context.Context, b *block.Block) error
	// GetBlock retrieves a block from persistence by its hash.
	GetBlock(ctx context.Context, hash string) (*block.Block, error)
	// GetBlockAtHeight retrieves a block from persistence by index.
	GetBlockAtHeight(ctx context.Context, height uint64) (*block.Block, error)
	// GetLastBlock returns the latest block from persistence.
	GetLastBlock(ctx context.Context) (*block.Block, error)
}

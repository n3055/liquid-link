package blockchain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/dgraph-io/badger/v4"
	"github.com/n3055/liquid-link/internal/block"
	"github.com/n3055/liquid-link/internal/crypto"
	"github.com/n3055/liquid-link/internal/storage"
)

var (
	ErrGenesisMismatch   = errors.New("genesis block validation failed")
	ErrIndexMismatch     = errors.New("block index mismatch")
	ErrPrevHashMismatch  = errors.New("previous block hash mismatch")
	ErrTimestampInvalid  = errors.New("block timestamp is not progressive")
	ErrHashMismatch      = errors.New("block hash mismatch")
	ErrMerkleRootInvalid = errors.New("block merkle root mismatch")
	ErrInvalidSignature  = errors.New("block proposer signature is invalid")
	ErrUnknownProposer   = errors.New("unknown block proposer")
)

const (
	GenesisPrevHash = "0000000000000000000000000000000000000000000000000000000000000000"
)

// blockchainImpl implements the Blockchain interface.
type blockchainImpl struct {
	repo       Repository
	verifier   crypto.Verifier
	validators map[string]crypto.PublicKey
}

// NewBlockchain creates a new Blockchain instance with verification parameters.
func NewBlockchain(repo Repository, verifier crypto.Verifier, validators map[string]crypto.PublicKey) Blockchain {
	return &blockchainImpl{
		repo:       repo,
		verifier:   verifier,
		validators: validators,
	}
}

// AddBlock validates and appends a new block to the chain.
func (bc *blockchainImpl) AddBlock(ctx context.Context, b *block.Block) error {
	if b == nil {
		return errors.New("block cannot be nil")
	}

	// 1. Get the current blockchain tip
	tip, err := bc.repo.GetLastBlock(ctx)
	if err != nil && !errors.Is(err, badger.ErrKeyNotFound) && !os.IsNotExist(err) {
		return fmt.Errorf("failed to retrieve blockchain tip: %w", err)
	}

	// 2. Validate block metadata and link
	if tip == nil {
		// This must be the Genesis block
		if b.Header.Index != 0 {
			return fmt.Errorf("%w: expected index 0 for genesis, got %d", ErrGenesisMismatch, b.Header.Index)
		}
		if b.Header.PrevHash != GenesisPrevHash {
			return fmt.Errorf("%w: expected prev hash '%s', got '%s'", ErrGenesisMismatch, GenesisPrevHash, b.Header.PrevHash)
		}
	} else {
		// Validate progressive index
		if b.Header.Index != tip.Header.Index+1 {
			return fmt.Errorf("%w: expected index %d, got %d", ErrIndexMismatch, tip.Header.Index+1, b.Header.Index)
		}
		// Validate prev hash link
		if b.Header.PrevHash != tip.Hash {
			return fmt.Errorf("%w: expected prev hash '%s', got '%s'", ErrPrevHashMismatch, tip.Hash, b.Header.PrevHash)
		}
		// Validate progressive timestamp
		if !b.Header.Timestamp.After(tip.Header.Timestamp) {
			return fmt.Errorf("%w: block timestamp %v is not after tip timestamp %v", ErrTimestampInvalid, b.Header.Timestamp, tip.Header.Timestamp)
		}
	}

	// 3. Validate block hash calculation
	expectedHash := b.CalculateHash()
	if b.Hash != expectedHash {
		return fmt.Errorf("%w: expected '%s', got '%s'", ErrHashMismatch, expectedHash, b.Hash)
	}

	// 4. Validate Merkle root calculation
	expectedMerkle := b.CalculateMerkleRoot()
	if b.Header.MerkleRoot != expectedMerkle {
		return fmt.Errorf("%w: expected '%s', got '%s'", ErrMerkleRootInvalid, expectedMerkle, b.Header.MerkleRoot)
	}

	// 5. Validate proposer and proposer signature
	pubKey, exists := bc.validators[b.Header.Proposer]
	if !exists {
		return fmt.Errorf("%w: '%s'", ErrUnknownProposer, b.Header.Proposer)
	}

	signingHash := b.SigningHash()
	if !bc.verifier.Verify(signingHash[:], b.Header.Signature, pubKey) {
		return ErrInvalidSignature
	}

	// 6. Save the block
	if err := bc.repo.SaveBlock(ctx, b); err != nil {
		return fmt.Errorf("failed to save block: %w", err)
	}

	return nil
}

func (bc *blockchainImpl) GetBlockByHeight(ctx context.Context, height uint64) (*block.Block, error) {
	return bc.repo.GetBlockAtHeight(ctx, height)
}

func (bc *blockchainImpl) GetBlockByHash(ctx context.Context, hash string) (*block.Block, error) {
	return bc.repo.GetBlock(ctx, hash)
}

func (bc *blockchainImpl) CurrentBlock(ctx context.Context) (*block.Block, error) {
	return bc.repo.GetLastBlock(ctx)
}

func (bc *blockchainImpl) Height(ctx context.Context) (uint64, error) {
	tip, err := bc.repo.GetLastBlock(ctx)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) || os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	return tip.Header.Index + 1, nil
}

// badgerRepository implements Repository using storage.Store.
type badgerRepository struct {
	store storage.Store
}

// NewRepository creates a new BadgerDB-backed Repository.
func NewRepository(store storage.Store) Repository {
	return &badgerRepository{store: store}
}

func (r *badgerRepository) SaveBlock(ctx context.Context, b *block.Block) error {
	encoded, err := json.Marshal(b)
	if err != nil {
		return fmt.Errorf("failed to marshal block: %w", err)
	}

	// Use transactional Update to write atomically
	return r.store.Update(func(txn *badger.Txn) error {
		// Save block by Hash
		hashKey := []byte(fmt.Sprintf("block:hash:%s", b.Hash))
		if err := txn.Set(hashKey, encoded); err != nil {
			return err
		}

		// Map Height to Hash
		heightKey := []byte(fmt.Sprintf("block:height:%d", b.Header.Index))
		if err := txn.Set(heightKey, []byte(b.Hash)); err != nil {
			return err
		}

		// Update tip key to current hash
		tipKey := []byte("block:tip")
		return txn.Set(tipKey, []byte(b.Hash))
	})
}

func (r *badgerRepository) GetBlock(ctx context.Context, hash string) (*block.Block, error) {
	hashKey := []byte(fmt.Sprintf("block:hash:%s", hash))
	data, err := r.store.Get(ctx, hashKey)
	if err != nil {
		return nil, err
	}

	var b block.Block
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, fmt.Errorf("failed to unmarshal block: %w", err)
	}

	return &b, nil
}

func (r *badgerRepository) GetBlockAtHeight(ctx context.Context, height uint64) (*block.Block, error) {
	heightKey := []byte(fmt.Sprintf("block:height:%d", height))
	hashBytes, err := r.store.Get(ctx, heightKey)
	if err != nil {
		return nil, err
	}

	return r.GetBlock(ctx, string(hashBytes))
}

func (r *badgerRepository) GetLastBlock(ctx context.Context) (*block.Block, error) {
	tipKey := []byte("block:tip")
	hashBytes, err := r.store.Get(ctx, tipKey)
	if err != nil {
		return nil, err
	}

	return r.GetBlock(ctx, string(hashBytes))
}

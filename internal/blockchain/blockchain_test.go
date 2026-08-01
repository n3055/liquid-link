package blockchain

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/n3055/liquid-link/internal/block"
	"github.com/n3055/liquid-link/internal/config"
	"github.com/n3055/liquid-link/internal/crypto"
	"github.com/n3055/liquid-link/internal/storage"
	"github.com/n3055/liquid-link/internal/transaction"
)

func TestBlockchainGenesisAndAppends(t *testing.T) {
	// Setup storage
	tmpDir := t.TempDir()
	cfg := &config.StorageConfig{Path: tmpDir, ValueLogFileSize: 1024 * 1024}
	store, err := storage.NewBadgerStore(cfg)
	if err != nil {
		t.Fatalf("failed to open Badger store: %v", err)
	}
	defer store.Close()

	repo := NewRepository(store)
	gen := crypto.NewGenerator()

	// Create validators
	kp, err := gen.GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}
	proposerID := "bankA"
	validators := map[string]crypto.PublicKey{
		proposerID: kp.PublicKey,
	}

	bc := NewBlockchain(repo, crypto.NewVerifier(), validators)
	ctx := context.Background()

	// 1. Height should start at 0
	h, err := bc.Height(ctx)
	if err != nil {
		t.Fatalf("failed to check height: %v", err)
	}
	if h != 0 {
		t.Errorf("expected height 0, got %d", h)
	}

	// 2. Try appending a block with index 1 (should fail, no genesis)
	invalidBlock := block.NewBlock(1, GenesisPrevHash, nil, proposerID)
	err = bc.AddBlock(ctx, invalidBlock)
	if !errors.Is(err, ErrGenesisMismatch) {
		t.Errorf("expected ErrGenesisMismatch, got %v", err)
	}

	// 3. Create a valid Genesis block (Index 0, PrevHash is GenesisPrevHash)
	genesis := block.NewBlock(0, GenesisPrevHash, nil, proposerID)
	
	// Proposer signs it
	signer, err := crypto.NewSigner(kp.PrivateKey)
	if err != nil {
		t.Fatalf("failed to create signer: %v", err)
	}
	sigHash := genesis.SigningHash()
	sig, err := signer.Sign(sigHash[:])
	if err != nil {
		t.Fatalf("failed to sign: %v", err)
	}
	genesis.Header.Signature = sig
	genesis.Hash = genesis.CalculateHash()

	// Add genesis
	err = bc.AddBlock(ctx, genesis)
	if err != nil {
		t.Fatalf("failed to add Genesis block: %v", err)
	}

	// 4. Height should be 1
	h, err = bc.Height(ctx)
	if err != nil {
		t.Fatalf("failed to check height: %v", err)
	}
	if h != 1 {
		t.Errorf("expected height 1, got %d", h)
	}

	// 5. Add Block 1
	block1 := block.NewBlock(1, genesis.Hash, []*transaction.Transaction{
		{ID: "tx1", Sender: "0x1", Receiver: "0x2", Amount: 50, Nonce: 1, Timestamp: time.Now()},
	}, proposerID)
	
	// Proposer signs block1
	sigHash1 := block1.SigningHash()
	sig1, err := signer.Sign(sigHash1[:])
	if err != nil {
		t.Fatalf("failed to sign block 1: %v", err)
	}
	block1.Header.Signature = sig1
	block1.Hash = block1.CalculateHash()

	err = bc.AddBlock(ctx, block1)
	if err != nil {
		t.Fatalf("failed to add block 1: %v", err)
	}

	// 6. Height should be 2
	h, err = bc.Height(ctx)
	if err != nil {
		t.Fatalf("failed to check height: %v", err)
	}
	if h != 2 {
		t.Errorf("expected height 2, got %d", h)
	}

	// 7. Verify retrieval by height and hash
	bByHeight, err := bc.GetBlockByHeight(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get block by height: %v", err)
	}
	if bByHeight.Hash != block1.Hash {
		t.Errorf("block hash mismatch on retrieval by height")
	}

	bByHash, err := bc.GetBlockByHash(ctx, genesis.Hash)
	if err != nil {
		t.Fatalf("failed to get block by hash: %v", err)
	}
	if bByHash.Header.Index != 0 {
		t.Errorf("block index mismatch on retrieval by hash")
	}

	// 8. Test validation failure cases
	// Mismatched index
	badBlockIndex := block.NewBlock(3, block1.Hash, nil, proposerID)
	hBad := badBlockIndex.SigningHash()
	sigHashBad, _ := signer.Sign(hBad[:])
	badBlockIndex.Header.Signature = sigHashBad
	badBlockIndex.Hash = badBlockIndex.CalculateHash()
	err = bc.AddBlock(ctx, badBlockIndex)
	if !errors.Is(err, ErrIndexMismatch) {
		t.Errorf("expected ErrIndexMismatch, got %v", err)
	}

	// Mismatched prev hash
	badBlockPrev := block.NewBlock(2, "wrongprevhash", nil, proposerID)
	hBadPrev := badBlockPrev.SigningHash()
	sigHashBadPrev, _ := signer.Sign(hBadPrev[:])
	badBlockPrev.Header.Signature = sigHashBadPrev
	badBlockPrev.Hash = badBlockPrev.CalculateHash()
	err = bc.AddBlock(ctx, badBlockPrev)
	if !errors.Is(err, ErrPrevHashMismatch) {
		t.Errorf("expected ErrPrevHashMismatch, got %v", err)
	}

	// Non-progressive timestamp
	badBlockTime := block.NewBlock(2, block1.Hash, nil, proposerID)
	badBlockTime.Header.Timestamp = block1.Header.Timestamp.Add(-1 * time.Minute)
	hBadTime := badBlockTime.SigningHash()
	sigHashBadTime, _ := signer.Sign(hBadTime[:])
	badBlockTime.Header.Signature = sigHashBadTime
	badBlockTime.Hash = badBlockTime.CalculateHash()
	err = bc.AddBlock(ctx, badBlockTime)
	if !errors.Is(err, ErrTimestampInvalid) {
		t.Errorf("expected ErrTimestampInvalid, got %v", err)
	}

	// Unknown Proposer
	badBlockProposer := block.NewBlock(2, block1.Hash, nil, "bankB")
	hBadProp := badBlockProposer.SigningHash()
	sigHashBadProp, _ := signer.Sign(hBadProp[:])
	badBlockProposer.Header.Signature = sigHashBadProp
	badBlockProposer.Hash = badBlockProposer.CalculateHash()
	err = bc.AddBlock(ctx, badBlockProposer)
	if !errors.Is(err, ErrUnknownProposer) {
		t.Errorf("expected ErrUnknownProposer, got %v", err)
	}
}

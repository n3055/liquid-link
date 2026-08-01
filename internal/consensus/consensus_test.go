package consensus

import (
	"context"
	"testing"
	"time"

	"github.com/n3055/liquid-link/internal/block"
	"github.com/n3055/liquid-link/internal/blockchain"
	"github.com/n3055/liquid-link/internal/config"
	"github.com/n3055/liquid-link/internal/crypto"
	"github.com/n3055/liquid-link/internal/ledger"
	"github.com/n3055/liquid-link/internal/storage"
	"github.com/n3055/liquid-link/internal/transaction"
	"github.com/n3055/liquid-link/internal/wallet"
)

func TestConsensusFlow(t *testing.T) {
	// Setup storage
	tmpDir := t.TempDir()
	cfg := &config.StorageConfig{Path: tmpDir, ValueLogFileSize: 1024 * 1024}
	store, err := storage.NewBadgerStore(cfg)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer store.Close()

	// Setup blockchain & ledger dependencies
	blockRepo := blockchain.NewRepository(store)
	stateRepo := ledger.NewStateRepository(store)
	l := ledger.NewLedger(stateRepo)

	gen := crypto.NewGenerator()

	// Keys for banks
	bankAkp, _ := gen.GenerateKeyPair()
	bankBkp, _ := gen.GenerateKeyPair()
	bankCkp, _ := gen.GenerateKeyPair()

	validators := map[string]crypto.PublicKey{
		"bankA": bankAkp.PublicKey,
		"bankB": bankBkp.PublicKey,
		"bankC": bankCkp.PublicKey,
	}

	bc := blockchain.NewBlockchain(blockRepo, crypto.NewVerifier(), validators)
	ctx := context.Background()

	// Initialize Engine for Node B (bankB)
	engine, err := NewEngine("bankB", bankBkp.PrivateKey, bc, l, validators, 0.66)
	if err != nil {
		t.Fatalf("failed to create consensus engine: %v", err)
	}

	// 1. Create a valid Genesis block proposed by bankA
	genesis := block.NewBlock(0, blockchain.GenesisPrevHash, nil, "bankA")
	
	signerA, _ := crypto.NewSigner(bankAkp.PrivateKey)
	sigHash := genesis.SigningHash()
	sig, _ := signerA.Sign(sigHash[:])
	genesis.Header.Signature = sig
	genesis.Hash = genesis.CalculateHash()

	// Node B receives proposal
	vote, err := engine.ReceiveProposal(ctx, genesis)
	if err != nil {
		t.Fatalf("ReceiveProposal failed: %v", err)
	}

	if vote.Decision != VoteAccept {
		t.Errorf("expected Node B to ACCEPT genesis block, got %v", vote.Decision)
	}

	// Signatures on vote should be valid
	verifier := crypto.NewVerifier()
	voteData := vote.Voter + "|" + vote.BlockHash + "|" + string(vote.Decision)
	if !verifier.Verify([]byte(voteData), vote.Signature, bankBkp.PublicKey) {
		t.Error("vote signature is invalid")
	}

	// Commit Genesis block to blockchain to prepare for block 1 testing
	err = bc.AddBlock(ctx, genesis)
	if err != nil {
		t.Fatalf("failed to add genesis block: %v", err)
	}

	// Seed some funds to bankA using SystemAddress in ledger so we can test transaction verification
	bankAAddr := gen.DeriveAddress(bankAkp.PublicKey)
	err = l.ApplyTransaction(ctx, ledger.SystemAddress, bankAAddr, 1000, 0)
	if err != nil {
		t.Fatalf("minting failed: %v", err)
	}
	err = l.CommitState(ctx, 0)
	if err != nil {
		t.Fatalf("commit failed: %v", err)
	}

	// 2. Propose block 1 with transaction: bankA -> bankC (200 tokens)
	bankCAddr := gen.DeriveAddress(bankCkp.PublicKey)
	tx := &transaction.Transaction{
		Sender:    bankAAddr,
		Receiver:  bankCAddr,
		Amount:    200,
		Nonce:     1,
		Timestamp: time.Now(),
	}

	// Sign transaction using bankA's wallet
	walletA := wallet.NewWallet(bankAkp)
	err = walletA.SignTransaction(ctx, tx)
	if err != nil {
		t.Fatalf("failed to sign transaction: %v", err)
	}

	block1 := block.NewBlock(1, genesis.Hash, []*transaction.Transaction{tx}, "bankA")
	sigHash1 := block1.SigningHash()
	sig1, _ := signerA.Sign(sigHash1[:])
	block1.Header.Signature = sig1
	block1.Hash = block1.CalculateHash()

	// Node B receives proposal for block 1
	vote1, err := engine.ReceiveProposal(ctx, block1)
	if err != nil {
		t.Fatalf("ReceiveProposal failed for block 1: %v", err)
	}

	if vote1.Decision != VoteAccept {
		t.Errorf("expected Node B to ACCEPT block 1 proposal, got %v", vote1.Decision)
	}

	// 3. Propose invalid block 1: Sender balance is too low
	invalidTx := &transaction.Transaction{
		Sender:    bankAAddr,
		Receiver:  bankCAddr,
		Amount:    20000, // exceeds 1000 balance!
		Nonce:     1,
		Timestamp: time.Now(),
	}
	_ = walletA.SignTransaction(ctx, invalidTx)

	block1Bad := block.NewBlock(1, genesis.Hash, []*transaction.Transaction{invalidTx}, "bankA")
	sigHash1Bad := block1Bad.SigningHash()
	sig1Bad, _ := signerA.Sign(sigHash1Bad[:])
	block1Bad.Header.Signature = sig1Bad
	block1Bad.Hash = block1Bad.CalculateHash()

	vote1Bad, err := engine.ReceiveProposal(ctx, block1Bad)
	if err != nil {
		t.Fatalf("ReceiveProposal failed for block 1 bad: %v", err)
	}
	if vote1Bad.Decision != VoteReject {
		t.Error("expected Node B to REJECT block 1 due to insufficient funds")
	}

	// 4. Test CollectVotes majority tallying
	// Create votes from B and C
	voteB := vote1 // ACCEPT

	voteC := &Vote{
		Voter:     "bankC",
		BlockHash: block1.Hash,
		Decision:  VoteAccept,
	}
	signerC, _ := crypto.NewSigner(bankCkp.PrivateKey)
	voteCData := voteC.Voter + "|" + voteC.BlockHash + "|" + string(voteC.Decision)
	sigC, _ := signerC.Sign([]byte(voteCData))
	voteC.Signature = sigC

	// Tally votes: threshold = 0.66, total = 3 validators. 
	// 2/3 = 66.6% -> should pass
	passed, err := engine.CollectVotes(ctx, block1.Hash, []*Vote{voteB, voteC})
	if err != nil {
		t.Fatalf("CollectVotes failed: %v", err)
	}
	if !passed {
		t.Error("expected block to pass with 2/3 votes")
	}

	// Tally votes with only 1/3 (should fail)
	passed, _ = engine.CollectVotes(ctx, block1.Hash, []*Vote{voteB})
	if passed {
		t.Error("expected block to fail with only 1/3 votes")
	}
}

func TestLeaderSelector(t *testing.T) {
	validators := []string{"bankA", "bankC", "bankB"}
	ls := NewLeaderSelector(validators)
	ctx := context.Background()

	// Should be sorted: bankA, bankB, bankC
	leader, err := ls.CurrentLeader(ctx)
	if err != nil {
		t.Fatalf("failed to get current leader: %v", err)
	}
	if leader != "bankA" {
		t.Errorf("expected bankA, got %s", leader)
	}

	// Rotate
	err = ls.Rotate(ctx)
	if err != nil {
		t.Fatalf("failed to rotate leader: %v", err)
	}
	leader, _ = ls.CurrentLeader(ctx)
	if leader != "bankB" {
		t.Errorf("expected bankB, got %s", leader)
	}

	// Rotate again
	_ = ls.Rotate(ctx)
	leader, _ = ls.CurrentLeader(ctx)
	if leader != "bankC" {
		t.Errorf("expected bankC, got %s", leader)
	}

	// Rotate wraps around
	_ = ls.Rotate(ctx)
	leader, _ = ls.CurrentLeader(ctx)
	if leader != "bankA" {
		t.Errorf("expected bankA, got %s", leader)
	}
}

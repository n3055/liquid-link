package wallet

import (
	"context"
	"crypto/sha256"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/n3055/liquid-link/internal/crypto"
	"github.com/n3055/liquid-link/internal/transaction"
)

func TestGenerateWallet(t *testing.T) {
	w, err := GenerateWallet()
	if err != nil {
		t.Fatalf("failed to generate wallet: %v", err)
	}

	if w.Address() == "" {
		t.Error("wallet address is empty")
	}

	pub := w.ExportPublicKey()
	if len(pub) == 0 {
		t.Error("exported public key is empty")
	}

	lw, ok := w.(interface{ ExportPrivateKey() []byte })
	if !ok {
		t.Error("wallet does not implement ExportPrivateKey")
	} else if len(lw.ExportPrivateKey()) == 0 {
		t.Error("exported private key is empty")
	}
}

func TestSignTransaction(t *testing.T) {
	w, err := GenerateWallet()
	if err != nil {
		t.Fatalf("failed to generate wallet: %v", err)
	}

	tx := &transaction.Transaction{
		Receiver:  crypto.Address("0xreceiveraddress"),
		Amount:    100,
		Nonce:     1,
		Timestamp: time.Now(),
	}

	ctx := context.Background()
	err = w.SignTransaction(ctx, tx)
	if err != nil {
		t.Fatalf("failed to sign transaction: %v", err)
	}

	if tx.Sender != w.Address() {
		t.Errorf("expected sender to be %s, got %s", w.Address(), tx.Sender)
	}

	if tx.ID == "" {
		t.Error("transaction ID (hash) was not populated")
	}

	if len(tx.Signature) == 0 {
		t.Error("transaction signature was not populated")
	}

	// Verify the signature using the verifier
	verifier := crypto.NewVerifier()
	payload := SerializeTx(tx)
	
	// Wait, we signed the hash of the payload in localWallet.SignTransaction:
	// hash := sha256.Sum256(payload)
	// signer.Sign(hash[:])
	// So let's reproduce the hash check
	// We can use a SHA-256 hash
	hash := sha256Sum(payload)
	if !verifier.Verify(hash[:], tx.Signature, w.ExportPublicKey()) {
		t.Error("signature verification failed")
	}
}

func TestSignTransactionSenderMismatch(t *testing.T) {
	w, err := GenerateWallet()
	if err != nil {
		t.Fatalf("failed to generate wallet: %v", err)
	}

	tx := &transaction.Transaction{
		Sender:    crypto.Address("0xotheraddress"),
		Receiver:  crypto.Address("0xreceiveraddress"),
		Amount:    100,
		Nonce:     1,
		Timestamp: time.Now(),
	}

	ctx := context.Background()
	err = w.SignTransaction(ctx, tx)
	if err == nil {
		t.Error("expected error when signing transaction with sender address mismatch, got nil")
	}
}

func TestFileRepository(t *testing.T) {
	tmpDir := t.TempDir()

	repo, err := NewFileRepository(tmpDir)
	if err != nil {
		t.Fatalf("failed to create FileRepository: %v", err)
	}

	w, err := GenerateWallet()
	if err != nil {
		t.Fatalf("failed to generate wallet: %v", err)
	}

	ctx := context.Background()
	alias := "my-test-wallet"

	// Test Save
	err = repo.Save(ctx, alias, w)
	if err != nil {
		t.Fatalf("failed to save wallet: %v", err)
	}

	// Verify file was created and permissions are 0600
	filePath := filepath.Join(tmpDir, alias+".json")
	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("wallet file was not created: %v", err)
	}

	// Check file permissions (on Unix-like systems, we can check 0600)
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("expected wallet file permissions 0600, got %o", perm)
	}

	// Test Load
	loadedW, err := repo.Load(ctx, alias)
	if err != nil {
		t.Fatalf("failed to load wallet: %v", err)
	}

	if loadedW.Address() != w.Address() {
		t.Errorf("loaded address %s does not match original %s", loadedW.Address(), w.Address())
	}

	// Test List
	wallets, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("failed to list wallets: %v", err)
	}

	addr, exists := wallets[alias]
	if !exists {
		t.Errorf("wallet with alias %s not found in list", alias)
	}
	if addr != w.Address() {
		t.Errorf("listed address %s does not match original %s", addr, w.Address())
	}

	// Test Load Non-existent Wallet
	_, err = repo.Load(ctx, "non-existent")
	if err == nil {
		t.Error("expected error loading non-existent wallet, got nil")
	}
}

// Helper function to hash payload for verifying signature in test.
func sha256Sum(data []byte) [32]byte {
	return sha256.Sum256(data)
}

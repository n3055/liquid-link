package wallet

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/n3055/liquid-link/internal/crypto"
	"github.com/n3055/liquid-link/internal/transaction"
)

// localWallet is the concrete implementation of the Wallet interface.
type localWallet struct {
	keyPair *crypto.KeyPair
	address crypto.Address
}

// NewWallet creates a new Wallet instance from an existing KeyPair.
func NewWallet(keyPair *crypto.KeyPair) Wallet {
	gen := crypto.NewGenerator()
	addr := gen.DeriveAddress(keyPair.PublicKey)
	return &localWallet{
		keyPair: keyPair,
		address: addr,
	}
}

// GenerateWallet generates a new random Wallet.
func GenerateWallet() (Wallet, error) {
	gen := crypto.NewGenerator()
	kp, err := gen.GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate keypair: %w", err)
	}
	return NewWallet(kp), nil
}

// Address returns the primary address of the wallet.
func (w *localWallet) Address() crypto.Address {
	return w.address
}

// ExportPublicKey exports the raw public key of the wallet.
func (w *localWallet) ExportPublicKey() crypto.PublicKey {
	return w.keyPair.PublicKey
}

// ExportPrivateKey exports the raw private key of the wallet.
func (w *localWallet) ExportPrivateKey() []byte {
	return []byte(w.keyPair.PrivateKey)
}

// SignTransaction signs a transaction payload using the wallet's private key.
func (w *localWallet) SignTransaction(ctx context.Context, tx *transaction.Transaction) error {
	if tx == nil {
		return errors.New("transaction cannot be nil")
	}

	// Override or verify Sender address
	if tx.Sender == "" {
		tx.Sender = w.address
	} else if tx.Sender != w.address {
		return fmt.Errorf("transaction sender address %s does not match wallet address %s", tx.Sender, w.address)
	}

	// Compute serialization format
	payload := SerializeTx(tx)

	// Hash payload to produce ID
	hash := sha256.Sum256(payload)
	tx.ID = hex.EncodeToString(hash[:])

	// Sign the hash
	signer, err := crypto.NewSigner(w.keyPair.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to initialize signer: %w", err)
	}

	sig, err := signer.Sign(hash[:])
	if err != nil {
		return fmt.Errorf("failed to sign transaction ID: %w", err)
	}

	tx.Signature = sig
	return nil
}

// SerializeTx serializes a transaction into a deterministic byte format for hashing and signing.
func SerializeTx(tx *transaction.Transaction) []byte {
	// Format: sender|receiver|amount|nonce|timestamp
	return []byte(fmt.Sprintf("%s|%s|%d|%d|%d",
		tx.Sender,
		tx.Receiver,
		tx.Amount,
		tx.Nonce,
		tx.Timestamp.UnixNano(),
	))
}

// walletJSON defines the serialization format for storing wallets on disk.
type walletJSON struct {
	PrivateKey string         `json:"private_key"`
	PublicKey  string         `json:"public_key"`
	Address    crypto.Address `json:"address"`
}

// fileRepository implements the Repository interface for local storage.
type fileRepository struct {
	baseDir string
}

// NewFileRepository creates a new FileRepository at the specified directory.
func NewFileRepository(baseDir string) (Repository, error) {
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create wallet directory: %w", err)
	}
	return &fileRepository{baseDir: baseDir}, nil
}

// Save stores a wallet securely in local storage.
func (r *fileRepository) Save(ctx context.Context, alias string, w Wallet) error {
	if alias == "" {
		return errors.New("wallet alias cannot be empty")
	}

	lw, ok := w.(*localWallet)
	if !ok {
		return errors.New("unsupported wallet implementation")
	}

	data := walletJSON{
		PrivateKey: hex.EncodeToString(lw.keyPair.PrivateKey),
		PublicKey:  hex.EncodeToString(lw.keyPair.PublicKey),
		Address:    lw.address,
	}

	encoded, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal wallet: %w", err)
	}

	filePath := filepath.Join(r.baseDir, alias+".json")
	if err := os.WriteFile(filePath, encoded, 0600); err != nil {
		return fmt.Errorf("failed to write wallet file: %w", err)
	}

	return nil
}

// Load retrieves a wallet by its alias from local storage.
func (r *fileRepository) Load(ctx context.Context, alias string) (Wallet, error) {
	if alias == "" {
		return nil, errors.New("wallet alias cannot be empty")
	}

	filePath := filepath.Join(r.baseDir, alias+".json")
	encoded, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("wallet with alias %s not found", alias)
		}
		return nil, fmt.Errorf("failed to read wallet file: %w", err)
	}

	var data walletJSON
	if err := json.Unmarshal(encoded, &data); err != nil {
		return nil, fmt.Errorf("failed to parse wallet JSON: %w", err)
	}

	privBytes, err := hex.DecodeString(data.PrivateKey)
	if err != nil || len(privBytes) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key in wallet file")
	}

	pubBytes, err := hex.DecodeString(data.PublicKey)
	if err != nil || len(pubBytes) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key in wallet file")
	}

	kp := &crypto.KeyPair{
		PrivateKey: crypto.PrivateKey(privBytes),
		PublicKey:  crypto.PublicKey(pubBytes),
	}

	return &localWallet{
		keyPair: kp,
		address: data.Address,
	}, nil
}

// List returns all stored wallet addresses and their aliases.
func (r *fileRepository) List(ctx context.Context) (map[string]crypto.Address, error) {
	entries, err := os.ReadDir(r.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read wallet directory: %w", err)
	}

	wallets := make(map[string]crypto.Address)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		alias := entry.Name()[:len(entry.Name())-5] // strip .json
		w, err := r.Load(ctx, alias)
		if err != nil {
			continue // skip corrupted or unreadable wallets
		}
		wallets[alias] = w.Address()
	}

	return wallets, nil
}

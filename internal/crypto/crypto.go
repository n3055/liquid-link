package crypto

import (
	"crypto/ed25519"
)

// PrivateKey represents an Ed25519 private key.
type PrivateKey ed25519.PrivateKey

// PublicKey represents an Ed25519 public key.
type PublicKey ed25519.PublicKey

// Address represents a base58 or hex encoded string derived from a PublicKey.
type Address string

// KeyPair holds a matched pair of public and private keys.
type KeyPair struct {
	PrivateKey PrivateKey
	PublicKey  PublicKey
}

// Signer defines the interface for signing arbitrary payloads.
type Signer interface {
	// Sign generates a signature for the given message.
	Sign(message []byte) ([]byte, error)
	// PublicKey returns the signer's corresponding public key.
	PublicKey() PublicKey
}

// Verifier defines the interface for validating signatures.
type Verifier interface {
	// Verify checks if the signature is valid for the message under the public key.
	Verify(message, signature []byte, pubKey PublicKey) bool
}

// Generator defines the interface for keypair and address generation.
type Generator interface {
	// GenerateKeyPair generates a new cryptographically secure Ed25519 keypair.
	GenerateKeyPair() (*KeyPair, error)
	// DeriveAddress derives a unique human-readable address from a public key.
	DeriveAddress(pubKey PublicKey) Address
}

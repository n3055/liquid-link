package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

// ed25519Generator implements the Generator interface.
type ed25519Generator struct{}

// NewGenerator creates a new default implementation of the Generator interface.
func NewGenerator() Generator {
	return &ed25519Generator{}
}

// GenerateKeyPair generates a new cryptographically secure Ed25519 keypair.
func (g *ed25519Generator) GenerateKeyPair() (*KeyPair, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}
	return &KeyPair{
		PrivateKey: PrivateKey(priv),
		PublicKey:  PublicKey(pub),
	}, nil
}

// DeriveAddress derives a unique human-readable address from a public key.
// It computes the SHA-256 hash of the public key and hex-encodes the first 20 bytes,
// prepending a "0x" prefix.
func (g *ed25519Generator) DeriveAddress(pubKey PublicKey) Address {
	hash := sha256.Sum256([]byte(pubKey))
	return Address("0x" + hex.EncodeToString(hash[:20]))
}

// ed25519Signer implements the Signer interface.
type ed25519Signer struct {
	privKey PrivateKey
}

// NewSigner creates a new Signer instance wrapping an Ed25519 private key.
func NewSigner(privKey PrivateKey) (Signer, error) {
	if len(privKey) != ed25519.PrivateKeySize {
		return nil, errors.New("invalid private key size")
	}
	return &ed25519Signer{privKey: privKey}, nil
}

// Sign generates an Ed25519 signature for the given message.
func (s *ed25519Signer) Sign(message []byte) ([]byte, error) {
	sig := ed25519.Sign(ed25519.PrivateKey(s.privKey), message)
	return sig, nil
}

// PublicKey returns the signer's corresponding public key.
func (s *ed25519Signer) PublicKey() PublicKey {
	// An ed25519 private key contains the public key in its trailing 32 bytes.
	return PublicKey(ed25519.PrivateKey(s.privKey).Public().(ed25519.PublicKey))
}

// ed25519Verifier implements the Verifier interface.
type ed25519Verifier struct{}

// NewVerifier creates a new Verifier instance.
func NewVerifier() Verifier {
	return &ed25519Verifier{}
}

// Verify checks if the signature is valid for the message under the public key.
func (v *ed25519Verifier) Verify(message, signature []byte, pubKey PublicKey) bool {
	if len(pubKey) != ed25519.PublicKeySize {
		return false
	}
	return ed25519.Verify(ed25519.PublicKey(pubKey), message, signature)
}

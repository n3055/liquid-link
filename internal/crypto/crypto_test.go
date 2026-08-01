package crypto

import (
	"bytes"
	"testing"
)

func TestGenerateKeyPair(t *testing.T) {
	gen := NewGenerator()
	kp, err := gen.GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	if len(kp.PrivateKey) == 0 {
		t.Error("private key is empty")
	}
	if len(kp.PublicKey) == 0 {
		t.Error("public key is empty")
	}
}

func TestDeriveAddress(t *testing.T) {
	gen := NewGenerator()
	kp, err := gen.GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	addr := gen.DeriveAddress(kp.PublicKey)
	if len(addr) != 42 { // "0x" + 40 hex chars
		t.Errorf("expected address length 42, got %d", len(addr))
	}
	if addr[:2] != "0x" {
		t.Errorf("expected address to start with '0x', got %s", addr[:2])
	}
}

func TestSignAndVerify(t *testing.T) {
	gen := NewGenerator()
	kp, err := gen.GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	signer, err := NewSigner(kp.PrivateKey)
	if err != nil {
		t.Fatalf("failed to create signer: %v", err)
	}

	message := []byte("hello liquid-link")
	sig, err := signer.Sign(message)
	if err != nil {
		t.Fatalf("failed to sign message: %v", err)
	}

	verifier := NewVerifier()
	if !verifier.Verify(message, sig, kp.PublicKey) {
		t.Error("failed to verify signature")
	}

	// Verify failure with wrong message
	wrongMessage := []byte("hello liquid-link!")
	if verifier.Verify(wrongMessage, sig, kp.PublicKey) {
		t.Error("expected verification to fail for wrong message")
	}

	// Verify failure with wrong public key
	kp2, err := gen.GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate second key pair: %v", err)
	}
	if verifier.Verify(message, sig, kp2.PublicKey) {
		t.Error("expected verification to fail for wrong public key")
	}
}

func TestSignerInvalidPrivateKey(t *testing.T) {
	_, err := NewSigner(PrivateKey([]byte("invalid-key")))
	if err == nil {
		t.Error("expected error creating signer with invalid private key, got nil")
	}
}

func TestVerifierInvalidPublicKey(t *testing.T) {
	verifier := NewVerifier()
	ok := verifier.Verify([]byte("msg"), []byte("sig"), PublicKey([]byte("invalid")))
	if ok {
		t.Error("expected verify to fail for invalid public key length")
	}
}

func TestPublicKeyDerivation(t *testing.T) {
	gen := NewGenerator()
	kp, err := gen.GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	signer, err := NewSigner(kp.PrivateKey)
	if err != nil {
		t.Fatalf("failed to create signer: %v", err)
	}

	pubKey := signer.PublicKey()
	if !bytes.Equal(pubKey, kp.PublicKey) {
		t.Error("derived public key does not match generated public key")
	}
}

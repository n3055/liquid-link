package consensus

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/dgraph-io/badger/v4"
	"github.com/n3055/liquid-link/internal/block"
	"github.com/n3055/liquid-link/internal/blockchain"
	"github.com/n3055/liquid-link/internal/crypto"
	"github.com/n3055/liquid-link/internal/ledger"
	"github.com/n3055/liquid-link/internal/wallet"
)

// consensusEngine manages proposals and counts majority votes from trusted validators.
type consensusEngine struct {
	nodeID     string
	signer     crypto.Signer
	verifier   crypto.Verifier
	bc         blockchain.Blockchain
	ledger     ledger.Ledger
	validators map[string]crypto.PublicKey
	threshold  float64
}

// NewEngine creates a new consensus Engine instance.
func NewEngine(
	nodeID string,
	privKey crypto.PrivateKey,
	bc blockchain.Blockchain,
	l ledger.Ledger,
	validators map[string]crypto.PublicKey,
	threshold float64,
) (Engine, error) {
	signer, err := crypto.NewSigner(privKey)
	if err != nil {
		return nil, fmt.Errorf("failed to construct signer: %w", err)
	}

	return &consensusEngine{
		nodeID:     nodeID,
		signer:     signer,
		verifier:   crypto.NewVerifier(),
		bc:         bc,
		ledger:     l,
		validators: validators,
		threshold:  threshold,
	}, nil
}

// ProposeBlock broadcasts a newly generated block to all peer nodes (simulated or API request).
func (e *consensusEngine) ProposeBlock(ctx context.Context, b *block.Block) error {
	if b == nil {
		return errors.New("cannot propose nil block")
	}
	// Proposal signature verification is handled in ReceiveProposal
	return nil
}

// ReceiveProposal processes an incoming block proposal, validates it, and casts a vote.
func (e *consensusEngine) ReceiveProposal(ctx context.Context, b *block.Block) (*Vote, error) {
	if b == nil {
		return nil, errors.New("received nil block proposal")
	}

	decision := VoteAccept

	// 1. Dry run metadata validations against blockchain tip
	tip, err := e.bc.CurrentBlock(ctx)
	if err != nil && !errors.Is(err, badger.ErrKeyNotFound) && !os.IsNotExist(err) {
		decision = VoteReject
	} else {
		isEmpty := tip == nil || (err != nil && (errors.Is(err, badger.ErrKeyNotFound) || os.IsNotExist(err)))
		if isEmpty {
			if b.Header.Index != 0 || b.Header.PrevHash != blockchain.GenesisPrevHash {
				decision = VoteReject
			}
		} else {
			if b.Header.Index != tip.Header.Index+1 || b.Header.PrevHash != tip.Hash || !b.Header.Timestamp.After(tip.Header.Timestamp) {
				decision = VoteReject
			}
		}
	}

	// 2. Validate hash consistency
	if decision == VoteAccept {
		if b.Hash != b.CalculateHash() || b.Header.MerkleRoot != b.CalculateMerkleRoot() {
			decision = VoteReject
		}
	}

	// 3. Verify proposer identity and signature
	if decision == VoteAccept {
		pubKey, exists := e.validators[b.Header.Proposer]
		if !exists {
			decision = VoteReject
		} else {
			signingHash := b.SigningHash()
			if !e.verifier.Verify(signingHash[:], b.Header.Signature, pubKey) {
				decision = VoteReject
			}
		}
	}

	// 4. Dry-run transaction validations and execution on the ledger
	if decision == VoteAccept {
		var applyErr error
		for _, tx := range b.Transactions {
			if tx.Sender != ledger.SystemAddress {
				// Verify sender public key matches derived address
				pubKey, found := findPublicKeyForAddress(tx.Sender, e.validators)
				if !found {
					applyErr = fmt.Errorf("public key not found for sender address: %s", tx.Sender)
					break
				}

				// Verify tx signature
				payload := wallet.SerializeTx(tx)
				txHash := sha256.Sum256(payload)
				if !e.verifier.Verify(txHash[:], tx.Signature, pubKey) {
					applyErr = errors.New("invalid transaction signature")
					break
				}
			}

			// Apply transition to ledger cache
			if err := e.ledger.ApplyTransaction(ctx, tx.Sender, tx.Receiver, tx.Amount, tx.Nonce); err != nil {
				applyErr = err
				break
			}
		}

		// Rollback state changes since this is a dry-run check
		_ = e.ledger.RollbackState(ctx)

		if applyErr != nil {
			decision = VoteReject
		}
	}

	// 5. Build and sign vote ballot
	vote := &Vote{
		Voter:     e.nodeID,
		BlockHash: b.Hash,
		Decision:  decision,
	}

	voteData := fmt.Sprintf("%s|%s|%s", vote.Voter, vote.BlockHash, vote.Decision)
	sig, err := e.signer.Sign([]byte(voteData))
	if err != nil {
		return nil, fmt.Errorf("failed to sign vote: %w", err)
	}
	vote.Signature = sig

	return vote, nil
}

// CollectVotes gathers and tallies votes for a proposed block.
func (e *consensusEngine) CollectVotes(ctx context.Context, blockHash string, votes []*Vote) (bool, error) {
	if len(votes) == 0 {
		return false, nil
	}

	seen := make(map[string]bool)
	acceptCount := 0

	for _, vote := range votes {
		if vote == nil || vote.BlockHash != blockHash {
			continue
		}

		pubKey, exists := e.validators[vote.Voter]
		if !exists {
			continue
		}

		if seen[vote.Voter] {
			continue // prevent duplicate votes
		}
		seen[vote.Voter] = true

		// Verify vote signature
		voteData := fmt.Sprintf("%s|%s|%s", vote.Voter, vote.BlockHash, vote.Decision)
		if !e.verifier.Verify([]byte(voteData), vote.Signature, pubKey) {
			continue // skip invalid signatures
		}

		if vote.Decision == VoteAccept {
			acceptCount++
		}
	}

	totalValidators := len(e.validators)
	if totalValidators == 0 {
		return false, errors.New("empty validator set")
	}

	ratio := float64(acceptCount) / float64(totalValidators)
	return ratio >= e.threshold, nil
}

// roundRobinLeaderSelector implements deterministic rotation of block proposers.
type roundRobinLeaderSelector struct {
	mu           sync.RWMutex
	validators   []string
	currentIndex int
}

// NewLeaderSelector creates a round-robin LeaderSelector instance.
func NewLeaderSelector(validatorIDs []string) LeaderSelector {
	sortedIDs := make([]string, len(validatorIDs))
	copy(sortedIDs, validatorIDs)
	sort.Strings(sortedIDs)

	return &roundRobinLeaderSelector{
		validators: sortedIDs,
	}
}

func (s *roundRobinLeaderSelector) CurrentLeader(ctx context.Context) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.validators) == 0 {
		return "", errors.New("no validators available")
	}
	return s.validators[s.currentIndex], nil
}

func (s *roundRobinLeaderSelector) IsLeader(ctx context.Context, nodeID string) (bool, error) {
	leader, err := s.CurrentLeader(ctx)
	if err != nil {
		return false, err
	}
	return leader == nodeID, nil
}

func (s *roundRobinLeaderSelector) Rotate(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.validators) == 0 {
		return errors.New("no validators available")
	}
	s.currentIndex = (s.currentIndex + 1) % len(s.validators)
	return nil
}

// findPublicKeyForAddress looks up the matching public key for a given Address from the trusted validator set.
func findPublicKeyForAddress(addr crypto.Address, validators map[string]crypto.PublicKey) (crypto.PublicKey, bool) {
	gen := crypto.NewGenerator()
	for _, pub := range validators {
		if gen.DeriveAddress(pub) == addr {
			return pub, true
		}
	}
	return nil, false
}

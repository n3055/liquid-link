package consensus

import (
	"context"

	"github.com/n3055/liquid-link/internal/block"
)

// VoteType represents a node's decision on a proposed block.
type VoteType string

const (
	VoteAccept VoteType = "ACCEPT"
	VoteReject VoteType = "REJECT"
)

// Vote represents a signed decision from a consensus-participating node.
type Vote struct {
	Voter     string   `json:"voter"`      // Node ID of the voter
	BlockHash string   `json:"block_hash"` // Hash of the block being voted on
	Decision  VoteType `json:"decision"`   // ACCEPT or REJECT
	Signature []byte   `json:"signature"`  // Voter's cryptographic signature
}

// Engine manages the consensus flow, receiving blocks, voting, and counting majority.
type Engine interface {
	// ProposeBlock broadcasts a newly generated block to all peer nodes.
	ProposeBlock(ctx context.Context, b *block.Block) error

	// ReceiveProposal processes an incoming block proposal, validates it, and casts a vote.
	ReceiveProposal(ctx context.Context, b *block.Block) (*Vote, error)

	// CollectVotes gathers and tallies votes for a proposed block.
	CollectVotes(ctx context.Context, blockHash string, votes []*Vote) (bool, error)
}

// LeaderSelector handles the deterministic leader rotation mechanism.
type LeaderSelector interface {
	// CurrentLeader returns the node ID of the active leader.
	CurrentLeader(ctx context.Context) (string, error)

	// IsLeader checks if the calling node is currently the leader.
	IsLeader(ctx context.Context, nodeID string) (bool, error)

	// Rotate increments the round or time window to select the next leader.
	Rotate(ctx context.Context) error
}

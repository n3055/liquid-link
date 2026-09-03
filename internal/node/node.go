package node

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"time"

	"github.com/n3055/liquid-link/internal/api"
	"github.com/n3055/liquid-link/internal/block"
	"github.com/n3055/liquid-link/internal/blockchain"
	"github.com/n3055/liquid-link/internal/config"
	"github.com/n3055/liquid-link/internal/consensus"
	"github.com/n3055/liquid-link/internal/crypto"
	"github.com/n3055/liquid-link/internal/ledger"
	"github.com/n3055/liquid-link/internal/pcn"
	"github.com/n3055/liquid-link/internal/peer"
	"github.com/n3055/liquid-link/internal/receipt"
	"github.com/n3055/liquid-link/internal/storage"
	"github.com/n3055/liquid-link/internal/token"
	"github.com/n3055/liquid-link/internal/transport"
	"github.com/n3055/liquid-link/internal/wallet"
)

// Node represents the primary execution block for the LiquidLink daemon.
type Node struct {
	cfg          *config.Config
	logger       *slog.Logger
	store        storage.Store
	ledger       ledger.Ledger
	blockchain   blockchain.Blockchain
	consensus    consensus.Engine
	pcnManager   pcn.Manager
	bridge       receipt.Bridge
	transports   []transport.Transport
	peerRegistry peer.Registry
	wallet       wallet.Wallet
	assetManager token.AssetManager
	signer       crypto.Signer
	stopChan     chan struct{}
	apiServer    *api.Server
	httpServer   *http.Server
}

// NewNode instantiates the node, injecting all concrete implementations.
func NewNode(cfg *config.Config, logger *slog.Logger) (*Node, error) {
	// 1. Initialize Storage
	store, err := storage.NewBadgerStore(&cfg.Storage)
	if err != nil {
		return nil, fmt.Errorf("failed to load storage: %w", err)
	}

	// 2. Initialize Ledger State & Adapter
	stateRepo := ledger.NewStateRepository(store)
	led := ledger.NewLedger(stateRepo)
	balanceAdapter := ledger.NewLedgerBalanceAdapter(led.(ledger.StateRepository))

	// 3. Initialize Asset Manager
	assetMgr := token.NewAssetManager(balanceAdapter, store)

	// 4. Crypto Keys & Wallet — persistent wallet storage
	// [Wallet] Derive wallet storage directory from the BadgerDB storage path.
	walletDir := filepath.Join(filepath.Dir(cfg.Storage.Path), "wallets")
	walletRepo, err := wallet.NewFileRepository(walletDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet repository: %w", err)
	}

	ctx := context.Background()
	walletAlias := cfg.Node.ID
	isNewWallet := false

	var w wallet.Wallet
	w, err = walletRepo.Load(ctx, walletAlias)
	if err != nil {
		// No existing wallet found — generate a new one and save it
		logger.Info("generating new wallet for node", slog.String("node_id", cfg.Node.ID))
		w, err = wallet.GenerateWallet()
		if err != nil {
			return nil, fmt.Errorf("failed to generate node wallet: %w", err)
		}
		if err := walletRepo.Save(ctx, walletAlias, w); err != nil {
			return nil, fmt.Errorf("failed to persist node wallet: %w", err)
		}
		isNewWallet = true
	} else {
		logger.Info("loaded existing wallet", slog.String("node_id", cfg.Node.ID), slog.String("address", string(w.Address())))
	}

	// Extract key pair for signer and consensus validator registration
	pubKey := w.ExportPublicKey()

	keyExporter, ok := w.(interface{ ExportPrivateKey() []byte })
	if !ok {
		return nil, fmt.Errorf("node wallet does not support private key export")
	}
	privKeyBytes := keyExporter.ExportPrivateKey()

	signer, err := crypto.NewSigner(crypto.PrivateKey(privKeyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize signer: %w", err)
	}

	verifier := crypto.NewVerifier()

	// Mint initial genesis allocation only on first-ever wallet creation
	if isNewWallet {
		_ = assetMgr.Mint(ctx, w.Address(), 100000)
		logger.Info("minted genesis allocation", slog.String("address", string(w.Address())), slog.Int("amount", 100000))
	}

	// 5. Blockchain & Consensus
	blockRepo := blockchain.NewRepository(store)
	privKey := crypto.PrivateKey(privKeyBytes)
	validators := map[string]crypto.PublicKey{
		cfg.Node.ID: pubKey,
	}

	bc := blockchain.NewBlockchain(blockRepo, verifier, validators)

	cons, err := consensus.NewEngine(cfg.Node.ID, privKey, bc, led, validators, 0.5)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize consensus engine: %w", err)
	}

	// 6. PCN Manager & Registry
	pcnRegistry := pcn.NewHolderRegistry(store)
	pcnMgr := pcn.NewManager(w.Address(), signer, led, pcnRegistry, assetMgr)

	// 7. Receipt Bridge & Registry
	receiptReg := receipt.NewRegistry(store, verifier, pubKey)
	br := receipt.NewBridge(receiptReg, assetMgr, signer, store)

	// 8. Peer Registry
	peerReg := peer.NewRegistry(store)

	// Register self in peer registry
	_ = peerReg.RegisterPeer(ctx, &peer.Info{
		ID:        cfg.Node.ID,
		Address:   w.Address(),
		Endpoint:  fmt.Sprintf("http://localhost:%d", cfg.Node.Port),
		Connected: true,
	})

	transports := make([]transport.Transport, 0)

	// 9. API Server with full dependencies injected
	apiSrv := api.NewServer(
		&cfg.Node,
		led,
		bc,
		cons,
		pcnMgr,
		br,
		w,
		peerReg,
		assetMgr,
		pcnRegistry,
		receiptReg,
	)

	httpSrv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Node.Port),
		Handler: apiSrv.Handler(),
	}

	return &Node{
		cfg:          cfg,
		logger:       logger,
		store:        store,
		ledger:       led,
		blockchain:   bc,
		consensus:    cons,
		pcnManager:   pcnMgr,
		bridge:       br,
		transports:   transports,
		peerRegistry: peerReg,
		wallet:       w,
		assetManager: assetMgr,
		signer:       signer,
		stopChan:     make(chan struct{}),
		apiServer:    apiSrv,
		httpServer:   httpSrv,
	}, nil
}

// Start boots all transports, network listener ports, background synchronization, and the HTTP server.
func (n *Node) Start(ctx context.Context) error {
	n.logger.Info("starting LiquidLink node", slog.String("node_id", n.cfg.Node.ID))

	// Start all transports
	for _, t := range n.transports {
		if err := t.Start(ctx); err != nil {
			return fmt.Errorf("failed to start transport %s: %w", t.Medium(), err)
		}
	}

	// Start API Server
	go func() {
		n.logger.Info("starting HTTP API Server", slog.Int("port", n.cfg.Node.Port))
		if err := n.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			n.logger.Error("HTTP server failure", slog.Any("error", err))
		}
	}()

	// [Consensus Block] Launch background automated block producer ticker loop
	go n.startBlockProducer(ctx)

	return nil
}

// Stop executes graceful termination of all active transports, DB hooks, and HTTP servers.
func (n *Node) Stop(ctx context.Context) error {
	n.logger.Info("stopping LiquidLink node")

	if n.stopChan != nil {
		close(n.stopChan)
	}

	// Shutdown HTTP Server
	if err := n.httpServer.Shutdown(ctx); err != nil {
		n.logger.Error("HTTP server shutdown error", slog.Any("error", err))
	}

	// Stop all transports
	for _, t := range n.transports {
		if err := t.Stop(ctx); err != nil {
			n.logger.Error("failed to stop transport", slog.Any("medium", t.Medium()), slog.Any("error", err))
		}
	}

	// Close database
	if err := n.store.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	return nil
}

// [Consensus Block] startBlockProducer runs a background ticker loop proposing and committing blocks to the blockchain.
func (n *Node) startBlockProducer(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// 1. Ensure Genesis block exists on startup
	tip, err := n.blockchain.CurrentBlock(ctx)
	if err != nil || tip == nil {
		genBlock := block.NewBlock(0, blockchain.GenesisPrevHash, nil, n.cfg.Node.ID)
		sigHash := genBlock.SigningHash()
		if n.signer != nil {
			if sig, err := n.signer.Sign(sigHash[:]); err == nil {
				genBlock.Header.Signature = sig
			}
		}
		genBlock.Hash = genBlock.CalculateHash()
		_ = n.blockchain.AddBlock(ctx, genBlock)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-n.stopChan:
			return
		case <-ticker.C:
			currentTip, err := n.blockchain.CurrentBlock(ctx)
			if err != nil || currentTip == nil {
				continue
			}

			nextIndex := currentTip.Header.Index + 1
			nextBlock := block.NewBlock(nextIndex, currentTip.Hash, nil, n.cfg.Node.ID)
			sigHash := nextBlock.SigningHash()
			if n.signer != nil {
				sig, err := n.signer.Sign(sigHash[:])
				if err != nil {
					continue
				}
				nextBlock.Header.Signature = sig
			}
			nextBlock.Hash = nextBlock.CalculateHash()

			// Validate with consensus engine and commit to blockchain
			if _, err := n.consensus.ReceiveProposal(ctx, nextBlock); err == nil {
				_ = n.blockchain.AddBlock(ctx, nextBlock)
			}
		}
	}
}

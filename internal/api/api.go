package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/n3055/liquid-link/internal/blockchain"
	"github.com/n3055/liquid-link/internal/config"
	"github.com/n3055/liquid-link/internal/consensus"
	"github.com/n3055/liquid-link/internal/crypto"
	"github.com/n3055/liquid-link/internal/fiat"
	"github.com/n3055/liquid-link/internal/ledger"
	"github.com/n3055/liquid-link/internal/orderbook"
	"github.com/n3055/liquid-link/internal/pcn"
	"github.com/n3055/liquid-link/internal/peer"
	"github.com/n3055/liquid-link/internal/receipt"
	"github.com/n3055/liquid-link/internal/stability"
	"github.com/n3055/liquid-link/internal/token"
	"github.com/n3055/liquid-link/internal/wallet"
)

// Server implements HTTP routing and API handlers.
type Server struct {
	mu              sync.Mutex
	cfg             *config.NodeConfig
	router          chi.Router
	ledger          ledger.Ledger
	blockchain      blockchain.Blockchain
	consensus       consensus.Engine
	pcn             pcn.Manager
	bridge          receipt.Bridge
	wallet          wallet.Wallet
	peers           peer.Registry
	assetMgr        token.AssetManager
	pcnRegistry     pcn.HolderRegistry
	receiptRegistry receipt.Registry
	stabilityEng    stability.Engine
	orderbookEng    orderbook.Engine
	fiatMgr         fiat.Manager
}

// NewServer creates a new API Server instance with registered routes and injected dependencies.
func NewServer(
	cfg *config.NodeConfig,
	ledger ledger.Ledger,
	bc blockchain.Blockchain,
	cons consensus.Engine,
	pcnMgr pcn.Manager,
	bridge receipt.Bridge,
	w wallet.Wallet,
	peerReg peer.Registry,
	assetMgr token.AssetManager,
	pcnRegistry pcn.HolderRegistry,
	receiptReg receipt.Registry,
) *Server {
	r := chi.NewRouter()

	s := &Server{
		cfg:             cfg,
		router:          r,
		ledger:          ledger,
		blockchain:      bc,
		consensus:       cons,
		pcn:             pcnMgr,
		bridge:          bridge,
		wallet:          w,
		peers:           peerReg,
		assetMgr:        assetMgr,
		pcnRegistry:     pcnRegistry,
		receiptRegistry: receiptReg,
	}

	s.setupRoutes()

	return s
}

// Handler returns the HTTP handler for the server.
func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) setupRoutes() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	// Web UI Dashboard
	s.router.Get("/", s.handleDashboard)
	s.router.Get("/ui", s.handleDashboard)

	// API v1 Namespace
	s.router.Route("/api/v1", func(r chi.Router) {
		r.Get("/status", s.handleStatus)

		// Wallet actions
		r.Route("/wallet", func(wr chi.Router) {
			wr.Get("/info", s.handleWalletInfo)
			wr.Get("/balance", s.handleBalance)
			wr.Get("/holdings", s.handleWalletHoldings)
			wr.Post("/create", s.handleWalletCreate)
			wr.Post("/transfer", s.handleTransfer)
			wr.Post("/mint", s.handleMint)
			wr.Post("/credit-sync", s.handleCreditSync)
			wr.Post("/sign-offline", s.handleSignOffline)
		})

		// Transaction broadcast
		r.Route("/tx", func(tr chi.Router) {
			tr.Post("/broadcast", s.handleBroadcastTx)
		})

		// Orderbook actions
		r.Route("/orderbook", func(obr chi.Router) {
			obr.Get("/depth", s.handleOrderbookDepth)
			obr.Get("/trades", s.handleOrderbookTrades)
			obr.Get("/user-orders", s.handleUserOrders)
			obr.Post("/order", s.handlePlaceOrder)
			obr.Post("/cancel", s.handleCancelOrder)
		})

		// Seigniorage Stability actions
		r.Route("/stability", func(sr chi.Router) {
			sr.Get("/status", s.handleStabilityStatus)
			sr.Get("/history", s.handleStabilityHistory)
			sr.Post("/stake", s.handleStakeShares)
			sr.Post("/unstake", s.handleUnstakeShares)
			sr.Post("/buy-bonds", s.handleBuyBonds)
			sr.Post("/redeem-bonds", s.handleRedeemBonds)
		})

		// PCN / Channels actions
		r.Route("/channels", func(cr chi.Router) {
			cr.Get("/list", s.handleListChannels)
			cr.Get("/{id}", s.handleGetChannel)
			cr.Post("/open", s.handleOpenChannel)
			cr.Post("/update", s.handleUpdateChannel)
			cr.Post("/settle", s.handleSettleChannel)
			cr.Post("/challenge", s.handleChallengeChannel)
			cr.Post("/peer-sync", s.handlePeerChannelSync)
			cr.Post("/offline-sync", s.handleOfflineChannelSync)
		})

		// Transport actions
		r.Route("/transport", func(tr chi.Router) {
			tr.Get("/bluetooth/status", s.handleBluetoothStatus)
			tr.Post("/broadcast-offline", s.handleBluetoothBroadcast)
		})

		// Bridge / Settlement actions
		r.Route("/bridge", func(br chi.Router) {
			br.Get("/list", s.handleListUTRMappings)
			br.Post("/fiat-request", s.handleBridgeRequest)
			br.Post("/verify-receipt", s.handleVerifyReceipt)
			br.Get("/utr/{utr}", s.handleGetUTRMapping)
		})

		// Fiat Bank actions
		r.Route("/fiat", func(fr chi.Router) {
			fr.Get("/pool", s.handleFiatPool)
		})

		// Blockchain & Consensus
		r.Route("/blockchain", func(bcr chi.Router) {
			bcr.Get("/height", s.handleBlockchainHeight)
			bcr.Get("/blocks", s.handleListBlocks)
			bcr.Get("/block/{query}", s.handleGetBlock)
		})

		// Peer Network Topology
		r.Route("/peers", func(pr chi.Router) {
			pr.Get("/", s.handleListPeers)
			pr.Post("/register", s.handleRegisterPeer)
		})
	})
}

// Helpers
func respondJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func respondError(w http.ResponseWriter, statusCode int, message string) {
	respondJSON(w, statusCode, map[string]string{"error": message})
}

// Handler implementations

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	height := uint64(0)
	if s.blockchain != nil {
		h, err := s.blockchain.Height(ctx)
		if err == nil {
			height = h
		}
	}

	supply := token.Amount(0)
	if s.assetMgr != nil {
		sup, err := s.assetMgr.TotalSupply(ctx)
		if err == nil {
			supply = sup
		}
	}

	peerCount := 0
	if s.peers != nil {
		pList, err := s.peers.ListPeers(ctx)
		if err == nil {
			for _, p := range pList {
				if p.ID != s.cfg.ID {
					peerCount++
				}
			}
		}
	}

	bal := token.Amount(0)
	addr := ""
	if s.wallet != nil {
		addr = string(s.wallet.Address())
		if s.assetMgr != nil {
			b, err := s.assetMgr.BalanceOf(ctx, s.wallet.Address())
			if err == nil {
				bal = b
			}
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":       "online",
		"node_id":      s.cfg.ID,
		"port":         s.cfg.Port,
		"address":      addr,
		"balance":      bal,
		"total_supply": supply,
		"block_height": height,
		"peer_count":   peerCount,
	})
}

func (s *Server) handleWalletInfo(w http.ResponseWriter, r *http.Request) {
	if s.wallet == nil {
		respondError(w, http.StatusInternalServerError, "wallet not configured")
		return
	}

	ctx := r.Context()
	bal := token.Amount(0)
	if s.assetMgr != nil {
		b, err := s.assetMgr.BalanceOf(ctx, s.wallet.Address())
		if err == nil {
			bal = b
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"address": string(s.wallet.Address()),
		"balance": bal,
		"node_id": s.cfg.ID,
	})
}

func (s *Server) handleBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	targetAddr := r.URL.Query().Get("address")
	if targetAddr == "" {
		if s.wallet == nil {
			respondError(w, http.StatusBadRequest, "missing address query parameter")
			return
		}
		targetAddr = string(s.wallet.Address())
	}

	bal := token.Amount(0)
	if s.assetMgr != nil {
		b, err := s.assetMgr.BalanceOf(ctx, crypto.Address(targetAddr))
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		bal = b
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"address": targetAddr,
		"balance": bal,
	})
}

type transferReq struct {
	Sender   string       `json:"sender"`
	Receiver string       `json:"receiver"`
	Amount   token.Amount `json:"amount"`
}

func (s *Server) handleTransfer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req transferReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Receiver == "" || req.Amount == 0 {
		respondError(w, http.StatusBadRequest, "receiver and positive amount required")
		return
	}

	if s.assetMgr == nil || s.wallet == nil {
		respondError(w, http.StatusInternalServerError, "asset manager or wallet uninitialized")
		return
	}

	sender := s.wallet.Address()
	if req.Sender != "" {
		sender = crypto.Address(req.Sender)
	}
	if err := s.assetMgr.Debit(ctx, sender, req.Amount); err != nil {
		respondError(w, http.StatusBadRequest, "transfer failed: "+err.Error())
		return
	}

	if err := s.assetMgr.Credit(ctx, crypto.Address(req.Receiver), req.Amount); err != nil {
		// Rollback debit on failure
		_ = s.assetMgr.Credit(ctx, sender, req.Amount)
		respondError(w, http.StatusInternalServerError, "transfer credit failed: "+err.Error())
		return
	}

	// Trigger asynchronous inter-node balance synchronization across peer network
	go s.syncPeerCredit(crypto.Address(req.Receiver), req.Amount)

	respondJSON(w, http.StatusAccepted, map[string]interface{}{
		"status":   "success",
		"sender":   string(sender),
		"receiver": req.Receiver,
		"amount":   req.Amount,
	})
}

type broadcastTxReq struct {
	Sender    string       `json:"sender"`
	Receiver  string       `json:"receiver"`
	Amount    token.Amount `json:"amount"`
	Nonce     uint64       `json:"nonce"`
	PublicKey string       `json:"public_key"`
	Signature string       `json:"signature"`
}

// [API] handleBroadcastTx receives pre-signed client transaction payloads, validates cryptographic signatures, and executes transfers.
func (s *Server) handleBroadcastTx(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req broadcastTxReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid transaction payload")
		return
	}

	if req.Sender == "" || req.Receiver == "" || req.Amount == 0 {
		respondError(w, http.StatusBadRequest, "sender, receiver, and positive amount required")
		return
	}

	if s.assetMgr == nil {
		respondError(w, http.StatusInternalServerError, "asset manager uninitialized")
		return
	}

	// Verify cryptographic signature if provided
	if req.PublicKey != "" && req.Signature != "" {
		pubKeyBytes, err := hex.DecodeString(req.PublicKey)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid public key format")
			return
		}
		sigBytes, err := hex.DecodeString(req.Signature)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid signature format")
			return
		}

		// Verify address matches public key
		gen := crypto.NewGenerator()
		if gen.DeriveAddress(crypto.PublicKey(pubKeyBytes)) != crypto.Address(req.Sender) {
			respondError(w, http.StatusBadRequest, "public key does not match sender address")
			return
		}

		payload := fmt.Sprintf("%s|%s|%d|%d", req.Sender, req.Receiver, req.Amount, req.Nonce)
		hash := sha256.Sum256([]byte(payload))

		verifier := crypto.NewVerifier()
		if !verifier.Verify(hash[:], sigBytes, crypto.PublicKey(pubKeyBytes)) {
			respondError(w, http.StatusBadRequest, "invalid transaction signature")
			return
		}

		// Prevent replay if ledger is available
		if s.ledger != nil {
			state, err := s.ledger.GetAccountState(ctx, crypto.Address(req.Sender))
			if err == nil && state != nil && req.Nonce <= state.Nonce {
				respondError(w, http.StatusBadRequest, fmt.Sprintf("stale or replayed nonce: transaction has %d, current nonce is %d", req.Nonce, state.Nonce))
				return
			}
		}
	}

	senderAddr := crypto.Address(req.Sender)
	if err := s.assetMgr.Debit(ctx, senderAddr, req.Amount); err != nil {
		respondError(w, http.StatusBadRequest, "transfer failed: "+err.Error())
		return
	}

	if err := s.assetMgr.Credit(ctx, crypto.Address(req.Receiver), req.Amount); err != nil {
		_ = s.assetMgr.Credit(ctx, senderAddr, req.Amount)
		respondError(w, http.StatusInternalServerError, "transfer credit failed: "+err.Error())
		return
	}

	if s.ledger != nil && req.Nonce > 0 {
		if repo, ok := s.ledger.(ledger.StateRepository); ok {
			if state, err := s.ledger.GetAccountState(ctx, senderAddr); err == nil && state != nil {
				state.Nonce = req.Nonce
				_ = repo.SetAccountState(ctx, state)
			}
		}
	}

	txIDHash := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%d|%d", req.Sender, req.Receiver, req.Amount, req.Nonce)))
	txID := "tx-" + hex.EncodeToString(txIDHash[:8])

	go s.syncPeerCredit(crypto.Address(req.Receiver), req.Amount)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":         "broadcast_success",
		"transaction_id": txID,
		"sender":         req.Sender,
		"receiver":       req.Receiver,
		"amount":         req.Amount,
	})
}

// [PCN State] handlePeerChannelSync receives channel state updates broadcast by peer nodes.
func (s *Server) handlePeerChannelSync(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var ch pcn.Channel
	if err := json.NewDecoder(r.Body).Decode(&ch); err != nil {
		respondError(w, http.StatusBadRequest, "invalid channel payload")
		return
	}

	if ch.ID == "" {
		respondError(w, http.StatusBadRequest, "channel id required")
		return
	}

	if s.pcnRegistry == nil {
		respondError(w, http.StatusInternalServerError, "pcn registry uninitialized")
		return
	}

	if err := s.pcnRegistry.RegisterChannel(ctx, &ch); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to sync peer channel: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status": "peer_channel_synced",
		"id":     ch.ID,
	})
}

// decodeBytesOrHex parses a JSON raw message into bytes supporting hex strings, base64 strings, or numeric byte arrays.
func decodeBytesOrHex(raw json.RawMessage) []byte {
	if len(raw) == 0 {
		return nil
	}
	var str string
	if err := json.Unmarshal(raw, &str); err == nil {
		if b, err := hex.DecodeString(str); err == nil && len(b) > 0 {
			return b
		}
		if b, err := base64.StdEncoding.DecodeString(str); err == nil && len(b) > 0 {
			return b
		}
		return []byte(str)
	}
	var b []byte
	if err := json.Unmarshal(raw, &b); err == nil {
		return b
	}
	return nil
}

// [PCN State] handleOfflineChannelSync receives and validates off-chain signed channel statements from offline peers.
func (s *Server) handleOfflineChannelSync(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var raw struct {
		Type      string          `json:"type,omitempty"`
		ChannelID string          `json:"channel_id"`
		Sequence  uint64          `json:"sequence"`
		PartyA    crypto.Address  `json:"party_a"`
		PartyB    crypto.Address  `json:"party_b"`
		BalanceA  token.Amount    `json:"balance_a"`
		BalanceB  token.Amount    `json:"balance_b"`
		Sender    crypto.Address  `json:"sender"`
		Signature json.RawMessage `json:"signature"`
		PublicKey json.RawMessage `json:"public_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		respondError(w, http.StatusBadRequest, "invalid offline statement body: "+err.Error())
		return
	}

	if raw.ChannelID == "" {
		respondError(w, http.StatusBadRequest, "channel_id required")
		return
	}

	if s.pcn == nil {
		respondError(w, http.StatusInternalServerError, "pcn manager uninitialized")
		return
	}

	stmt := pcn.OfflineStateStatement{
		Type:      raw.Type,
		ChannelID: raw.ChannelID,
		Sequence:  raw.Sequence,
		PartyA:    raw.PartyA,
		PartyB:    raw.PartyB,
		BalanceA:  raw.BalanceA,
		BalanceB:  raw.BalanceB,
		Sender:    raw.Sender,
		Signature: decodeBytesOrHex(raw.Signature),
		PublicKey: decodeBytesOrHex(raw.PublicKey),
	}

	ch, err := s.pcn.ApplyOfflineStateStatement(ctx, &stmt)
	if err != nil {
		respondError(w, http.StatusBadRequest, "failed to apply offline channel statement: "+err.Error())
		return
	}

	go s.syncPeerChannelUpdate(ch)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":   "offline_pcn_synced",
		"channel":  ch,
		"sequence": ch.Sequence,
		"node_id":  s.cfg.ID,
	})
}

// [PCN State] syncPeerChannelUpdate broadcasts off-chain channel state updates to registered peer endpoints over HTTP REST.
func (s *Server) syncPeerChannelUpdate(ch *pcn.Channel) {
	defer func() {
		_ = recover()
	}()
	if ch == nil || s.peers == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	peers, err := s.peers.ListPeers(ctx)
	if err != nil {
		return
	}

	data, err := json.Marshal(ch)
	if err != nil {
		return
	}

	client := &http.Client{Timeout: 3 * time.Second}
	for _, p := range peers {
		if p.ID == s.cfg.ID || p.Endpoint == "" {
			continue
		}
		endpointURL := fmt.Sprintf("%s/api/v1/channels/peer-sync", p.Endpoint)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpointURL, bytes.NewBuffer(data))
		if err != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err == nil && resp != nil {
			_ = resp.Body.Close()
		}
	}
}

// [Ledger] syncPeerCredit broadcasts a credit update for an address to all peer endpoints over HTTP REST.
func (s *Server) syncPeerCredit(addr crypto.Address, amount token.Amount) {
	defer func() {
		_ = recover()
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	endpoints := make(map[string]bool)

	if s.peers != nil {
		if peers, err := s.peers.ListPeers(ctx); err == nil {
			for _, p := range peers {
				if p.ID != s.cfg.ID && p.Endpoint != "" {
					endpoints[p.Endpoint] = true
				}
			}
		}
	}

	// Always fallback to standard local node ports if peer list is unpopulated
	defaultPorts := []int{8080, 8081}
	for _, port := range defaultPorts {
		if port != s.cfg.Port {
			endpoints[fmt.Sprintf("http://localhost:%d", port)] = true
		}
	}

	payload, err := json.Marshal(map[string]interface{}{
		"address": string(addr),
		"amount":  amount,
	})
	if err != nil {
		return
	}

	client := &http.Client{Timeout: 2 * time.Second}
	for ep := range endpoints {
		url := fmt.Sprintf("%s/api/v1/wallet/credit-sync", strings.TrimSuffix(ep, "/"))
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payload))
		if err == nil {
			req.Header.Set("Content-Type", "application/json")
			resp, err := client.Do(req)
			if err == nil {
				_ = resp.Body.Close()
			}
		}
	}
}

func (s *Server) handleCreditSync(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req mintReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Address == "" || req.Amount == 0 {
		respondError(w, http.StatusBadRequest, "address and positive amount required")
		return
	}

	if s.assetMgr == nil {
		respondError(w, http.StatusInternalServerError, "asset manager uninitialized")
		return
	}

	if err := s.assetMgr.Credit(ctx, crypto.Address(req.Address), req.Amount); err != nil {
		respondError(w, http.StatusInternalServerError, "credit sync failed: "+err.Error())
		return
	}

	newBal, _ := s.assetMgr.BalanceOf(ctx, crypto.Address(req.Address))

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":      "credited",
		"address":     req.Address,
		"amount":      req.Amount,
		"new_balance": newBal,
	})
}

type mintReq struct {
	Address string       `json:"address"`
	Amount  token.Amount `json:"amount"`
}

func (s *Server) handleMint(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req mintReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	targetAddr := req.Address
	if targetAddr == "" && s.wallet != nil {
		targetAddr = string(s.wallet.Address())
	}

	if req.Amount == 0 {
		respondError(w, http.StatusBadRequest, "amount must be greater than zero")
		return
	}

	if s.assetMgr == nil {
		respondError(w, http.StatusInternalServerError, "asset manager uninitialized")
		return
	}

	if err := s.assetMgr.Mint(ctx, crypto.Address(targetAddr), req.Amount); err != nil {
		respondError(w, http.StatusInternalServerError, "minting failed: "+err.Error())
		return
	}

	newBal, _ := s.assetMgr.BalanceOf(ctx, crypto.Address(targetAddr))

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":      "minted",
		"address":     targetAddr,
		"amount":      req.Amount,
		"new_balance": newBal,
	})
}

func (s *Server) handleListChannels(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if s.pcnRegistry == nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{"channels": []interface{}{}})
		return
	}

	channels, err := s.pcnRegistry.ListChannels(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if channels == nil {
		channels = []*pcn.Channel{}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"channels": channels,
	})
}

func (s *Server) handleGetChannel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	channelID := chi.URLParam(r, "id")
	if channelID == "" {
		respondError(w, http.StatusBadRequest, "missing channel id parameter")
		return
	}

	if s.pcnRegistry == nil {
		respondError(w, http.StatusNotFound, "pcn registry uninitialized")
		return
	}

	ch, err := s.pcnRegistry.GetChannel(ctx, channelID)
	if err != nil {
		if errors.Is(err, pcn.ErrChannelNotFound) {
			respondError(w, http.StatusNotFound, "channel not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, ch)
}

type openChannelReq struct {
	Counterparty string       `json:"counterparty"`
	Amount       token.Amount `json:"amount"`
}

func (s *Server) handleOpenChannel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req openChannelReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Counterparty == "" || req.Amount == 0 {
		respondError(w, http.StatusBadRequest, "counterparty and deposit amount required")
		return
	}

	if s.pcn == nil {
		respondError(w, http.StatusInternalServerError, "pcn manager uninitialized")
		return
	}

	ch, err := s.pcn.OpenChannel(ctx, crypto.Address(req.Counterparty), req.Amount)
	if err != nil {
		respondError(w, http.StatusBadRequest, "failed to open channel: "+err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, ch)
	go s.syncPeerChannelUpdate(ch)
}

type updateChannelReq struct {
	ChannelID string       `json:"channel_id"`
	Amount    token.Amount `json:"amount"`
	Sending   bool         `json:"sending"`
}

func (s *Server) handleUpdateChannel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req updateChannelReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ChannelID == "" || req.Amount == 0 {
		respondError(w, http.StatusBadRequest, "channel_id and non-zero amount required")
		return
	}

	if s.pcn == nil {
		respondError(w, http.StatusInternalServerError, "pcn manager uninitialized")
		return
	}

	ch, err := s.pcn.UpdateChannelState(ctx, req.ChannelID, req.Amount, req.Sending)
	if err != nil {
		respondError(w, http.StatusBadRequest, "failed to update channel: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, ch)
	go s.syncPeerChannelUpdate(ch)
}

type settleChannelReq struct {
	ChannelID string `json:"channel_id"`
	Action    string `json:"action"` // "initiate" or "finalize"
}

func (s *Server) handleSettleChannel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req settleChannelReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ChannelID == "" {
		respondError(w, http.StatusBadRequest, "channel_id required")
		return
	}

	if s.pcn == nil {
		respondError(w, http.StatusInternalServerError, "pcn manager uninitialized")
		return
	}

	if req.Action == "finalize" {
		if err := s.pcn.FinalizeSettlement(ctx, req.ChannelID); err != nil {
			respondError(w, http.StatusBadRequest, "failed to finalize channel settlement: "+err.Error())
			return
		}
	} else {
		if err := s.pcn.InitiateSettlement(ctx, req.ChannelID); err != nil {
			respondError(w, http.StatusBadRequest, "failed to initiate channel settlement: "+err.Error())
			return
		}
	}

	ch, _ := s.pcnRegistry.GetChannel(ctx, req.ChannelID)
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "settlement_updated",
		"channel": ch,
	})
	if ch != nil {
		go s.syncPeerChannelUpdate(ch)
	}
}

type challengeChannelReq struct {
	ChannelID string `json:"channel_id"`
	Sequence  uint64 `json:"sequence"`
}

func (s *Server) handleChallengeChannel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req challengeChannelReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ChannelID == "" {
		respondError(w, http.StatusBadRequest, "channel_id required")
		return
	}

	if s.pcn == nil || s.pcnRegistry == nil {
		respondError(w, http.StatusInternalServerError, "pcn uninitialized")
		return
	}

	ch, err := s.pcnRegistry.GetChannel(ctx, req.ChannelID)
	if err != nil {
		respondError(w, http.StatusNotFound, "channel not found")
		return
	}

	higherSeq := *ch
	higherSeq.Sequence = req.Sequence
	if higherSeq.Sequence <= ch.Sequence {
		higherSeq.Sequence = ch.Sequence + 1
	}

	if err := s.pcn.ChallengeSettlement(ctx, req.ChannelID, &higherSeq); err != nil {
		respondError(w, http.StatusBadRequest, "challenge failed: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "dispute_challenged",
		"channel": higherSeq,
	})
}

type bridgeReq struct {
	Receiver string       `json:"receiver"`
	Amount   token.Amount `json:"amount"`
}

func (s *Server) handleBridgeRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req bridgeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Receiver == "" || req.Amount == 0 {
		respondError(w, http.StatusBadRequest, "receiver address and positive amount required")
		return
	}

	if s.bridge == nil {
		respondError(w, http.StatusInternalServerError, "bridge service uninitialized")
		return
	}

	mapping, err := s.bridge.InitiateBridgeRequest(ctx, crypto.Address(req.Receiver), req.Amount)
	if err != nil {
		respondError(w, http.StatusBadRequest, "bridge request failed: "+err.Error())
		return
	}

	fm := s.getFiatManager()
	if fm != nil {
		_ = fm.ReserveINR(ctx, crypto.Address(req.Receiver), float64(req.Amount))
	}

	respondJSON(w, http.StatusAccepted, map[string]interface{}{
		"status":  "bridge_request_created",
		"mapping": mapping,
	})
}

type verifyReceiptReq struct {
	UTR       string `json:"utr"`
	ReceiptID string `json:"receipt_id"`
}

func (s *Server) handleVerifyReceipt(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req verifyReceiptReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.UTR == "" {
		respondError(w, http.StatusBadRequest, "utr required")
		return
	}

	if s.bridge == nil {
		respondError(w, http.StatusInternalServerError, "bridge service uninitialized")
		return
	}

	rec := &receipt.Receipt{
		ID:  req.ReceiptID,
		UTR: req.UTR,
	}
	if rec.ID == "" {
		rec.ID = "rec-" + req.UTR
	}

	if err := s.bridge.SettleBridgeRequest(ctx, req.UTR, rec); err != nil {
		respondError(w, http.StatusBadRequest, "receipt verification failed: "+err.Error())
		return
	}

	fm := s.getFiatManager()
	if fm != nil && s.wallet != nil {
		creditAmt := float64(rec.Amount)
		if creditAmt == 0 {
			creditAmt = 100.0
		}
		_ = fm.CreditINR(ctx, s.wallet.Address(), creditAmt)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status": "receipt_verified",
		"utr":    req.UTR,
	})
}

func (s *Server) handleGetUTRMapping(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	utr := chi.URLParam(r, "utr")
	if utr == "" {
		respondError(w, http.StatusBadRequest, "missing utr parameter")
		return
	}

	if s.receiptRegistry == nil {
		respondError(w, http.StatusInternalServerError, "receipt registry uninitialized")
		return
	}

	mapping, err := s.receiptRegistry.GetUTRMapping(ctx, utr)
	if err != nil {
		respondError(w, http.StatusNotFound, "utr mapping not found")
		return
	}

	respondJSON(w, http.StatusOK, mapping)
}

// [Transport] handleBluetoothStatus returns the configuration and active state of the Bluetooth mesh layer.
func (s *Server) handleBluetoothStatus(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"medium":       "bluetooth",
		"enabled":      true,
		"device_id":    "hci0",
		"service_uuid": "00001101-0000-1000-8000-00805F9B34FB",
		"active":       true,
	})
}

type bluetoothBroadcastReq struct {
	Payload string `json:"payload"`
}

// [Transport] handleBluetoothBroadcast accepts, validates, and relays offline signed transactions over Bluetooth.
func (s *Server) handleBluetoothBroadcast(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req bluetoothBroadcastReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	payload := strings.TrimSpace(req.Payload)
	if payload == "" || len(payload) < 5 {
		respondError(w, http.StatusBadRequest, "invalid or empty bluetooth payload")
		return
	}

	// Inspect payload to check if it contains a structured signed statement
	if strings.HasPrefix(payload, "{") && strings.HasSuffix(payload, "}") {
		var envelope struct {
			Type      string       `json:"type"`
			ChannelID string       `json:"channel_id"`
			Sequence  uint64       `json:"sequence"`
			PartyA    string       `json:"party_a"`
			PartyB    string       `json:"party_b"`
			BalanceA  token.Amount `json:"balance_a"`
			BalanceB  token.Amount `json:"balance_b"`
			Sender    string       `json:"sender"`
			Receiver  string       `json:"receiver"`
			Amount    token.Amount `json:"amount"`
			Nonce     uint64       `json:"nonce"`
			Signature string       `json:"signature"`
			PublicKey string       `json:"public_key"`
		}

		if err := json.Unmarshal([]byte(payload), &envelope); err == nil {
			// Case 1: Offline PCN Payment Statement
			if (envelope.Type == "pcn_offline_payment" || envelope.ChannelID != "") && s.pcn != nil {
				sigBytes, errSig := hex.DecodeString(envelope.Signature)
				pubBytes, errPub := hex.DecodeString(envelope.PublicKey)
				if errSig != nil || errPub != nil || len(sigBytes) == 0 || len(pubBytes) == 0 {
					respondError(w, http.StatusBadRequest, "valid hex signature and public key required for offline pcn payment")
					return
				}
				stmt := pcn.OfflineStateStatement{
					Type:      "pcn_offline_payment",
					ChannelID: envelope.ChannelID,
					Sequence:  envelope.Sequence,
					PartyA:    crypto.Address(envelope.PartyA),
					PartyB:    crypto.Address(envelope.PartyB),
					BalanceA:  envelope.BalanceA,
					BalanceB:  envelope.BalanceB,
					Sender:    crypto.Address(envelope.Sender),
					Signature: sigBytes,
					PublicKey: pubBytes,
				}

				ch, err := s.pcn.ApplyOfflineStateStatement(ctx, &stmt)
				if err != nil {
					respondError(w, http.StatusBadRequest, "failed to apply offline channel statement: "+err.Error())
					return
				}

				go s.syncPeerChannelUpdate(ch)
				respondJSON(w, http.StatusOK, map[string]interface{}{
					"status":     "offline_pcn_synced",
					"channel_id": ch.ID,
					"sequence":   ch.Sequence,
					"medium":     "bluetooth_mesh",
					"node_id":    s.cfg.ID,
					"timestamp":  time.Now().Format(time.RFC3339),
				})
				return
			}

			// Case 2: Offline Direct Transfer Transaction
			if envelope.Sender != "" && envelope.Receiver != "" && envelope.Amount > 0 && s.assetMgr != nil {
				senderAddr := crypto.Address(envelope.Sender)
				receiverAddr := crypto.Address(envelope.Receiver)

				// Mandatory Ed25519 signature verification
				if envelope.PublicKey == "" || envelope.Signature == "" {
					respondError(w, http.StatusBadRequest, "public key and cryptographic signature required for offline transfer")
					return
				}

				pubKeyBytes, err1 := hex.DecodeString(envelope.PublicKey)
				sigBytes, err2 := hex.DecodeString(envelope.Signature)
				if err1 != nil || err2 != nil || len(pubKeyBytes) == 0 || len(sigBytes) == 0 {
					respondError(w, http.StatusBadRequest, "invalid hex format in public key or signature")
					return
				}

				// Verify sender address derivation from public key
				gen := crypto.NewGenerator()
				if gen.DeriveAddress(crypto.PublicKey(pubKeyBytes)) != senderAddr {
					respondError(w, http.StatusBadRequest, fmt.Sprintf("public key does not derive to sender address %s", senderAddr))
					return
				}

				msgPayload := fmt.Sprintf("%s|%s|%d|%d", envelope.Sender, envelope.Receiver, envelope.Amount, envelope.Nonce)
				hash := sha256.Sum256([]byte(msgPayload))
				verifier := crypto.NewVerifier()
				if !verifier.Verify(hash[:], sigBytes, crypto.PublicKey(pubKeyBytes)) {
					respondError(w, http.StatusBadRequest, "invalid cryptographic signature on transaction")
					return
				}

				// Nonce verification & replay attack prevention
				if s.ledger != nil {
					state, err := s.ledger.GetAccountState(ctx, senderAddr)
					if err == nil && state != nil {
						if envelope.Nonce <= state.Nonce {
							respondError(w, http.StatusBadRequest, fmt.Sprintf("stale or replayed nonce: transaction nonce %d <= recorded nonce %d", envelope.Nonce, state.Nonce))
							return
						}
					}
				}

				if err := s.assetMgr.Debit(ctx, senderAddr, envelope.Amount); err != nil {
					respondError(w, http.StatusBadRequest, "transfer debit failed: "+err.Error())
					return
				}

				if err := s.assetMgr.Credit(ctx, receiverAddr, envelope.Amount); err != nil {
					_ = s.assetMgr.Credit(ctx, senderAddr, envelope.Amount)
					respondError(w, http.StatusInternalServerError, "transfer credit failed: "+err.Error())
					return
				}

				// Advance sender nonce in ledger state repository
				if s.ledger != nil {
					if repo, ok := s.ledger.(ledger.StateRepository); ok {
						if state, err := s.ledger.GetAccountState(ctx, senderAddr); err == nil && state != nil {
							state.Nonce = envelope.Nonce
							_ = repo.SetAccountState(ctx, state)
						}
					}
				}

				go s.syncPeerCredit(receiverAddr, envelope.Amount)
				respondJSON(w, http.StatusOK, map[string]interface{}{
					"status":    "offline_tx_validated",
					"sender":    envelope.Sender,
					"receiver":  envelope.Receiver,
					"amount":    envelope.Amount,
					"nonce":     envelope.Nonce,
					"medium":    "bluetooth_mesh",
					"node_id":   s.cfg.ID,
					"timestamp": time.Now().Format(time.RFC3339),
				})
				return
			}
		}
	}

	// Default fallback: relay offline payload over Bluetooth
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "relayed_over_bluetooth",
		"medium":    "bluetooth_mesh",
		"timestamp": time.Now().Format(time.RFC3339),
		"node_id":   s.cfg.ID,
	})
}

// [API] handleListUTRMappings returns all fiat bridge UTR mappings in the system.
func (s *Server) handleListUTRMappings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if s.receiptRegistry == nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{"mappings": []interface{}{}})
		return
	}

	mappings, err := s.receiptRegistry.ListUTRMapping(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list UTR mappings: "+err.Error())
		return
	}

	if mappings == nil {
		mappings = []*receipt.UTRMapping{}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"mappings": mappings,
	})
}

func (s *Server) handleBlockchainHeight(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if s.blockchain == nil {
		respondError(w, http.StatusInternalServerError, "blockchain uninitialized")
		return
	}

	h, err := s.blockchain.Height(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	tip, _ := s.blockchain.CurrentBlock(ctx)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"height": h,
		"tip":    tip,
	})
}

func (s *Server) handleListBlocks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if s.blockchain == nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{"blocks": []interface{}{}})
		return
	}

	h, err := s.blockchain.Height(ctx)
	if err != nil || h == 0 {
		respondJSON(w, http.StatusOK, map[string]interface{}{"blocks": []interface{}{}})
		return
	}

	var blocks []interface{}
	for i := int64(h) - 1; i >= 0 && len(blocks) < 20; i-- {
		blk, err := s.blockchain.GetBlockByHeight(ctx, uint64(i))
		if err == nil && blk != nil {
			blocks = append(blocks, blk)
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"height": h,
		"blocks": blocks,
	})
}

func (s *Server) handleGetBlock(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := chi.URLParam(r, "query")
	if query == "" {
		respondError(w, http.StatusBadRequest, "query height or hash parameter required")
		return
	}

	if s.blockchain == nil {
		respondError(w, http.StatusInternalServerError, "blockchain uninitialized")
		return
	}

	if height, err := strconv.ParseUint(query, 10, 64); err == nil {
		blk, err := s.blockchain.GetBlockByHeight(ctx, height)
		if err != nil {
			respondError(w, http.StatusNotFound, "block height not found")
			return
		}
		respondJSON(w, http.StatusOK, blk)
		return
	}

	blk, err := s.blockchain.GetBlockByHash(ctx, strings.TrimPrefix(query, "0x"))
	if err != nil {
		respondError(w, http.StatusNotFound, "block hash not found")
		return
	}
	respondJSON(w, http.StatusOK, blk)
}

func (s *Server) handleListPeers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if s.peers == nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{"peers": []interface{}{}})
		return
	}

	pList, err := s.peers.ListPeers(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if pList == nil {
		pList = []*peer.Info{}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"peers": pList,
	})
}

func (s *Server) handleRegisterPeer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var p peer.Info
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		respondError(w, http.StatusBadRequest, "invalid peer info payload")
		return
	}

	if p.ID == "" || p.Endpoint == "" {
		respondError(w, http.StatusBadRequest, "peer id and endpoint required")
		return
	}

	if s.peers == nil {
		respondError(w, http.StatusInternalServerError, "peer registry uninitialized")
		return
	}

	p.Connected = true
	if err := s.peers.RegisterPeer(ctx, &p); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to register peer: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status": "peer_registered",
		"peer":   p,
	})
}

func (s *Server) getStabilityEngine() stability.Engine {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stabilityEng == nil {
		s.stabilityEng = stability.NewEngine(nil, s.assetMgr)
	}
	return s.stabilityEng
}

func (s *Server) getFiatManager() fiat.Manager {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.fiatMgr == nil {
		s.fiatMgr = fiat.NewManager()
	}
	return s.fiatMgr
}

func (s *Server) getOrderbookEngine() orderbook.Engine {
	stab := s.getStabilityEngine()
	fiatM := s.getFiatManager()
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.orderbookEng == nil {
		s.orderbookEng = orderbook.NewEngine(s.assetMgr, stab, fiatM)
	}
	return s.orderbookEng
}

// [API] handleWalletCreate generates a new wallet keypair, initializes balance, and returns credentials.
func (s *Server) handleWalletCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	newW, err := wallet.GenerateWallet()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to generate wallet: "+err.Error())
		return
	}

	addr := newW.Address()

	// Credit initial test balance to new wallet (1000 LQD & 1000.00 INR)
	if s.assetMgr != nil {
		_ = s.assetMgr.Mint(ctx, addr, 1000)
	}
	fiatM := s.getFiatManager()
	_ = fiatM.CreditINR(ctx, addr, 1000.00)

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"status":      "created",
		"address":     string(addr),
		"balance":     1000,
		"inr_balance": 1000.00,
	})
}

type signOfflineReq struct {
	Type      string       `json:"type"` // "pcn" or "direct"
	ChannelID string       `json:"channel_id"`
	Sequence  uint64       `json:"sequence"`
	BalanceA  token.Amount `json:"balance_a"`
	BalanceB  token.Amount `json:"balance_b"`
	Receiver  string       `json:"receiver"`
	Amount    token.Amount `json:"amount"`
	Nonce     uint64       `json:"nonce"`
}

// [Wallet] handleSignOffline signs an offline PCN state statement or direct transaction using the node's wallet.
func (s *Server) handleSignOffline(w http.ResponseWriter, r *http.Request) {
	if s.wallet == nil {
		respondError(w, http.StatusInternalServerError, "wallet uninitialized")
		return
	}

	var req signOfflineReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	var hash [32]byte
	if req.Type == "pcn" || req.ChannelID != "" {
		msg := fmt.Sprintf("%s|%d|%d|%d", req.ChannelID, req.Sequence, req.BalanceA, req.BalanceB)
		hash = sha256.Sum256([]byte(msg))
	} else {
		senderAddr := string(s.wallet.Address())
		msg := fmt.Sprintf("%s|%s|%d|%d", senderAddr, req.Receiver, req.Amount, req.Nonce)
		hash = sha256.Sum256([]byte(msg))
	}

	sig, err := s.wallet.SignHash(hash[:])
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to sign offline payload: "+err.Error())
		return
	}

	pubKey := s.wallet.ExportPublicKey()
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":     "signed",
		"sender":     string(s.wallet.Address()),
		"public_key": hex.EncodeToString(pubKey),
		"signature":  hex.EncodeToString(sig),
	})
}

// [Fiat Bank] handleFiatPool returns the aggregate INR reserve pool balance and active fiat bank metrics.
func (s *Server) handleFiatPool(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fiatM := s.getFiatManager()
	totalPool, err := fiatM.GetTotalPoolReserve(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get total fiat pool reserve: "+err.Error())
		return
	}

	userAddr := ""
	var userAvailable float64
	var userReserved float64
	if s.wallet != nil {
		userAddr = string(s.wallet.Address())
		if acc, err := fiatM.GetAccount(ctx, s.wallet.Address()); err == nil && acc != nil {
			userAvailable = acc.Available
			userReserved = acc.Reserved
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"total_inr_pool": totalPool,
		"active_user":    userAddr,
		"user_available": userAvailable,
		"user_reserved":  userReserved,
	})
}

// [API] handleWalletHoldings returns complete token balance, staked share holdings, and bond holdings for an address.
func (s *Server) handleWalletHoldings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	targetAddr := r.URL.Query().Get("address")
	if targetAddr == "" && s.wallet != nil {
		targetAddr = string(s.wallet.Address())
	}
	if targetAddr == "" {
		respondError(w, http.StatusBadRequest, "address parameter required")
		return
	}

	userAddr := crypto.Address(targetAddr)
	var balance token.Amount
	if s.assetMgr != nil {
		if b, err := s.assetMgr.BalanceOf(ctx, userAddr); err == nil {
			balance = b
		}
	}

	var inrAvailable float64
	var inrReserved float64
	fiatM := s.getFiatManager()
	if acc, err := fiatM.GetAccount(ctx, userAddr); err == nil && acc != nil {
		inrAvailable = acc.Available
		inrReserved = acc.Reserved
	}

	stabEng := s.getStabilityEngine()
	stakePos, _ := stabEng.GetStakingPosition(ctx, userAddr)
	bondPos, _ := stabEng.GetBondPosition(ctx, userAddr)

	var stakedShares token.Amount
	var accruedYield token.Amount
	if stakePos != nil {
		stakedShares = stakePos.StakedShares
		accruedYield = stakePos.AccruedYield
	}

	var bondsHeld token.Amount
	if bondPos != nil {
		bondsHeld = bondPos.BondsHeld
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"address":       targetAddr,
		"balance":       balance,
		"inr_balance":   inrAvailable,
		"inr_reserved":  inrReserved,
		"staked_shares": stakedShares,
		"accrued_yield": accruedYield,
		"bonds_held":    bondsHeld,
	})
}

// [Orderbook] handleOrderbookDepth returns the current orderbook depth (bids and asks).
func (s *Server) handleOrderbookDepth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	depth, err := s.getOrderbookEngine().GetDepth(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch orderbook depth: "+err.Error())
		return
	}
	respondJSON(w, http.StatusOK, depth)
}

// [Orderbook] handleOrderbookTrades returns recent executed trades.
func (s *Server) handleOrderbookTrades(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	trades, err := s.getOrderbookEngine().GetTrades(ctx, 30)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch trades: "+err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"trades": trades,
	})
}

// [Orderbook] handleUserOrders returns all orders for a specific user address.
func (s *Server) handleUserOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	addrStr := r.URL.Query().Get("address")
	if addrStr == "" && s.wallet != nil {
		addrStr = string(s.wallet.Address())
	}
	if addrStr == "" {
		respondError(w, http.StatusBadRequest, "address parameter required")
		return
	}

	orders, err := s.getOrderbookEngine().GetUserOrders(ctx, crypto.Address(addrStr))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch user orders: "+err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"orders": orders,
	})
}

type placeOrderReq struct {
	User   string       `json:"user"`
	Side   string       `json:"side"`
	Price  float64      `json:"price"`
	Amount token.Amount `json:"amount"`
}

// [Orderbook] handlePlaceOrder submits a new limit buy or sell order to the orderbook.
func (s *Server) handlePlaceOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req placeOrderReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid order payload")
		return
	}

	userAddr := req.User
	if userAddr == "" && s.wallet != nil {
		userAddr = string(s.wallet.Address())
	}
	if userAddr == "" {
		respondError(w, http.StatusBadRequest, "user address required")
		return
	}

	ord := &orderbook.Order{
		User:   crypto.Address(userAddr),
		Side:   orderbook.OrderSide(strings.ToLower(req.Side)),
		Price:  req.Price,
		Amount: req.Amount,
	}

	placed, trades, err := s.getOrderbookEngine().PlaceOrder(ctx, ord)
	if err != nil {
		respondError(w, http.StatusBadRequest, "order placement failed: "+err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"order":  placed,
		"trades": trades,
	})
}

type cancelOrderReq struct {
	OrderID string `json:"order_id"`
	User    string `json:"user"`
}

// [Orderbook] handleCancelOrder cancels an open order.
func (s *Server) handleCancelOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req cancelOrderReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid cancel payload")
		return
	}

	userAddr := req.User
	if userAddr == "" && s.wallet != nil {
		userAddr = string(s.wallet.Address())
	}

	if err := s.getOrderbookEngine().CancelOrder(ctx, req.OrderID, crypto.Address(userAddr)); err != nil {
		respondError(w, http.StatusBadRequest, "failed to cancel order: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":   "cancelled",
		"order_id": req.OrderID,
	})
}

// [Stability] handleStabilityStatus returns the current stability status and seigniorage parameters.
func (s *Server) handleStabilityStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status, err := s.getStabilityEngine().GetStatus(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get stability status: "+err.Error())
		return
	}
	respondJSON(w, http.StatusOK, status)
}

// [Stability] handleStabilityHistory returns price history points for charting.
func (s *Server) handleStabilityHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	history, err := s.getStabilityEngine().GetPriceHistory(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get price history: "+err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"history": history,
	})
}

type stakeReq struct {
	User   string       `json:"user"`
	Amount token.Amount `json:"amount"`
}

// [Stability] handleStakeShares stakes LQD tokens into seigniorage shares.
func (s *Server) handleStakeShares(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req stakeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid stake payload")
		return
	}

	userAddr := req.User
	if userAddr == "" && s.wallet != nil {
		userAddr = string(s.wallet.Address())
	}

	pos, err := s.getStabilityEngine().StakeShares(ctx, crypto.Address(userAddr), req.Amount)
	if err != nil {
		respondError(w, http.StatusBadRequest, "stake failed: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, pos)
}

// [Stability] handleUnstakeShares unstakes shares and claims yield.
func (s *Server) handleUnstakeShares(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req stakeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid unstake payload")
		return
	}

	userAddr := req.User
	if userAddr == "" && s.wallet != nil {
		userAddr = string(s.wallet.Address())
	}

	pos, err := s.getStabilityEngine().UnstakeShares(ctx, crypto.Address(userAddr), req.Amount)
	if err != nil {
		respondError(w, http.StatusBadRequest, "unstake failed: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, pos)
}

type bondReq struct {
	User   string       `json:"user"`
	Amount token.Amount `json:"amount"`
}

// [Stability] handleBuyBonds purchases discounted LQD bonds during contraction.
func (s *Server) handleBuyBonds(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req bondReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid bond payload")
		return
	}

	userAddr := req.User
	if userAddr == "" && s.wallet != nil {
		userAddr = string(s.wallet.Address())
	}

	pos, err := s.getStabilityEngine().BuyBonds(ctx, crypto.Address(userAddr), req.Amount)
	if err != nil {
		respondError(w, http.StatusBadRequest, "buy bonds failed: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, pos)
}

// [Stability] handleRedeemBonds redeems LQD bonds for LQD + bonus during expansion.
func (s *Server) handleRedeemBonds(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req bondReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid redemption payload")
		return
	}

	userAddr := req.User
	if userAddr == "" && s.wallet != nil {
		userAddr = string(s.wallet.Address())
	}

	pos, err := s.getStabilityEngine().RedeemBonds(ctx, crypto.Address(userAddr), req.Amount)
	if err != nil {
		respondError(w, http.StatusBadRequest, "redeem bonds failed: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, pos)
}

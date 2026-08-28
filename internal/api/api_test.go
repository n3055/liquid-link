package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

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
	"github.com/n3055/liquid-link/internal/wallet"
)

func setupTestServer(t *testing.T) (*Server, crypto.Address, func()) {
	tmpDir := t.TempDir()
	cfg := &config.StorageConfig{Path: tmpDir, ValueLogFileSize: 1024 * 1024}
	store, err := storage.NewBadgerStore(cfg)
	if err != nil {
		t.Fatalf("failed to open Badger store: %v", err)
	}

	stateRepo := ledger.NewStateRepository(store)
	led := ledger.NewLedger(stateRepo)
	balanceAdapter := ledger.NewLedgerBalanceAdapter(stateRepo)

	assetMgr := token.NewAssetManager(balanceAdapter, store)

	cryptoGen := crypto.NewGenerator()
	kp, err := cryptoGen.GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	signer, err := crypto.NewSigner(kp.PrivateKey)
	if err != nil {
		t.Fatalf("failed to create signer: %v", err)
	}

	verifier := crypto.NewVerifier()
	w := wallet.NewWallet(kp)

	ctx := context.Background()
	_ = assetMgr.Mint(ctx, w.Address(), 1000)

	blockRepo := blockchain.NewRepository(store)
	nodeID := "test-node-1"
	validators := map[string]crypto.PublicKey{
		nodeID: kp.PublicKey,
	}

	bc := blockchain.NewBlockchain(blockRepo, verifier, validators)
	cons, err := consensus.NewEngine(nodeID, kp.PrivateKey, bc, led, validators, 0.5)
	if err != nil {
		t.Fatalf("failed to create consensus engine: %v", err)
	}

	pcnRegistry := pcn.NewHolderRegistry(store)
	pcnMgr := pcn.NewManager(w.Address(), signer, led, pcnRegistry, assetMgr)

	receiptReg := receipt.NewRegistry(store, verifier, kp.PublicKey)
	br := receipt.NewBridge(receiptReg, assetMgr, signer, store)

	peerReg := peer.NewRegistry(store)
	_ = peerReg.RegisterPeer(ctx, &peer.Info{
		ID:        nodeID,
		Address:   w.Address(),
		Endpoint:  "http://localhost:8080",
		Connected: true,
	})

	nodeCfg := &config.NodeConfig{
		ID:   nodeID,
		Port: 8080,
	}

	s := NewServer(
		nodeCfg,
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

	cleanup := func() {
		store.Close()
	}

	return s, w.Address(), cleanup
}

func TestStatusEndpoint(t *testing.T) {
	s, _, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/status", nil)
	rr := httptest.NewRecorder()

	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var res map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	if res["status"] != "online" {
		t.Errorf("expected status 'online', got %v", res["status"])
	}
	if res["node_id"] != "test-node-1" {
		t.Errorf("expected node_id 'test-node-1', got %v", res["node_id"])
	}
}

func TestWalletEndpoints(t *testing.T) {
	s, _, cleanup := setupTestServer(t)
	defer cleanup()

	// 1. Info
	req := httptest.NewRequest(http.MethodGet, "/api/v1/wallet/info", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for wallet info, got %d", rr.Code)
	}

	// 2. Mint
	mintPayload := []byte(`{"address":"receiver-1","amount":500}`)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/wallet/mint", bytes.NewBuffer(mintPayload))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for mint, got %d", rr.Code)
	}

	// 3. Balance check
	req = httptest.NewRequest(http.MethodGet, "/api/v1/wallet/balance?address=receiver-1", nil)
	rr = httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for balance, got %d", rr.Code)
	}

	var balRes map[string]interface{}
	_ = json.Unmarshal(rr.Body.Bytes(), &balRes)
	if balRes["balance"] != float64(500) {
		t.Errorf("expected balance 500, got %v", balRes["balance"])
	}

	// 4. Transfer
	transferPayload := []byte(`{"receiver":"receiver-2","amount":100}`)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/wallet/transfer", bytes.NewBuffer(transferPayload))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202 Accepted for transfer, got %d", rr.Code)
	}
}

func TestStateChannelEndpoints(t *testing.T) {
	s, _, cleanup := setupTestServer(t)
	defer cleanup()

	// 1. Open channel
	openBody := []byte(`{"counterparty":"party-b-address","amount":200}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/channels/open", bytes.NewBuffer(openBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for open channel, got %d: %s", rr.Code, rr.Body.String())
	}

	var chanObj pcn.Channel
	if err := json.Unmarshal(rr.Body.Bytes(), &chanObj); err != nil {
		t.Fatalf("failed to unmarshal channel: %v", err)
	}
	chanID := chanObj.ID

	// 2. List channels
	req = httptest.NewRequest(http.MethodGet, "/api/v1/channels/list", nil)
	rr = httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for channel list, got %d", rr.Code)
	}

	// 3. Get channel by ID
	req = httptest.NewRequest(http.MethodGet, "/api/v1/channels/"+chanID, nil)
	rr = httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for get channel, got %d", rr.Code)
	}

	// 4. Update channel
	updateBody := []byte(fmt.Sprintf(`{"channel_id":"%s","amount":50,"sending":true}`, chanID))
	req = httptest.NewRequest(http.MethodPost, "/api/v1/channels/update", bytes.NewBuffer(updateBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for update channel, got %d: %s", rr.Code, rr.Body.String())
	}

	// 5. Settle channel
	settleBody := []byte(fmt.Sprintf(`{"channel_id":"%s","action":"initiate"}`, chanID))
	req = httptest.NewRequest(http.MethodPost, "/api/v1/channels/settle", bytes.NewBuffer(settleBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for initiate settlement, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestFiatBridgeEndpoints(t *testing.T) {
	s, wAddr, cleanup := setupTestServer(t)
	defer cleanup()

	// 1. Initiate Bridge Request
	reqBody := []byte(fmt.Sprintf(`{"receiver":"%s","amount":150}`, string(wAddr)))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/bridge/fiat-request", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202 Accepted for bridge request, got %d: %s", rr.Code, rr.Body.String())
	}

	var res map[string]interface{}
	_ = json.Unmarshal(rr.Body.Bytes(), &res)
	mapping := res["mapping"].(map[string]interface{})
	utr := mapping["utr"].(string)

	// 2. Query UTR Mapping
	req = httptest.NewRequest(http.MethodGet, "/api/v1/bridge/utr/"+utr, nil)
	rr = httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for UTR query, got %d", rr.Code)
	}

	// 3. Verify Receipt
	verifyBody := []byte(fmt.Sprintf(`{"utr":"%s","receipt_id":"rec-123"}`, utr))
	req = httptest.NewRequest(http.MethodPost, "/api/v1/bridge/verify-receipt", bytes.NewBuffer(verifyBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for verify receipt, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestPeerEndpoints(t *testing.T) {
	s, _, cleanup := setupTestServer(t)
	defer cleanup()

	// List peers
	req := httptest.NewRequest(http.MethodGet, "/api/v1/peers", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for list peers, got %d", rr.Code)
	}

	// Register new peer
	regBody := []byte(`{"id":"peer-node-2","endpoint":"http://localhost:8081","address":"0xpeer2"}`)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/peers/register", bytes.NewBuffer(regBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for peer registration, got %d", rr.Code)
	}
}

func TestBlockchainEndpoints(t *testing.T) {
	s, _, cleanup := setupTestServer(t)
	defer cleanup()

	// Height
	req := httptest.NewRequest(http.MethodGet, "/api/v1/blockchain/height", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for blockchain height, got %d", rr.Code)
	}

	// Blocks
	req = httptest.NewRequest(http.MethodGet, "/api/v1/blockchain/blocks", nil)
	rr = httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for list blocks, got %d", rr.Code)
	}
}

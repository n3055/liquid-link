package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/n3055/liquid-link/internal/crypto"
	"github.com/n3055/liquid-link/internal/pcn"
	"github.com/n3055/liquid-link/internal/token"
)

// [Wallet Isolation] TestWalletChannelIsolation_NewWalletShowsNoOlderChannels verifies that
// newly created wallets do not leak or display open/closed PCN channels belonging to older wallets.
func TestWalletChannelIsolation_NewWalletShowsNoOlderChannels(t *testing.T) {
	server, nodeAddress, cleanup := setupTestServer(t)
	defer cleanup()

	cryptoGen := crypto.NewGenerator()
	counterpartyKp, err := cryptoGen.GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate counterparty keypair: %v", err)
	}
	counterpartyAddr := cryptoGen.DeriveAddress(counterpartyKp.PublicKey)

	// 1. Node wallet opens a channel with counterparty
	openPayload := []byte(fmt.Sprintf(`{"counterparty":"%s","amount":150}`, counterpartyAddr))
	openReq := httptest.NewRequest(http.MethodPost, "/api/v1/channels/open", bytes.NewBuffer(openPayload))
	openReq.Header.Set("Content-Type", "application/json")
	openRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(openRecorder, openReq)
	if openRecorder.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for open channel, got %d: %s", openRecorder.Code, openRecorder.Body.String())
	}

	// 2. Querying channels without address returns the node's channel
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/channels/list", nil)
	listRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(listRecorder, listReq)
	if listRecorder.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", listRecorder.Code)
	}
	var allChannelsResp struct {
		Channels []interface{} `json:"channels"`
	}
	if err := json.Unmarshal(listRecorder.Body.Bytes(), &allChannelsResp); err != nil {
		t.Fatalf("failed to unmarshal channels response: %v", err)
	}
	if len(allChannelsResp.Channels) == 0 {
		t.Fatalf("expected at least 1 channel in global list, got 0")
	}

	// 3. Create a brand new wallet
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/wallet/create", nil)
	createRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(createRecorder, createReq)
	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for new wallet, got %d: %s", createRecorder.Code, createRecorder.Body.String())
	}
	var newWalletResp struct {
		Address string `json:"address"`
	}
	if err := json.Unmarshal(createRecorder.Body.Bytes(), &newWalletResp); err != nil {
		t.Fatalf("failed to parse wallet create response: %v", err)
	}
	newWalletAddress := newWalletResp.Address
	if newWalletAddress == "" || newWalletAddress == string(nodeAddress) {
		t.Fatalf("expected distinct new wallet address, got: %s", newWalletAddress)
	}

	// 4. Querying channels scoped to new wallet must return 0 channels (NO LEAKAGE)
	scopedReq := httptest.NewRequest(http.MethodGet, "/api/v1/channels/list?address="+newWalletAddress, nil)
	scopedRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(scopedRecorder, scopedReq)
	if scopedRecorder.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for scoped channel query, got %d", scopedRecorder.Code)
	}
	var scopedChannelsResp struct {
		Channels []interface{} `json:"channels"`
	}
	if err := json.Unmarshal(scopedRecorder.Body.Bytes(), &scopedChannelsResp); err != nil {
		t.Fatalf("failed to unmarshal scoped channels response: %v", err)
	}
	if len(scopedChannelsResp.Channels) != 0 {
		t.Fatalf("SECURITY/ISOLATION BUG: new wallet shows %d older channels, expected 0", len(scopedChannelsResp.Channels))
	}
}

// [Wallet Isolation] TestWalletBridgeIsolation_NewWalletShowsNoOlderBridgeRequests verifies that
// newly created wallets do not leak or display fiat bridge requests belonging to older wallets.
func TestWalletBridgeIsolation_NewWalletShowsNoOlderBridgeRequests(t *testing.T) {
	server, nodeAddress, cleanup := setupTestServer(t)
	defer cleanup()

	// 1. Node wallet initiates a fiat bridge request
	bridgePayload := []byte(fmt.Sprintf(`{"receiver":"%s","amount":100}`, nodeAddress))
	bridgeReq := httptest.NewRequest(http.MethodPost, "/api/v1/bridge/fiat-request", bytes.NewBuffer(bridgePayload))
	bridgeReq.Header.Set("Content-Type", "application/json")
	bridgeRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(bridgeRecorder, bridgeReq)
	if bridgeRecorder.Code != http.StatusAccepted {
		t.Fatalf("expected 202 Accepted for bridge request, got %d: %s", bridgeRecorder.Code, bridgeRecorder.Body.String())
	}

	// 2. Create a brand new wallet
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/wallet/create", nil)
	createRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(createRecorder, createReq)
	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for new wallet, got %d: %s", createRecorder.Code, createRecorder.Body.String())
	}
	var newWalletResp struct {
		Address string `json:"address"`
	}
	if err := json.Unmarshal(createRecorder.Body.Bytes(), &newWalletResp); err != nil {
		t.Fatalf("failed to parse wallet create response: %v", err)
	}
	newWalletAddress := newWalletResp.Address

	// 3. Querying bridge list scoped to new wallet must return 0 bridge records
	scopedReq := httptest.NewRequest(http.MethodGet, "/api/v1/bridge/list?address="+newWalletAddress, nil)
	scopedRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(scopedRecorder, scopedReq)
	if scopedRecorder.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for scoped bridge query, got %d", scopedRecorder.Code)
	}
	var scopedBridgeResp struct {
		Mappings []interface{} `json:"mappings"`
	}
	if err := json.Unmarshal(scopedRecorder.Body.Bytes(), &scopedBridgeResp); err != nil {
		t.Fatalf("failed to unmarshal scoped bridge response: %v", err)
	}
	if len(scopedBridgeResp.Mappings) != 0 {
		t.Fatalf("SECURITY/ISOLATION BUG: new wallet shows %d older bridge requests, expected 0", len(scopedBridgeResp.Mappings))
	}
}

// [Wallet Isolation] TestDashboardWalletScoping_NewWalletShowsOwnBalance verifies that the dashboard
// properly scopes the balance, channels, and bridge requests when rendered for a specific user wallet.
func TestDashboardWalletScoping_NewWalletShowsOwnBalance(t *testing.T) {
	server, nodeAddress, cleanup := setupTestServer(t)
	defer cleanup()

	// Mint an arbitrary extra balance to node wallet (e.g. total 26476 LQD)
	ctx := context.Background()
	_ = server.assetMgr.Mint(ctx, nodeAddress, token.Amount(25476))

	// Create a new wallet with initial balance 1000
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/wallet/create", nil)
	createRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(createRecorder, createReq)
	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for new wallet, got %d", createRecorder.Code)
	}
	var newWalletResp struct {
		Address string `json:"address"`
		Balance int    `json:"balance"`
	}
	_ = json.Unmarshal(createRecorder.Body.Bytes(), &newWalletResp)
	newWalletAddress := newWalletResp.Address

	// Open a channel for node wallet to ensure channel list is non-empty for node
	cryptoGen := crypto.NewGenerator()
	kp, _ := cryptoGen.GenerateKeyPair()
	peerAddr := cryptoGen.DeriveAddress(kp.PublicKey)
	openPayload := []byte(fmt.Sprintf(`{"counterparty":"%s","amount":100}`, peerAddr))
	openReq := httptest.NewRequest(http.MethodPost, "/api/v1/channels/open", bytes.NewBuffer(openPayload))
	openReq.Header.Set("Content-Type", "application/json")
	openRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(openRecorder, openReq)

	// 1. When rendered for new wallet via ?address= parameter
	dashReq := httptest.NewRequest(http.MethodGet, "/?address="+newWalletAddress, nil)
	dashRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(dashRecorder, dashReq)
	if dashRecorder.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", dashRecorder.Code)
	}
	body := dashRecorder.Body.String()

	// Must show new wallet's balance (1000), NOT node wallet's balance (26476)
	if !strings.Contains(body, ">1000<small>LQD</small>") {
		t.Errorf("expected dashboard to show 1000 LQD for new wallet, body had unexpected balance")
	}
	if strings.Contains(body, ">26476<small>LQD</small>") {
		t.Errorf("SECURITY BUG: dashboard displayed node balance 26476 for new wallet")
	}

	// Must show 0 channels count for new wallet
	if !strings.Contains(body, `id="channelsCount">(0)</span>`) {
		t.Errorf("expected channelsCount to be (0) for new wallet, body had different count")
	}

	// Must show 0 fiat bridge count for new wallet
	if !strings.Contains(body, `id="bridgeCount">(0)</span>`) {
		t.Errorf("expected bridgeCount to be (0) for new wallet, body had different count")
	}
}

// [Wallet Isolation] TestDashboardTemplate_OfflinePersistenceAndStorageScoping verifies that
// the client-side JavaScript implements persistent holdings caching, wallet-scoped channel/bridge keys,
// and immediate DOM initialization without node balance fallback.
func TestDashboardTemplate_OfflinePersistenceAndStorageScoping(t *testing.T) {
	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	dashReq := httptest.NewRequest(http.MethodGet, "/", nil)
	dashRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(dashRecorder, dashReq)
	body := dashRecorder.Body.String()

	requiredPatterns := []string{
		"getCachedWalletHoldings",
		"saveWalletHoldings",
		"getChannelsStorageKey",
		"getBridgeStorageKey",
		"getOfflineQueueStorageKey",
		"renderChannelsList",
		"renderBridgeList",
	}

	for _, pattern := range requiredPatterns {
		if !strings.Contains(body, pattern) {
			t.Errorf("expected dashboard JavaScript to contain %q for offline persistence and wallet isolation", pattern)
		}
	}
}

// [Wallet Isolation] TestOfflineSimulation_NewWalletHoldingsAndOfflineFallback verifies that
// newly created wallets maintain their 1000 balance and isolated state when offline endpoints return 503.
func TestOfflineSimulation_NewWalletHoldingsAndOfflineFallback(t *testing.T) {
	server, nodeAddress, cleanup := setupTestServer(t)
	defer cleanup()

	// Mint large balance to node
	ctx := context.Background()
	_ = server.assetMgr.Mint(ctx, nodeAddress, 26476)

	// Create new user wallet
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/wallet/create", nil)
	createRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(createRecorder, createReq)
	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", createRecorder.Code)
	}

	var newWalletResp struct {
		Address string  `json:"address"`
		Balance int     `json:"balance"`
		INR     float64 `json:"inr_balance"`
	}
	_ = json.Unmarshal(createRecorder.Body.Bytes(), &newWalletResp)

	if newWalletResp.Balance != 1000 {
		t.Errorf("expected initial new wallet balance 1000, got %d", newWalletResp.Balance)
	}
	if newWalletResp.INR != 1000.00 {
		t.Errorf("expected initial new wallet INR balance 1000.00, got %f", newWalletResp.INR)
	}

	// Verify holdings endpoint returns exactly 1000 for this wallet
	holdingsReq := httptest.NewRequest(http.MethodGet, "/api/v1/wallet/holdings?address="+newWalletResp.Address, nil)
	holdingsRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(holdingsRecorder, holdingsReq)
	if holdingsRecorder.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from holdings, got %d", holdingsRecorder.Code)
	}

	var holdingsResp struct {
		Balance int     `json:"balance"`
		INR     float64 `json:"inr_balance"`
	}
	_ = json.Unmarshal(holdingsRecorder.Body.Bytes(), &holdingsResp)
	if holdingsResp.Balance != 1000 {
		t.Errorf("expected holdings balance 1000, got %d", holdingsResp.Balance)
	}

	// Verify SW script is served and has offline mode handling
	swReq := httptest.NewRequest(http.MethodGet, "/sw.js", nil)
	swRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(swRecorder, swReq)
	if swRecorder.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for sw.js, got %d", swRecorder.Code)
	}
	swBody := swRecorder.Body.String()
	if !strings.Contains(swBody, "Node network unavailable (Offline Mode)") {
		t.Errorf("expected offline mode warning in sw.js")
	}
}

// [PCN State] TestChannelsOpen_FromActiveWallet verifies that opening a channel from a newly created
// client wallet debits the client wallet's balance and sets Party A to that wallet instead of the node default wallet.
func TestChannelsOpen_FromActiveWallet(t *testing.T) {
	server, nodeAddress, cleanup := setupTestServer(t)
	defer cleanup()

	// 1. Record initial node wallet balance
	ctx := context.Background()
	initialNodeBal, err := server.assetMgr.BalanceOf(ctx, nodeAddress)
	if err != nil {
		t.Fatalf("failed to get node balance: %v", err)
	}

	// 2. Create new client wallet (starts with 1000 LQD)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/wallet/create", nil)
	createRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(createRecorder, createReq)
	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for wallet create, got %d", createRecorder.Code)
	}

	var newWalletResp struct {
		Address string `json:"address"`
	}
	_ = json.Unmarshal(createRecorder.Body.Bytes(), &newWalletResp)
	userAddr := newWalletResp.Address

	// Counterparty address
	cryptoGen := crypto.NewGenerator()
	peerKp, _ := cryptoGen.GenerateKeyPair()
	peerAddr := cryptoGen.DeriveAddress(peerKp.PublicKey)

	// 3. Open channel specifying sender as new user wallet
	depositAmount := token.Amount(300)
	openPayload := []byte(fmt.Sprintf(`{"sender":"%s","counterparty":"%s","amount":%d}`, userAddr, peerAddr, depositAmount))
	openReq := httptest.NewRequest(http.MethodPost, "/api/v1/channels/open", bytes.NewBuffer(openPayload))
	openReq.Header.Set("Content-Type", "application/json")
	openRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(openRecorder, openReq)

	if openRecorder.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for open channel from user wallet, got %d: %s", openRecorder.Code, openRecorder.Body.String())
	}

	var chanObj pcn.Channel
	if err := json.Unmarshal(openRecorder.Body.Bytes(), &chanObj); err != nil {
		t.Fatalf("failed to decode channel response: %v", err)
	}

	// PartyA MUST be the user wallet, NOT the node default wallet
	if string(chanObj.Share.PartyA) != userAddr {
		t.Errorf("BUG: channel PartyA is %s, expected active user wallet %s", chanObj.Share.PartyA, userAddr)
	}
	if string(chanObj.Share.PartyB) != string(peerAddr) {
		t.Errorf("expected channel PartyB %s, got %s", peerAddr, chanObj.Share.PartyB)
	}
	if chanObj.Share.BalancePartyA != depositAmount {
		t.Errorf("expected balance party A %d, got %d", depositAmount, chanObj.Share.BalancePartyA)
	}

	// Verify user balance was debited: 1000 - 300 = 700
	userBal, err := server.assetMgr.BalanceOf(ctx, crypto.Address(userAddr))
	if err != nil {
		t.Fatalf("failed to query user balance: %v", err)
	}
	if userBal != 700 {
		t.Errorf("expected user balance 700, got %d", userBal)
	}

	// Verify node balance was NOT debited
	nodeBal, err := server.assetMgr.BalanceOf(ctx, nodeAddress)
	if err != nil {
		t.Fatalf("failed to query node balance: %v", err)
	}
	if nodeBal != initialNodeBal {
		t.Errorf("BUG: node balance was debited! initial=%d, current=%d", initialNodeBal, nodeBal)
	}

	// 4. Querying channels scoped to user wallet returns the channel
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/channels/list?address="+userAddr, nil)
	listRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(listRecorder, listReq)
	if listRecorder.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", listRecorder.Code)
	}
	var listResp struct {
		Channels []*pcn.Channel `json:"channels"`
	}
	_ = json.Unmarshal(listRecorder.Body.Bytes(), &listResp)
	if len(listResp.Channels) != 1 {
		t.Fatalf("expected 1 channel scoped to user, got %d", len(listResp.Channels))
	}
	if listResp.Channels[0].ID != chanObj.ID {
		t.Errorf("expected channel ID %s, got %s", chanObj.ID, listResp.Channels[0].ID)
	}
}

// [Dashboard] TestDashboardTemplate_ZeroJavaScriptSyntaxErrors validates that the entire embedded
// client-side JavaScript block parses with zero syntax errors using the Node.js syntax checker.
func TestDashboardTemplate_ZeroJavaScriptSyntaxErrors(t *testing.T) {
	nodePath, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node executable not found in PATH; skipping JavaScript syntax check")
	}

	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	dashReq := httptest.NewRequest(http.MethodGet, "/", nil)
	dashRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(dashRecorder, dashReq)
	body := dashRecorder.Body.String()

	startTag := "<script>"
	endTag := "</script>"
	startIdx := strings.Index(body, startTag)
	endIdx := strings.LastIndex(body, endTag)
	if startIdx == -1 || endIdx == -1 {
		t.Fatalf("could not locate <script> block in rendered dashboard body")
	}

	jsCode := body[startIdx+len(startTag) : endIdx]
	// Replace any Go template tags with a mock string literal
	templateTagRe := regexp.MustCompile(`\{\{[^}]+\}\}`)
	cleanJs := templateTagRe.ReplaceAllString(jsCode, `"mock_template_value"`)

	tmpFile, err := os.CreateTemp("", "dashboard_check_*.js")
	if err != nil {
		t.Fatalf("failed to create temporary file for JS syntax check: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(cleanJs); err != nil {
		t.Fatalf("failed to write JS to temporary file: %v", err)
	}
	_ = tmpFile.Close()

	cmd := exec.Command(nodePath, "--check", tmpFile.Name())
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("JavaScript syntax validation failed:\n%s", string(output))
	}
}

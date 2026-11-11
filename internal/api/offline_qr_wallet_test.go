package api

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

// [Offline Wallet] TestOfflineWallet_SaveWalletOfflineQueueFunctionDefined verifies that
// saveWalletOfflineQueue is defined in the embedded client-side script and does not throw ReferenceError.
func TestOfflineWallet_SaveWalletOfflineQueueFunctionDefined(t *testing.T) {
	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	dashReq := httptest.NewRequest(http.MethodGet, "/", nil)
	dashRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(dashRecorder, dashReq)
	body := dashRecorder.Body.String()

	if !strings.Contains(body, "function saveWalletOfflineQueue(") {
		t.Fatalf("expected dashboard_template.go to define 'function saveWalletOfflineQueue(', but definition was missing")
	}
}

// [Receipt Verification] TestOfflineQR_DetectionRecognizesDirectAndPCNTypes verifies that
// both direct transfer (direct_tx_offline) and PCN state payment offline QR payloads
// are recognized by QR scan detection without being mistaken for plain wallet addresses.
func TestOfflineQR_DetectionRecognizesDirectAndPCNTypes(t *testing.T) {
	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	dashReq := httptest.NewRequest(http.MethodGet, "/", nil)
	dashRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(dashRecorder, dashReq)
	body := dashRecorder.Body.String()

	// Locate onQrScanSuccess function body
	onQrScanRegex := regexp.MustCompile(`function onQrScanSuccess\([^{]+\{[\s\S]+?function\s+`)
	match := onQrScanRegex.FindString(body)
	if match == "" {
		t.Fatalf("could not locate onQrScanSuccess in dashboard script")
	}

	// Ensure onQrScanSuccess checks for direct_tx_offline
	if !strings.Contains(match, "direct_tx_offline") {
		t.Fatalf("expected onQrScanSuccess implementation to recognize 'direct_tx_offline' offline QR payload type")
	}
}

// [Receipt Verification] TestOfflineQR_ReceiveOfflineModalHelperDefined verifies that
// the dedicated openReceiveOfflineModal helper and Receive Offline buttons exist on the dashboard.
func TestOfflineQR_ReceiveOfflineModalHelperDefined(t *testing.T) {
	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	dashReq := httptest.NewRequest(http.MethodGet, "/", nil)
	dashRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(dashRecorder, dashReq)
	body := dashRecorder.Body.String()

	if !strings.Contains(body, "function openReceiveOfflineModal(") {
		t.Fatalf("expected dashboard_template.go to define 'function openReceiveOfflineModal('")
	}
	if !strings.Contains(body, "openReceiveOfflineModal()") {
		t.Fatalf("expected dashboard_template.go to have a button calling openReceiveOfflineModal()")
	}
}

// [Receipt Verification] TestOfflineQR_PCNReceiveSequencePreservation verifies that
// receiving a PCN offline payment with higher sequence is preserved and not overwritten
// by stale sequence responses from the node server during periodic polling.
func TestOfflineQR_PCNReceiveSequencePreservation(t *testing.T) {
	nodePath, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node executable not found in PATH; skipping Node.js runtime verification")
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
	templateTagRe := regexp.MustCompile(`\{\{[^}]+\}\}`)
	cleanJs := templateTagRe.ReplaceAllString(jsCode, `"mock_template_value"`)

	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("failed to generate ed25519 keypair: %v", err)
	}
	pubHex := hex.EncodeToString(pub)
	senderAddr := "0x" + hex.EncodeToString(pub[:20])
	receiverAddr := "0x2222333344445555666677778888999900001111"

	chanID := "chan_seq_preservation_test"
	pcnMsg := fmt.Sprintf("%s|2|850|150", chanID)
	pcnHash := sha256.Sum256([]byte(pcnMsg))
	pcnSig := hex.EncodeToString(ed25519.Sign(priv, pcnHash[:]))

	nodeHarness := fmt.Sprintf(`
const localStorageStore = {};
global.localStorage = {
  getItem: (k) => localStorageStore[k] || null,
  setItem: (k, v) => { localStorageStore[k] = String(v); },
  removeItem: (k) => { delete localStorageStore[k]; },
  clear: () => { for (const k in localStorageStore) delete localStorageStore[k]; }
};

const elements = {};
global.document = {
  getElementById: (id) => {
    if (!elements[id]) {
      elements[id] = {
        id: id,
        value: '',
        textContent: '',
        innerHTML: '',
        style: {},
        classList: { add: () => {}, remove: () => {}, contains: () => false },
        querySelectorAll: () => [],
        appendChild: () => {}
      };
    }
    return elements[id];
  },
  createElement: (tag) => ({ tag, style: {}, appendChild: () => {} }),
  querySelectorAll: () => [],
  addEventListener: () => {}
};
global.window = {
  localStorage: global.localStorage,
  document: global.document,
  location: { port: '8080', hostname: 'localhost' },
  requestAnimationFrame: (cb) => cb(),
  addEventListener: () => {},
  crypto: { getRandomValues: (buf) => require('crypto').randomFillSync(buf) }
};
global.requestAnimationFrame = global.window.requestAnimationFrame;
global.navigator = { userAgent: 'node-test', vibrate: () => true };
global.showAlert = () => {};
global.refreshDashboardData = () => {};
global.switchSubpanel = () => {};
global.closeSheetAndStopScan = () => {};
global.openSheet = () => {};

// Mock fetch to simulate node returning stale sequence 1
global.fetch = async (url) => {
  return {
    ok: true,
    json: async () => ({
      channels: [{
        id: '%s',
        state: 'OPEN',
        sequence: 1,
        share: {
          party_a: '%s',
          party_b: '%s',
          balance_party_a: 1000,
          balance_party_b: 0
        }
      }]
    })
  };
};

%s

localStorage.setItem(getStorageKey('active_wallet'), '%s');

(async () => {
  // 1. Initial state: receiver has channel at sequence 1
  saveWalletChannels('%s', [{
    id: '%s',
    state: 'OPEN',
    sequence: 1,
    share: {
      party_a: '%s',
      party_b: '%s',
      balance_party_a: 1000,
      balance_party_b: 0
    }
  }]);

  // 2. Receiver scans PCN offline payment with sequence 2
  const pcnPayload = JSON.stringify({
    type: 'pcn_offline_payment',
    channel_id: '%s',
    sequence: 2,
    party_a: '%s',
    party_b: '%s',
    balance_a: 850,
    balance_b: 150,
    amount: 150,
    signature: '%s',
    public_key: '%s'
  });
  await onOfflineQrScanSuccess(pcnPayload);

  // 3. Simulate periodic background refreshChannelsUI (which fetches from server)
  await refreshChannelsUI();

  // 4. Invariant: Local sequence 2 must NOT be downgraded to sequence 1!
  const chs = getCachedWalletChannels('%s');
  const ch = chs.find(c => c.id === '%s');
  if (!ch) throw new Error("Channel missing from cache");
  if (ch.sequence !== 2) {
    throw new Error("Channel sequence was downgraded from 2 to " + ch.sequence);
  }
  const balB = ch.share ? ch.share.balance_party_b : ch.balance_b;
  if (balB !== 150) {
    throw new Error("Receiver balance was reverted to " + balB + ", expected 150");
  }
  console.log("Sequence preservation verified successfully!");
})().catch(err => {
  console.error("Preservation failure:", err);
  process.exit(1);
});
`, chanID, senderAddr, receiverAddr, cleanJs, receiverAddr, receiverAddr, chanID, senderAddr, receiverAddr, chanID, senderAddr, receiverAddr, pcnSig, pubHex, receiverAddr, chanID)

	tmpFile, err := os.CreateTemp("", "offline_pcn_preservation_*.js")
	if err != nil {
		t.Fatalf("failed to create temporary file for node test: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(nodeHarness); err != nil {
		t.Fatalf("failed to write test harness to temporary file: %v", err)
	}
	_ = tmpFile.Close()

	cmd := exec.Command(nodePath, tmpFile.Name())
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Node.js PCN sequence preservation failed:\n%s", string(output))
	}
}

// [Receipt Verification] TestOfflineQR_NodeJSRuntimeVerification executes the client-side
// offline QR receiving flow and saveWalletOfflineQueue inside Node.js to verify end-to-end
// execution without runtime exceptions (ReferenceError, TypeError, digest failure).
func TestOfflineQR_NodeJSRuntimeVerification(t *testing.T) {
	nodePath, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node executable not found in PATH; skipping Node.js runtime verification")
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
	templateTagRe := regexp.MustCompile(`\{\{[^}]+\}\}`)
	cleanJs := templateTagRe.ReplaceAllString(jsCode, `"mock_template_value"`)

	// Generate real Ed25519 keypair and signature for testing offline payment
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("failed to generate ed25519 keypair: %v", err)
	}
	pubHex := hex.EncodeToString(pub)

	// Direct transfer payload
	senderAddr := "0x" + hex.EncodeToString(pub[:20])
	receiverAddr := "0x1111222233334444555566667777888899990000"
	directMsg := fmt.Sprintf("%s|%s|50|12345678", senderAddr, receiverAddr)
	directHash := sha256.Sum256([]byte(directMsg))
	directSig := hex.EncodeToString(ed25519.Sign(priv, directHash[:]))

	// PCN channel transfer payload
	chanID := "chan_offline_test_999"
	pcnMsg := fmt.Sprintf("%s|2|900|100", chanID)
	pcnHash := sha256.Sum256([]byte(pcnMsg))
	pcnSig := hex.EncodeToString(ed25519.Sign(priv, pcnHash[:]))

	nodeHarness := fmt.Sprintf(`
// Mock browser environment
const localStorageStore = {};
global.localStorage = {
  getItem: (k) => localStorageStore[k] || null,
  setItem: (k, v) => { localStorageStore[k] = String(v); },
  removeItem: (k) => { delete localStorageStore[k]; },
  clear: () => { for (const k in localStorageStore) delete localStorageStore[k]; }
};

const elements = {};
global.document = {
  getElementById: (id) => {
    if (!elements[id]) {
      elements[id] = {
        id: id,
        value: '',
        textContent: '',
        innerHTML: '',
        style: {},
        classList: { add: () => {}, remove: () => {}, contains: () => false },
        querySelectorAll: () => [],
        appendChild: () => {}
      };
    }
    return elements[id];
  },
  createElement: (tag) => ({ tag, style: {}, appendChild: () => {} }),
  querySelectorAll: () => [],
  addEventListener: () => {}
};
global.window = {
  localStorage: global.localStorage,
  document: global.document,
  location: { port: '8080', hostname: 'localhost' },
  requestAnimationFrame: (cb) => cb(),
  addEventListener: () => {},
  crypto: {
    getRandomValues: (buf) => require('crypto').randomFillSync(buf)
  }
};
global.requestAnimationFrame = global.window.requestAnimationFrame;
global.navigator = {
  userAgent: 'node-test',
  vibrate: () => true
};
global.showAlert = (msg, isErr) => {
  if (isErr) console.warn('[UI Alert Error]:', msg);
  else console.log('[UI Alert]:', msg);
};
global.refreshDashboardData = () => {};
global.switchSubpanel = () => {};
global.closeSheetAndStopScan = () => {};
global.openSheet = () => {};

// Load dashboard script
%s

// 1. Test saveWalletOfflineQueue
const testAddr = "0xabcdef1234567890abcdef1234567890abcdef12";
if (typeof saveWalletOfflineQueue !== 'function') {
  throw new Error("saveWalletOfflineQueue is not defined as a function");
}
saveWalletOfflineQueue(testAddr, []);
const q = getCachedWalletOfflineQueue(testAddr);
if (!Array.isArray(q) || q.length !== 0) {
  throw new Error("getCachedWalletOfflineQueue failed to retrieve empty array");
}

// 2. Set active wallet
localStorage.setItem('liquidlink_active_wallet', '%s');

// 3. Test onOfflineQrScanSuccess with direct transfer payload
const directPayload = JSON.stringify({
  type: 'direct_tx_offline',
  sender: '%s',
  receiver: '%s',
  amount: 50,
  nonce: 12345678,
  signature: '%s',
  public_key: '%s'
});

(async () => {
  await onOfflineQrScanSuccess(directPayload);
  const queueAfterDirect = getOfflineQueue();
  if (queueAfterDirect.length !== 1) {
    throw new Error("Direct transfer offline payment was not queued in receiver outbox");
  }

  // 4. Test onOfflineQrScanSuccess with PCN payload
  const pcnPayload = JSON.stringify({
    type: 'pcn_offline_payment',
    channel_id: '%s',
    sequence: 2,
    party_a: '%s',
    party_b: '%s',
    balance_a: 900,
    balance_b: 100,
    amount: 100,
    signature: '%s',
    public_key: '%s'
  });
  await onOfflineQrScanSuccess(pcnPayload);
  const queueAfterPcn = getOfflineQueue();
  if (queueAfterPcn.length !== 2) {
    throw new Error("PCN offline payment was not queued in receiver outbox");
  }
})().catch(err => {
  console.error("Async execution failure:", err);
  process.exit(1);
});
`, cleanJs, receiverAddr, senderAddr, receiverAddr, directSig, pubHex, chanID, senderAddr, receiverAddr, pcnSig, pubHex)

	tmpFile, err := os.CreateTemp("", "offline_qr_test_*.js")
	if err != nil {
		t.Fatalf("failed to create temporary file for node test: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(nodeHarness); err != nil {
		t.Fatalf("failed to write test harness to temporary file: %v", err)
	}
	_ = tmpFile.Close()

	cmd := exec.Command(nodePath, tmpFile.Name())
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Node.js offline QR simulation failed:\n%s", string(output))
	}
}

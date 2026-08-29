package api

import (
	"html/template"
	"net/http"
)

// [Dashboard] renderDashboard compiles and executes the web dashboard HTML template.
func renderDashboard(w http.ResponseWriter, d *dashboardData) error {
	t, err := template.New("dash").Parse(dashboardHTML)
	if err != nil {
		return err
	}
	return t.Execute(w, d)
}

const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>LiquidLink Node Dashboard</title>
<style>
:root {
  --primary: #0969da;
  --primary-hover: #044289;
  --bg: #f6f8fa;
  --card-bg: #ffffff;
  --border: #d0d7de;
  --text: #24292e;
  --text-muted: #57606a;
  --success: #1a7f37;
  --success-bg: #dafbe1;
  --danger: #cf222e;
  --danger-bg: #ffebe9;
  --font-mono: ui-monospace, SFMono-Regular, 'SF Mono', Consolas, monospace;
}
* { box-sizing: border-box; }
body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; margin: 0; background: var(--bg); color: var(--text); line-height: 1.5; }
.header { background: #1f2328; color: #fff; padding: 20px 0; border-bottom: 1px solid var(--border); }
.header-container { max-width: 1200px; margin: 0 auto; padding: 0 20px; display: flex; justify-content: space-between; align-items: center; }
.brand { font-size: 1.5rem; font-weight: 700; color: #58a6ff; display: flex; align-items: center; gap: 8px; }
.node-badge { background: #21262d; border: 1px solid #30363d; font-size: 0.85rem; padding: 4px 12px; border-radius: 20px; color: #8b949e; }
.container { max-width: 1200px; margin: 30px auto; padding: 0 20px; }
.grid-metrics { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 16px; margin-bottom: 30px; }
.metric-card { background: var(--card-bg); border: 1px solid var(--border); border-radius: 8px; padding: 18px; box-shadow: 0 1px 3px rgba(0,0,0,0.04); }
.metric-title { font-size: 0.75rem; font-weight: 700; text-transform: uppercase; color: var(--text-muted); letter-spacing: 0.5px; margin-bottom: 6px; }
.metric-val { font-size: 1.15rem; font-weight: 700; word-break: break-all; }
.mono { font-family: var(--font-mono); }
.status-pill { display: inline-block; padding: 2px 10px; border-radius: 12px; font-size: 0.8rem; font-weight: 600; background: var(--success-bg); color: var(--success); }
.card { background: var(--card-bg); border: 1px solid var(--border); border-radius: 8px; margin-bottom: 28px; box-shadow: 0 1px 3px rgba(0,0,0,0.04); overflow: hidden; }
.card-header { padding: 16px 20px; border-bottom: 1px solid var(--border); background: #fcfcfc; display: flex; justify-content: space-between; align-items: center; }
.card-header h2 { font-size: 1.1rem; margin: 0; color: var(--text); font-weight: 600; }
.card-body { padding: 20px; }
.tabs { display: flex; border-bottom: 1px solid var(--border); background: #f6f8fa; padding: 0 16px; }
.tab-btn { padding: 12px 18px; border: none; background: none; font-size: 0.9rem; font-weight: 600; color: var(--text-muted); cursor: pointer; border-bottom: 2px solid transparent; }
.tab-btn.active { color: var(--primary); border-bottom-color: var(--primary); background: #fff; }
.tab-content { display: none; padding: 20px; }
.tab-content.active { display: block; }
.form-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 16px; }
.form-group { margin-bottom: 14px; }
label { display: block; font-size: 0.85rem; font-weight: 600; margin-bottom: 6px; color: var(--text); }
input[type="text"], input[type="number"], select { width: 100%; padding: 8px 12px; border: 1px solid var(--border); border-radius: 6px; font-size: 0.9rem; font-family: inherit; }
input:focus, select:focus { border-color: var(--primary); outline: none; box-shadow: 0 0 0 3px rgba(9,105,218,0.15); }
.btn { background: var(--primary); color: white; border: none; padding: 9px 18px; font-size: 0.9rem; font-weight: 600; border-radius: 6px; cursor: pointer; transition: background 0.15s ease; }
.btn:hover { background: var(--primary-hover); }
.btn-sm { padding: 4px 10px; font-size: 0.8rem; }
.btn-secondary { background: #f3f4f6; color: var(--text); border: 1px solid var(--border); }
.btn-secondary:hover { background: #e5e7eb; }
.table-wrap { overflow-x: auto; }
table { width: 100%; border-collapse: collapse; text-align: left; font-size: 0.88rem; }
th { background: #f6f8fa; padding: 10px 14px; font-weight: 600; color: var(--text-muted); border-bottom: 1px solid var(--border); }
td { padding: 12px 14px; border-bottom: 1px solid var(--border); vertical-align: middle; }
tr:hover { background: #fbfcfd; }
.alert { padding: 12px 16px; border-radius: 6px; font-size: 0.9rem; margin-bottom: 16px; display: none; }
.alert-success { background: var(--success-bg); color: var(--success); border: 1px solid #b4f4c6; }
.alert-danger { background: var(--danger-bg); color: var(--danger); border: 1px solid #ffc9c9; }
.badge { display: inline-block; padding: 2px 8px; border-radius: 10px; font-size: 0.75rem; font-weight: 600; text-transform: uppercase; }
.badge-active { background: var(--success-bg); color: var(--success); }
.badge-pending { background: #fff8c5; color: #9a6700; }
.badge-settled { background: #ddf4ff; color: #0969da; }
</style>
</head>
<body>

<header class="header">
  <div class="header-container">
    <div class="brand">⚡ LiquidLink Node Console</div>
    <div class="node-badge">Node ID: <strong>{{ .NodeID }}</strong> · Port: {{ .Port }}</div>
  </div>
</header>

<div class="container">

  <div id="alertBox" class="alert"></div>

  <!-- Metrics Grid -->
  <div class="grid-metrics">
    <div class="metric-card">
      <div class="metric-title">Wallet Address</div>
      <div class="metric-val mono" style="font-size: 0.85rem;">{{ if .Address }}{{ .Address }}{{ else }}—{{ end }}</div>
    </div>
    <div class="metric-card">
      <div class="metric-title">Wallet Balance</div>
      <div class="metric-val mono">{{ .Balance }}</div>
    </div>
    <div class="metric-card">
      <div class="metric-title">Total Token Supply</div>
      <div class="metric-val mono">{{ .TotalSupply }}</div>
    </div>
    <div class="metric-card">
      <div class="metric-title">Block Height</div>
      <div class="metric-val">{{ .BlockHeight }}</div>
    </div>
    <div class="metric-card">
      <div class="metric-title">Connected Peers</div>
      <div class="metric-val">{{ .PeerCount }}</div>
    </div>
    <div class="metric-card">
      <div class="metric-title">Node Status</div>
      <div class="metric-val"><span class="status-pill">● Online</span></div>
    </div>
  </div>

  <!-- Quick Actions Control Panel -->
  <div class="card">
    <div class="card-header">
      <h2>Interactive Node Operations</h2>
    </div>
    <div class="tabs">
      <button class="tab-btn active" onclick="switchTab('tab-transfer')">Transfer Tokens</button>
      <button class="tab-btn" onclick="switchTab('tab-mint')">Mint Tokens</button>
      <button class="tab-btn" onclick="switchTab('tab-channel')">State Channels</button>
      <button class="tab-btn" onclick="switchTab('tab-bridge')">Fiat Bridge</button>
    </div>

    <!-- Transfer Tab -->
    <div id="tab-transfer" class="tab-content active">
      <form onsubmit="submitTransfer(event)">
        <div class="form-grid">
          <div class="form-group">
            <label for="transferReceiver">Receiver Address</label>
            <input type="text" id="transferReceiver" placeholder="e.g. 0x1234... or node-2-addr" required>
          </div>
          <div class="form-group">
            <label for="transferAmount">Amount</label>
            <input type="number" id="transferAmount" min="1" placeholder="Amount to transfer" required>
          </div>
        </div>
        <button type="submit" class="btn">Send Transfer</button>
      </form>
    </div>

    <!-- Mint Tab -->
    <div id="tab-mint" class="tab-content">
      <form onsubmit="submitMint(event)">
        <div class="form-grid">
          <div class="form-group">
            <label for="mintAddress">Target Address (leave blank for local wallet)</label>
            <input type="text" id="mintAddress" placeholder="{{ .Address }}">
          </div>
          <div class="form-group">
            <label for="mintAmount">Mint Amount</label>
            <input type="number" id="mintAmount" min="1" placeholder="Amount to mint" required>
          </div>
        </div>
        <button type="submit" class="btn">Mint Tokens</button>
      </form>
    </div>

    <!-- State Channels Tab -->
    <div id="tab-channel" class="tab-content">
      <div style="display: flex; gap: 24px; flex-wrap: wrap;">
        <!-- Open Channel -->
        <div style="flex: 1; min-width: 280px; border-right: 1px solid var(--border); padding-right: 20px;">
          <h4 style="margin-top:0;">Open State Channel</h4>
          <form onsubmit="submitOpenChannel(event)">
            <div class="form-group">
              <label for="chanCounterparty">Counterparty Address</label>
              <input type="text" id="chanCounterparty" placeholder="Counterparty wallet address" required>
            </div>
            <div class="form-group">
              <label for="chanAmount">Deposit Amount</label>
              <input type="number" id="chanAmount" min="1" placeholder="Initial deposit" required>
            </div>
            <button type="submit" class="btn">Open Channel</button>
          </form>
        </div>

        <!-- Update / Settle Channel -->
        <div style="flex: 1; min-width: 280px;">
          <h4 style="margin-top:0;">Update State / Settle Channel</h4>
          <form onsubmit="submitUpdateChannel(event)">
            <div class="form-group">
              <label for="chanId">Channel ID</label>
              <input type="text" id="chanId" placeholder="chan-..." required>
            </div>
            <div class="form-grid">
              <div class="form-group">
                <label for="chanUpdateAmt">Amount</label>
                <input type="number" id="chanUpdateAmt" min="1" placeholder="Amount">
              </div>
              <div class="form-group">
                <label for="chanDirection">Direction</label>
                <select id="chanDirection">
                  <option value="true">Sending</option>
                  <option value="false">Receiving</option>
                </select>
              </div>
            </div>
            <div style="display:flex; gap:10px;">
              <button type="submit" class="btn">Update Balance</button>
              <button type="button" class="btn btn-secondary" onclick="submitSettleChannel('initiate')">Initiate Settlement</button>
              <button type="button" class="btn btn-secondary" onclick="submitSettleChannel('finalize')">Finalize</button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- Fiat Bridge Tab -->
    <div id="tab-bridge" class="tab-content">
      <div style="display: flex; gap: 24px; flex-wrap: wrap;">
        <!-- Request Bridge -->
        <div style="flex: 1; min-width: 280px; border-right: 1px solid var(--border); padding-right: 20px;">
          <h4 style="margin-top:0;">Initiate Fiat Bridge Request</h4>
          <form onsubmit="submitBridgeRequest(event)">
            <div class="form-group">
              <label for="bridgeReceiver">Receiver Account / Address</label>
              <input type="text" id="bridgeReceiver" placeholder="Fiat recipient account / crypto address" required>
            </div>
            <div class="form-group">
              <label for="bridgeAmount">Bridge Amount</label>
              <input type="number" id="bridgeAmount" min="1" placeholder="Amount" required>
            </div>
            <button type="submit" class="btn">Request Fiat Bridge</button>
          </form>
        </div>

        <!-- Verify Receipt -->
        <div style="flex: 1; min-width: 280px;">
          <h4 style="margin-top:0;">Verify UTR &amp; Settle Receipt</h4>
          <form onsubmit="submitVerifyReceipt(event)">
            <div class="form-group">
              <label for="verifyUTR">Unique Transaction Ref (UTR)</label>
              <input type="text" id="verifyUTR" placeholder="UTR-..." required>
            </div>
            <div class="form-group">
              <label for="verifyReceiptID">Receipt ID (Optional)</label>
              <input type="text" id="verifyReceiptID" placeholder="rec-...">
            </div>
            <button type="submit" class="btn">Settle Bridge Request</button>
          </form>
        </div>
      </div>
    </div>
  </div>

  <!-- State Channels Table -->
  <div class="card">
    <div class="card-header">
      <h2>Payment State Channels ({{ len .Channels }})</h2>
    </div>
    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Channel ID</th>
            <th>State</th>
            <th>Party A</th>
            <th>Party B</th>
            <th>Balance A</th>
            <th>Balance B</th>
            <th>Seq</th>
          </tr>
        </thead>
        <tbody>
          {{ range .Channels }}
          <tr>
            <td class="mono">{{ .ID }}</td>
            <td><span class="badge {{ if eq (printf "%v" .State) "OPEN" }}badge-active{{ else }}badge-pending{{ end }}">{{ .State }}</span></td>
            <td class="mono">{{ .Share.PartyA }}</td>
            <td class="mono">{{ .Share.PartyB }}</td>
            <td class="mono">{{ .Share.BalancePartyA }}</td>
            <td class="mono">{{ .Share.BalancePartyB }}</td>
            <td>{{ .Sequence }}</td>
          </tr>
          {{ else }}
          <tr><td colspan="7" style="text-align:center; color:var(--text-muted);">No active state channels registered.</td></tr>
          {{ end }}
        </tbody>
      </table>
    </div>
  </div>

  <!-- Fiat Bridge UTR Mappings Table -->
  <div class="card">
    <div class="card-header">
      <h2>Fiat Bridge UTR Mappings ({{ len .UTRMappings }})</h2>
    </div>
    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>UTR Reference</th>
            <th>Transaction ID</th>
            <th>Status</th>
            <th>Created At</th>
          </tr>
        </thead>
        <tbody>
          {{ range .UTRMappings }}
          <tr>
            <td class="mono">{{ .UTR }}</td>
            <td class="mono">{{ .TransactionID }}</td>
            <td><span class="badge {{ if eq (printf "%v" .Status) "SETTLED" }}badge-settled{{ else }}badge-pending{{ end }}">{{ .Status }}</span></td>
            <td>{{ .CreatedAt.Format "2006-01-02 15:04:05" }}</td>
          </tr>
          {{ else }}
          <tr><td colspan="4" style="text-align:center; color:var(--text-muted);">No fiat bridge requests recorded.</td></tr>
          {{ end }}
        </tbody>
      </table>
    </div>
  </div>

  <!-- Recent Blocks Table -->
  <div class="card">
    <div class="card-header">
      <h2>Blockchain Sequence — Recent Blocks ({{ len .RecentBlocks }})</h2>
    </div>
    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Height</th>
            <th>Block Hash</th>
            <th>Nonce</th>
            <th>Timestamp</th>
            <th>Tx Count</th>
          </tr>
        </thead>
        <tbody>
          {{ range .RecentBlocks }}
          <tr>
            <td><strong>{{ .Header.Index }}</strong></td>
            <td class="mono">{{ .Header.Hash }}</td>
            <td>{{ .Header.Nonce }}</td>
            <td>{{ .Header.Timestamp.Format "2006-01-02 15:04:05" }}</td>
            <td>{{ len .Transactions }}</td>
          </tr>
          {{ else }}
          <tr><td colspan="5" style="text-align:center; color:var(--text-muted);">No blocks produced yet.</td></tr>
          {{ end }}
        </tbody>
      </table>
    </div>
  </div>

</div>

<script>
function switchTab(tabId) {
  document.querySelectorAll('.tab-btn').forEach(btn => btn.classList.remove('active'));
  document.querySelectorAll('.tab-content').forEach(content => content.classList.remove('active'));
  event.target.classList.add('active');
  document.getElementById(tabId).classList.add('active');
}

function showAlert(msg, isError = false) {
  const box = document.getElementById('alertBox');
  box.className = 'alert ' + (isError ? 'alert-danger' : 'alert-success');
  box.textContent = msg;
  box.style.display = 'block';
  window.scrollTo({ top: 0, behavior: 'smooth' });
  setTimeout(() => { box.style.display = 'none'; }, 6000);
}

async function postData(url, data) {
  const resp = await fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data)
  });
  const resData = await resp.json();
  if (!resp.ok) {
    throw new Error(resData.error || 'Request failed with status ' + resp.status);
  }
  return resData;
}

async function submitTransfer(e) {
  e.preventDefault();
  try {
    const receiver = document.getElementById('transferReceiver').value;
    const amount = parseInt(document.getElementById('transferAmount').value, 10);
    const res = await postData('/api/v1/wallet/transfer', { receiver, amount });
    showAlert('Transfer successful! Amount ' + res.amount + ' sent to ' + res.receiver);
    setTimeout(() => location.reload(), 1500);
  } catch (err) {
    showAlert(err.message, true);
  }
}

async function submitMint(e) {
  e.preventDefault();
  try {
    const address = document.getElementById('mintAddress').value;
    const amount = parseInt(document.getElementById('mintAmount').value, 10);
    const res = await postData('/api/v1/wallet/mint', { address, amount });
    showAlert('Tokens minted successfully! New Balance: ' + res.new_balance);
    setTimeout(() => location.reload(), 1500);
  } catch (err) {
    showAlert(err.message, true);
  }
}

async function submitOpenChannel(e) {
  e.preventDefault();
  try {
    const counterparty = document.getElementById('chanCounterparty').value;
    const amount = parseInt(document.getElementById('chanAmount').value, 10);
    const res = await postData('/api/v1/channels/open', { counterparty, amount });
    showAlert('State channel opened successfully! Channel ID: ' + res.id);
    setTimeout(() => location.reload(), 1500);
  } catch (err) {
    showAlert(err.message, true);
  }
}

async function submitUpdateChannel(e) {
  e.preventDefault();
  try {
    const channel_id = document.getElementById('chanId').value;
    const amount = parseInt(document.getElementById('chanUpdateAmt').value, 10);
    const sending = document.getElementById('chanDirection').value === 'true';
    const res = await postData('/api/v1/channels/update', { channel_id, amount, sending });
    showAlert('Channel state updated successfully! Sequence: ' + res.sequence);
    setTimeout(() => location.reload(), 1500);
  } catch (err) {
    showAlert(err.message, true);
  }
}

async function submitSettleChannel(action) {
  try {
    const channel_id = document.getElementById('chanId').value;
    if (!channel_id) {
      showAlert('Channel ID is required for settlement', true);
      return;
    }
    const res = await postData('/api/v1/channels/settle', { channel_id, action });
    showAlert('Channel settlement ' + action + ' executed successfully!');
    setTimeout(() => location.reload(), 1500);
  } catch (err) {
    showAlert(err.message, true);
  }
}

async function submitBridgeRequest(e) {
  e.preventDefault();
  try {
    const receiver = document.getElementById('bridgeReceiver').value;
    const amount = parseInt(document.getElementById('bridgeAmount').value, 10);
    const res = await postData('/api/v1/bridge/fiat-request', { receiver, amount });
    showAlert('Fiat bridge request created! UTR: ' + res.mapping.utr);
    setTimeout(() => location.reload(), 1500);
  } catch (err) {
    showAlert(err.message, true);
  }
}

async function submitVerifyReceipt(e) {
  e.preventDefault();
  try {
    const utr = document.getElementById('verifyUTR').value;
    const receipt_id = document.getElementById('verifyReceiptID').value;
    const res = await postData('/api/v1/bridge/verify-receipt', { utr, receipt_id });
    showAlert('Receipt verified and bridge request settled! UTR: ' + res.utr);
    setTimeout(() => location.reload(), 1500);
  } catch (err) {
    showAlert(err.message, true);
  }
}
</script>
</body>
</html>`

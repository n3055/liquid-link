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
<meta name="viewport" content="width=device-width, initial-scale=1.0, viewport-fit=cover">
<title>LiquidLink Wallet</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=IBM+Plex+Mono:wght@400;500;600&display=swap" rel="stylesheet">
<script src="https://cdnjs.cloudflare.com/ajax/libs/qrcodejs/1.0.0/qrcode.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/html5-qrcode/2.3.8/html5-qrcode.min.js"></script>
<style>
:root {
  --bg: #0c0d10;
  --bg-elevated: #121317;
  --surface: #17181d;
  --surface-2: #1e2025;
  --border: #2b2d34;
  --border-soft: #1d1f24;
  --text: #eceef1;
  --text-dim: #989ba3;
  --text-faint: #5f6269;
  --accent: #4d84e2;
  --accent-dim: #396bb8;
  --accent-strong: #6d9bee;
  --accent-ink: #ffffff;
  --accent-soft: rgba(77,132,226,0.12);
  --success: #45a374;
  --success-soft: rgba(69,163,116,0.12);
  --amber: #b98f45;
  --amber-soft: rgba(185,143,69,0.14);
  --red: #c96a6a;
  --red-soft: rgba(201,106,106,0.14);
  --radius-lg: 18px;
  --radius: 12px;
  --radius-sm: 9px;
  --font-display: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  --font-body: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  --font-mono: 'IBM Plex Mono', ui-monospace, SFMono-Regular, 'SF Mono', Consolas, monospace;
  --nav-h: 68px;
}
* { box-sizing: border-box; }
html, body { height: 100%; }
body {
  font-family: var(--font-body);
  margin: 0;
  background: var(--bg);
  color: var(--text);
  line-height: 1.5;
  padding-bottom: calc(var(--nav-h) + 24px);
  -webkit-font-smoothing: antialiased;
  text-rendering: optimizeLegibility;
}
::selection { background: var(--accent-soft); color: var(--accent); }
button { font-family: inherit; }
svg { display: block; }

/* ---------- Top bar ---------- */
.topbar {
  position: sticky;
  top: 0;
  z-index: 30;
  background: rgba(8, 10, 16, 0.85);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.25);
}
.topbar-inner {
  max-width: 560px;
  margin: 0 auto;
  padding: 12px 18px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
}
.brand-mark {
  width: 32px;
  height: 32px;
  border-radius: 10px;
  background: linear-gradient(135deg, rgba(56,189,248,0.15) 0%, rgba(129,140,248,0.15) 100%);
  border: 1px solid rgba(56,189,248,0.3);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 0 12px rgba(56,189,248,0.2);
}
.brand-title {
  font-family: var(--font-display);
  font-weight: 800;
  font-size: 1.05rem;
  letter-spacing: -0.02em;
  background: linear-gradient(135deg, #ffffff 0%, #a5f3fc 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}
.node-pill {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  font-size: 0.74rem;
  background: rgba(15, 23, 42, 0.7);
  border: 1px solid rgba(56, 189, 248, 0.22);
  padding: 4px 11px 4px 9px;
  border-radius: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
  backdrop-filter: blur(8px);
}
.dot-pulse {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #10b981;
  flex-shrink: 0;
  box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.7);
  animation: pulseDot 2s infinite;
}
@keyframes pulseDot {
  0% { box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.7); }
  70% { box-shadow: 0 0 0 6px rgba(16, 185, 129, 0); }
  100% { box-shadow: 0 0 0 0 rgba(16, 185, 129, 0); }
}
.node-name {
  font-family: var(--font-mono);
  font-weight: 700;
  color: #f1f5f9;
  letter-spacing: -0.01em;
}
.node-port {
  font-family: var(--font-mono);
  font-weight: 700;
  color: #38bdf8;
  background: rgba(56, 189, 248, 0.12);
  padding: 2px 6px;
  border-radius: 5px;
  font-size: 0.7rem;
}
.avatar-btn {
  width: 36px !important;
  height: 36px !important;
  min-width: 36px !important;
  min-height: 36px !important;
  max-width: 36px !important;
  max-height: 36px !important;
  border-radius: 50% !important;
  aspect-ratio: 1 / 1 !important;
  background: var(--surface-2);
  border: 1.5px solid var(--border);
  display: inline-flex !important;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  flex-shrink: 0 !important;
  overflow: hidden !important;
  padding: 0 !important;
  transition: border-color 0.15s ease, transform 0.15s ease;
}
.avatar-btn:hover { border-color: var(--accent); transform: scale(1.05); }
.avatar-btn svg { width: 100% !important; height: 100% !important; border-radius: 50% !important; display: block !important; }

/* ---------- Layout ---------- */
.container { max-width: 560px; margin: 0 auto; padding: 18px 18px 6px; }

/* ---------- Hero balance card ---------- */
.hero { position: relative; border-radius: var(--radius-lg); margin-bottom: 18px; border: 1px solid var(--border); background: var(--surface); }
.hero-inner { position: relative; border-radius: var(--radius-lg); padding: 22px 20px 18px; }
.hero-label { font-size: 0.72rem; color: var(--text-faint); text-transform: uppercase; letter-spacing: 0.7px; font-weight: 600; }
.hero-balance { font-family: var(--font-display); font-size: 2.35rem; font-weight: 700; margin: 8px 0 16px; font-variant-numeric: tabular-nums; letter-spacing: -0.02em; word-break: break-all; }
.hero-balance small { font-size: 1rem; color: var(--text-faint); font-weight: 600; letter-spacing: 0; margin-left: 5px; }
.addr-chip { display: inline-flex; align-items: center; gap: 8px; background: var(--surface-2); border: 1px solid var(--border); border-radius: 8px; padding: 7px 8px 7px 12px; font-family: var(--font-mono); font-size: 0.78rem; color: var(--text-dim); max-width: 100%; }
.addr-chip span.a { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 220px; }
.icon-btn-mini { width: 26px; height: 26px; border-radius: 50%; border: none; background: var(--surface); color: var(--text-dim); display: flex; align-items: center; justify-content: center; cursor: pointer; flex-shrink: 0; }
.icon-btn-mini:hover { color: var(--accent); }

.quick-actions { display: grid; grid-template-columns: repeat(4, 1fr); gap: 8px; margin-top: 18px; }
.qa-btn { display: flex; flex-direction: column; align-items: center; gap: 8px; background: none; border: none; color: var(--text); cursor: pointer; padding: 4px 0; }
.qa-circle { width: 46px; height: 46px; border-radius: 50%; background: var(--surface-2); border: 1px solid var(--border); display: flex; align-items: center; justify-content: center; color: var(--accent); transition: border-color 0.12s ease; }
.qa-btn:hover .qa-circle { border-color: var(--accent); }
.qa-btn span { font-size: 0.71rem; color: var(--text-dim); font-weight: 600; letter-spacing: 0.1px; }

/* ---------- Stat strip ---------- */
.stat-strip { display: grid; grid-auto-flow: column; grid-auto-columns: minmax(120px, 1fr); gap: 10px; overflow-x: auto; padding-bottom: 4px; margin-bottom: 22px; scrollbar-width: none; }
.stat-strip::-webkit-scrollbar { display: none; }
.stat-card { background: var(--surface); border: 1px solid var(--border-soft); border-radius: var(--radius); padding: 12px 14px; }
.stat-title { font-size: 0.66rem; color: var(--text-faint); text-transform: uppercase; letter-spacing: 0.5px; font-weight: 600; margin-bottom: 6px; }
.stat-val { font-size: 0.95rem; font-weight: 600; font-variant-numeric: tabular-nums; letter-spacing: -0.01em; }
.status-pill { display: inline-flex; align-items: center; gap: 6px; font-size: 0.85rem; font-weight: 600; color: var(--success); }

/* ---------- Segmented control ---------- */
.segment { display: flex; gap: 4px; background: var(--surface); border: 1px solid var(--border-soft); border-radius: 12px; padding: 4px; margin-bottom: 14px; }
.segment button { flex: 1; border: none; background: none; padding: 9px 6px; border-radius: 9px; font-size: 0.82rem; font-weight: 600; color: var(--text-dim); cursor: pointer; }
.segment button.active { background: var(--surface-2); color: var(--text); border: 1px solid var(--border); }

.panel { display: none; }
.panel.active { display: block; }

/* ---------- Section card ---------- */
.section-card { background: var(--surface); border: 1px solid var(--border-soft); border-radius: var(--radius-lg); margin-bottom: 16px; overflow: hidden; }
.section-head { display: flex; align-items: center; justify-content: space-between; padding: 15px 16px; border-bottom: 1px solid var(--border-soft); }
.section-head h2 { font-family: var(--font-display); font-size: 0.95rem; margin: 0; font-weight: 600; letter-spacing: -0.01em; }
.section-head .count { color: var(--text-faint); font-weight: 500; font-size: 0.85rem; }
.head-actions { display: flex; gap: 8px; }

/* ---------- List rows (mobile-friendly, replaces dense tables) ---------- */
.list-row { display: flex; align-items: center; justify-content: space-between; gap: 12px; padding: 13px 16px; border-bottom: 1px solid var(--border-soft); }
.list-row:last-child { border-bottom: none; }
.list-row:hover { background: var(--surface-2); }
.row-main { min-width: 0; flex: 1; }
.row-title { display: flex; align-items: center; gap: 8px; font-family: var(--font-mono); font-size: 0.85rem; font-weight: 600; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.row-sub { font-size: 0.78rem; color: var(--text-dim); margin-top: 3px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.row-side { text-align: right; flex-shrink: 0; }
.row-side .top { font-size: 0.85rem; font-weight: 700; font-variant-numeric: tabular-nums; }
.row-side .bottom { font-size: 0.72rem; color: var(--text-faint); margin-top: 2px; }
.empty-row { padding: 30px 16px; text-align: center; color: var(--text-faint); font-size: 0.88rem; }

.badge { display: inline-block; padding: 3px 8px; border-radius: 6px; font-size: 0.65rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.3px; flex-shrink: 0; }
.badge-active { background: var(--success-soft); color: var(--success); }
.badge-pending { background: var(--amber-soft); color: var(--amber); }
.badge-settled { background: var(--accent-soft); color: var(--accent-strong); }

/* ---------- Buttons ---------- */
.btn { background: var(--accent); color: var(--accent-ink); border: none; padding: 12px 18px; font-size: 0.88rem; font-weight: 600; letter-spacing: 0.1px; border-radius: 10px; cursor: pointer; width: 100%; transition: background 0.12s ease; }
.btn:hover { background: var(--accent-strong); }
.btn-secondary { background: var(--surface-2); color: var(--text); border: 1px solid var(--border); }
.btn-ghost { background: none; color: var(--accent); border: 1px solid var(--border); }
.btn-sm { width: auto; padding: 7px 13px; font-size: 0.78rem; border-radius: 8px; }
.btn-row { display: flex; gap: 10px; }

/* ---------- Forms ---------- */
.form-group { margin-bottom: 14px; }
label { display: block; font-size: 0.8rem; font-weight: 600; margin-bottom: 6px; color: var(--text-dim); }
input[type="text"], input[type="number"], select {
  width: 100%; padding: 12px 14px; border: 1px solid var(--border); border-radius: var(--radius-sm);
  font-size: 0.92rem; font-family: inherit; background: var(--bg-elevated); color: var(--text);
}
input::placeholder { color: var(--text-faint); }
input:focus, select:focus { border-color: var(--accent); outline: none; box-shadow: 0 0 0 3px var(--accent-soft); }
.field-with-btn { display: flex; gap: 8px; }
.field-with-btn input { flex: 1; }
.field-with-btn button { flex-shrink: 0; }

/* ---------- Toast ---------- */
#alertBox { position: fixed; top: 14px; left: 50%; transform: translateX(-50%) translateY(-140%); z-index: 200; max-width: 440px; width: calc(100% - 32px); padding: 13px 18px; border-radius: var(--radius); font-size: 0.88rem; font-weight: 600; box-shadow: 0 12px 30px rgba(0,0,0,0.45); transition: transform 0.25s ease; }
#alertBox.show { transform: translateX(-50%) translateY(0); }
.alert-success { background: #0e2419; color: var(--accent); border: 1px solid var(--accent-dim); }
.alert-danger { background: #2a1216; color: var(--red); border: 1px solid #5a2129; }

/* ---------- Modal / Bottom sheet ---------- */
.overlay { position: fixed; inset: 0; background: rgba(4,5,8,0.65); backdrop-filter: blur(3px); z-index: 100; display: none; align-items: flex-end; justify-content: center; }
.overlay.open { display: flex; }
@media (min-width: 620px) { .overlay { align-items: center; } }
.sheet { background: var(--bg-elevated); border: 1px solid var(--border); width: 100%; max-width: 480px; border-radius: 22px 22px 0 0; padding: 20px 20px calc(20px + env(safe-area-inset-bottom)); max-height: 88vh; overflow-y: auto; animation: rise 0.22s ease; }
@media (min-width: 620px) { .sheet { border-radius: 22px; margin: 16px; } }
@keyframes rise { from { transform: translateY(24px); opacity: 0; } to { transform: translateY(0); opacity: 1; } }
.sheet-handle { width: 36px; height: 4px; background: var(--border); border-radius: 4px; margin: 0 auto 14px; }
.sheet-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; }
.sheet-head h3 { font-family: var(--font-display); margin: 0; font-size: 1.02rem; font-weight: 600; letter-spacing: -0.01em; }
.sheet-close { background: var(--surface-2); border: 1px solid var(--border); color: var(--text-dim); width: 30px; height: 30px; border-radius: 50%; cursor: pointer; display: flex; align-items: center; justify-content: center; transition: border-color 0.12s ease, color 0.12s ease; }
.sheet-close:hover { border-color: var(--accent); color: var(--accent); }
.sheet-tabs { display: flex; gap: 6px; margin-bottom: 16px; }
.sheet-tabs button { flex: 1; padding: 8px; border-radius: 8px; border: 1px solid var(--border); background: var(--surface); color: var(--text-dim); font-size: 0.78rem; font-weight: 600; cursor: pointer; }
.sheet-tabs button.active { background: var(--accent-soft); color: var(--accent); border-color: var(--accent-dim); }
.sheet-subpanel { display: none; }
.sheet-subpanel.active { display: block; }
.hint { font-size: 0.76rem; color: var(--text-faint); margin-top: -6px; margin-bottom: 14px; }

/* Receive / QR */
.qr-wrap { background: #fff; border-radius: 16px; padding: 18px; display: flex; align-items: center; justify-content: center; margin-bottom: 16px; }
.qr-address { text-align: center; font-family: var(--font-mono); font-size: 0.82rem; color: var(--text-dim); word-break: break-all; margin-bottom: 16px; }

/* Scanner */
.scan-frame { position: relative; border-radius: 16px; overflow: hidden; background: #000; aspect-ratio: 1/1; margin-bottom: 14px; min-height: 260px; }
#scannerRegion { width: 100% !important; height: 100% !important; min-height: 260px; display: flex; align-items: center; justify-content: center; }
#scannerRegion video { width: 100% !important; height: 100% !important; object-fit: cover !important; border-radius: 16px; display: block !important; }
#scannerRegion canvas { display: none !important; }
#scannerRegion img { width: 100% !important; height: 100% !important; object-fit: contain !important; }
.scan-frame::after { content: ""; position: absolute; inset: 14%; border: 1.5px solid rgba(255,255,255,0.6); border-radius: 16px; box-shadow: 0 0 0 999px rgba(0,0,0,0.4); pointer-events: none; }

/* Profile drawer */
.profile-avatar-lg { width: 72px; height: 72px; border-radius: 50%; display: flex; align-items: center; justify-content: center; margin: 0 auto 16px; box-shadow: 0 4px 16px rgba(0,0,0,0.4); border: 2px solid var(--border); overflow: hidden; }
.profile-item { display: flex; align-items: center; justify-content: space-between; padding: 13px 4px; border-bottom: 1px solid var(--border-soft); font-size: 0.88rem; }
.profile-item:last-child { border-bottom: none; }
.profile-item .k { color: var(--text-dim); }
.profile-item .v { font-family: var(--font-mono); font-weight: 600; }

/* ---------- Bottom nav (mobile) ---------- */
.bottom-nav { position: fixed; left: 0; right: 0; bottom: 0; height: var(--nav-h); background: rgba(15,17,23,0.92); backdrop-filter: blur(14px); border-top: 1px solid var(--border-soft); display: flex; align-items: center; justify-content: space-around; padding-bottom: env(safe-area-inset-bottom); z-index: 40; }
.nav-btn { background: none; border: none; color: var(--text-faint); display: flex; flex-direction: column; align-items: center; gap: 3px; font-size: 0.66rem; font-weight: 700; cursor: pointer; padding: 6px 10px; }
.nav-btn.active { color: var(--accent); }
.nav-btn.scan { color: #fff; }
.nav-scan-circle { width: 44px; height: 44px; border-radius: 50%; background: var(--accent); border: 1px solid var(--accent-dim); display: flex; align-items: center; justify-content: center; margin-top: -22px; }
@media (min-width: 720px) { .bottom-nav { display: none; } body { padding-bottom: 24px; } }
</style>
</head>
<body>

<div id="alertBox"></div>

<!-- Top bar -->
<header class="topbar">
  <div class="topbar-inner">
    <div class="brand">
      <span class="brand-mark">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
          <path d="M12 2L3 7v10l9 5 9-5V7l-9-5z" stroke="url(#brandGrad)" stroke-width="2" stroke-linejoin="round"/>
          <path d="M12 6l5 3v6l-5 3-5-3V9l5-3z" fill="url(#brandGrad)"/>
          <defs>
            <linearGradient id="brandGrad" x1="2" y1="2" x2="22" y2="22" gradientUnits="userSpaceOnUse">
              <stop stop-color="#38bdf8"/>
              <stop offset="1" stop-color="#818cf8"/>
            </linearGradient>
          </defs>
        </svg>
      </span>
      <span class="brand-title">LiquidLink Node Console</span>
    </div>
    <div class="node-pill">
      <span class="dot-pulse"></span>
      <span class="node-name">Node ID: {{ .NodeID }}</span>
      <span class="node-port">:{{ .Port }}</span>
    </div>
    <button class="avatar-btn" id="topbarAvatarBtn" onclick="openSheet('sheetProfile')" aria-label="Profile">
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none"><path d="M12 12a4.5 4.5 0 1 0 0-9 4.5 4.5 0 0 0 0 9Zm0 2c-4.2 0-7.5 2.2-7.5 5v1h15v-1c0-2.8-3.3-5-7.5-5Z" fill="currentColor"/></svg>
    </button>
  </div>
</header>

<div class="container">

  <!-- Hero balance -->
  <div class="hero">
    <div class="hero-inner">
      <div class="hero-label">Total balance</div>
      <div class="hero-balance">{{ .Balance }}<small>LQD</small></div>
      <div class="addr-chip">
        <span id="heroAddressAvatar" style="display:inline-flex;align-items:center;margin-right:2px;"></span>
        <span class="a" id="walletAddress">{{ if .Address }}{{ .Address }}{{ else }}No address assigned{{ end }}</span>
        <button class="icon-btn-mini" onclick="copyText('{{ .Address }}', this)" aria-label="Copy address">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none"><rect x="9" y="9" width="12" height="12" rx="2" stroke="currentColor" stroke-width="2"/><path d="M5 15V5a2 2 0 0 1 2-2h10" stroke="currentColor" stroke-width="2"/></svg>
        </button>
      </div>

      <div class="quick-actions">
        <button class="qa-btn" onclick="createNewWalletUI()">
          <span class="qa-circle"><svg width="19" height="19" viewBox="0 0 24 24" fill="none"><path d="M12 5v14M5 12h14" stroke="currentColor" stroke-width="2.2" stroke-linecap="round"/></svg></span>
          <span>New Wallet</span>
        </button>
        <button class="qa-btn" onclick="openSheet('sheetSend')">
          <span class="qa-circle"><svg width="19" height="19" viewBox="0 0 24 24" fill="none"><path d="M4 12h16M14 6l6 6-6 6" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"/></svg></span>
          <span>Send</span>
        </button>
        <button class="qa-btn" onclick="openReceive()">
          <span class="qa-circle"><svg width="19" height="19" viewBox="0 0 24 24" fill="none"><rect x="3" y="3" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="2"/><rect x="14" y="3" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="2"/><rect x="3" y="14" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="2"/><path d="M14 14h3v3h-3zM19 14h2v2h-2zM14 19h2v2h-2zM19 19h2v2h-2z" fill="currentColor"/></svg></span>
          <span>Receive</span>
        </button>
        <button class="qa-btn" onclick="openScanner('sendReceiver', 'sheetSend')">
          <span class="qa-circle"><svg width="19" height="19" viewBox="0 0 24 24" fill="none"><path d="M4 7V5a1 1 0 0 1 1-1h2M4 17v2a1 1 0 0 0 1 1h2M20 7V5a1 1 0 0 0-1-1h-2M20 17v2a1 1 0 0 1-1 1h-2" stroke="currentColor" stroke-width="2.2" stroke-linecap="round"/><path d="M4 12h16" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-dasharray="1 3.4"/></svg></span>
          <span>Scan</span>
        </button>
        <button class="qa-btn" onclick="openSheet('sheetMint')">
          <span class="qa-circle"><svg width="19" height="19" viewBox="0 0 24 24" fill="none"><path d="M12 5v14M5 12h14" stroke="currentColor" stroke-width="2.2" stroke-linecap="round"/></svg></span>
          <span>Mint</span>
        </button>
      </div>
    </div>
  </div>

  <!-- Secondary metrics -->
  <div class="stat-strip">
    <div class="stat-card"><div class="stat-title">Status</div><div class="status-pill"><span class="dot"></span>Online</div></div>
    <div class="stat-card"><div class="stat-title">Total supply</div><div class="stat-val">{{ .TotalSupply }}</div></div>
    <div class="stat-card"><div class="stat-title">Block height</div><div class="stat-val">{{ .BlockHeight }}</div></div>
    <div class="stat-card"><div class="stat-title">Peers</div><div class="stat-val">{{ .PeerCount }}</div></div>
  </div>

  <!-- Segmented control & Operations header -->
  <div style="margin-bottom:12px;">
    <h2 style="font-size:1.1rem; font-weight:600; color:var(--text); margin-bottom:12px;">Interactive Node Operations</h2>
  </div>
  <div class="segment">
    <button class="active" onclick="switchSegment('segOrderbook', this)">Orderbook (INR)</button>
    <button onclick="switchSegment('segStability', this)">Seigniorage & Holdings</button>
    <button onclick="switchSegment('segChannels', this)">Channels</button>
    <button onclick="switchSegment('segBridge', this)">Fiat bridge</button>
    <button onclick="switchSegment('segActivity', this)">Activity</button>
  </div>

  <!-- Orderbook Terminal -->
  <div id="segOrderbook" class="panel active">
    <div class="section-card">
      <div class="section-head">
        <h2>LQD / INR Orderbook <span class="count">Clearing Price: ₹{{ if .StabilityStatus }}{{ printf "%.2f" .StabilityStatus.Price }}{{ else }}1.00{{ end }} INR</span></h2>
      </div>

      <div style="display:grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap:16px; padding:16px;">
        <!-- Depth Bids / Asks -->
        <div style="background:var(--surface-2); border-radius:12px; padding:16px; border:1px solid var(--border);">
          <h3 style="font-size:0.95rem; margin-bottom:12px; color:var(--text);">Orderbook Depth (INR)</h3>
          <div style="display:grid; grid-template-columns: 1fr 1fr; gap:12px;">
            <div>
              <h4 style="font-size:0.8rem; color:var(--success); margin-bottom:6px;">Bids (Buy LQD)</h4>
              <div style="font-family:var(--font-mono); font-size:0.8rem;">
                {{ range .OrderbookDepth.Bids }}
                <div style="display:flex; justify-content:space-between; margin-bottom:4px;">
                  <span style="color:var(--success);">₹{{ printf "%.2f" .Price }}</span>
                  <span>{{ .Quantity }} LQD</span>
                </div>
                {{ else }}
                <div style="color:var(--text-faint);">No open bids</div>
                {{ end }}
              </div>
            </div>
            <div>
              <h4 style="font-size:0.8rem; color:var(--red); margin-bottom:6px;">Asks (Sell LQD)</h4>
              <div style="font-family:var(--font-mono); font-size:0.8rem;">
                {{ range .OrderbookDepth.Asks }}
                <div style="display:flex; justify-content:space-between; margin-bottom:4px;">
                  <span style="color:var(--red);">₹{{ printf "%.2f" .Price }}</span>
                  <span>{{ .Quantity }} LQD</span>
                </div>
                {{ else }}
                <div style="color:var(--text-faint);">No open asks</div>
                {{ end }}
              </div>
            </div>
          </div>
        </div>

        <!-- Place Order Form -->
        <div style="background:var(--surface-2); border-radius:12px; padding:16px; border:1px solid var(--border);">
          <h3 style="font-size:0.95rem; margin-bottom:12px; color:var(--text);">Place Limit Order</h3>
          <form onsubmit="submitOrderbookOrder(event)">
            <div class="field">
              <label>Order Type</label>
              <select id="obOrderSide" class="input" style="width:100%;">
                <option value="buy">Buy LQD (Bid in INR)</option>
                <option value="sell">Sell LQD (Ask in INR)</option>
              </select>
            </div>
            <div class="field" style="margin-top:10px;">
              <label>Limit Price in INR (₹)</label>
              <input type="number" id="obOrderPrice" step="0.01" value="1.00" class="input" placeholder="e.g. 1.00" required>
            </div>
            <div class="field" style="margin-top:10px;">
              <label>Token Amount (LQD)</label>
              <input type="number" id="obOrderAmount" min="1" value="100" class="input" placeholder="Quantity" required>
            </div>
            <button type="submit" class="btn btn-primary" style="width:100%; margin-top:14px;">Submit Order</button>
          </form>
        </div>
      </div>

      <!-- Executed Trades Log -->
      <div style="padding:16px; border-top:1px solid var(--border);">
        <h3 style="font-size:0.9rem; color:var(--text-dim); margin-bottom:10px;">Recent Trades</h3>
        {{ range .OrderbookTrades }}
        <div class="list-row">
          <div class="row-main">
            <div class="row-title">Trade #{{ .ID }}</div>
            <div class="row-sub">{{ .Buyer }} ← {{ .Seller }}</div>
          </div>
          <div class="row-side">
            <div class="top" style="color:var(--accent-strong);">₹{{ printf "%.2f" .Price }} INR</div>
            <div class="bottom">{{ .Amount }} LQD</div>
          </div>
        </div>
        {{ else }}
        <div class="empty-row">No trades executed yet on the orderbook.</div>
        {{ end }}
      </div>
    </div>
  </div>

  <!-- Seigniorage & Holdings -->
  <div id="segStability" class="panel">
    <div class="section-card">
      <div class="section-head">
        <h2>Seigniorage Peg & Holdings <span class="badge {{ if eq (printf "%v" .StabilityStatus.State) "expansion" }}badge-active{{ else }}badge-pending{{ end }}">{{ .StabilityStatus.State }}</span></h2>
      </div>

      <div style="padding:16px; display:grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap:12px; border-bottom:1px solid var(--border);">
        <div style="background:var(--surface-2); padding:12px; border-radius:10px;">
          <div style="font-size:0.75rem; color:var(--text-dim);">Target Peg</div>
          <div style="font-size:1.1rem; font-weight:700; color:var(--success);">₹1.00 INR</div>
        </div>
        <div style="background:var(--surface-2); padding:12px; border-radius:10px;">
          <div style="font-size:0.75rem; color:var(--text-dim);">Current Market Price</div>
          <div style="font-size:1.1rem; font-weight:700; color:var(--accent-strong);">₹{{ if .StabilityStatus }}{{ printf "%.2f" .StabilityStatus.Price }}{{ else }}1.00{{ end }} INR</div>
        </div>
        <div style="background:var(--surface-2); padding:12px; border-radius:10px;">
          <div style="font-size:0.75rem; color:var(--text-dim);">Staking Yield APY</div>
          <div style="font-size:1.1rem; font-weight:700; color:var(--amber);">{{ if .StabilityStatus }}{{ printf "%.1f" .StabilityStatus.CurrentAPY }}{{ else }}14.5{{ end }}%</div>
        </div>
        <div style="background:var(--surface-2); padding:12px; border-radius:10px;">
          <div style="font-size:0.75rem; color:var(--text-dim);">Staked Shares</div>
          <div style="font-size:1.1rem; font-weight:700;">{{ .StakedShares }} Shares</div>
        </div>
        <div style="background:var(--surface-2); padding:12px; border-radius:10px;">
          <div style="font-size:0.75rem; color:var(--text-dim);">Bonds Held</div>
          <div style="font-size:1.1rem; font-weight:700;">{{ .BondsHeld }} Bonds</div>
        </div>
      </div>

      <div style="padding:16px; display:grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap:16px;">
        <!-- Stake / Unstake -->
        <div style="background:var(--surface-2); border-radius:12px; padding:16px; border:1px solid var(--border);">
          <h3 style="font-size:0.95rem; margin-bottom:8px;">Share Holdings Staking</h3>
          <p style="font-size:0.8rem; color:var(--text-dim); margin-bottom:12px;">Stake LQD tokens to receive seigniorage expansion payouts during price growth.</p>
          <form onsubmit="submitStakeShares(event)">
            <div class="field">
              <label>Amount to Stake</label>
              <input type="number" id="stakeAmount" min="1" value="100" class="input" placeholder="LQD Amount" required>
            </div>
            <div style="display:flex; gap:10px; margin-top:12px;">
              <button type="submit" class="btn btn-primary" style="flex:1;">Stake Shares</button>
              <button type="button" onclick="submitUnstakeShares()" class="btn btn-secondary" style="flex:1;">Unstake & Claim</button>
            </div>
          </form>
        </div>

        <!-- Buy / Redeem Bonds -->
        <div style="background:var(--surface-2); border-radius:12px; padding:16px; border:1px solid var(--border);">
          <h3 style="font-size:0.95rem; margin-bottom:8px;">Bond Holdings Operations</h3>
          <p style="font-size:0.8rem; color:var(--text-dim); margin-bottom:12px;">Buy discounted LQD-Bonds during contraction (&lt; ₹1.00) or redeem for 1 LQD + 5% bonus during expansion (&gt; ₹1.00).</p>
          <form onsubmit="submitBuyBonds(event)">
            <div class="field">
              <label>Bond Amount</label>
              <input type="number" id="bondAmount" min="1" value="50" class="input" placeholder="LQD/Bond Amount" required>
            </div>
            <div style="display:flex; gap:10px; margin-top:12px;">
              <button type="submit" class="btn btn-primary" style="flex:1;">Buy Discounted Bonds</button>
              <button type="button" onclick="submitRedeemBonds()" class="btn btn-secondary" style="flex:1;">Redeem Bonds</button>
            </div>
          </form>
        </div>
      </div>

      <!-- Economist Design Explanation -->
      <div style="margin:16px; background:linear-gradient(135deg, rgba(77,132,226,0.1), rgba(129,140,248,0.1)); border:1px solid var(--accent-soft); border-radius:12px; padding:16px;">
        <h4 style="font-size:0.9rem; color:var(--accent-strong); margin-bottom:6px;">⚡ Economist Seigniorage Stability Engine</h4>
        <p style="font-size:0.82rem; color:var(--text-dim); line-height:1.4;">
          <strong>Expansion Phase (LQD &gt; ₹1.00 INR):</strong> When market demand drives price above peg, the protocol inflates supply by minting new LQD and distributing it directly to Share stakers. Bondholders can redeem LQD-Bonds for 1 LQD + 5% premium.<br>
          <strong>Contraction Phase (LQD &lt; ₹1.00 INR):</strong> When market price dips below peg, the protocol contracts circulating supply by selling LQD-Bonds at a 15% discount. Burning LQD removes excess liquidity and restores the ₹1.00 peg equilibrium.
        </p>
      </div>
    </div>
  </div>

  <!-- Channels -->
  <div id="segChannels" class="panel active">
    <div class="section-card">
      <div class="section-head">
        <h2>State channels <span class="count">({{ len .Channels }})</span></h2>
        <div class="head-actions">
          <button class="btn btn-sm btn-secondary" onclick="openSheet('sheetChannel')">Manage</button>
          <button class="btn btn-sm btn-ghost" onclick="openSheet('sheetChannel')">+ Open</button>
        </div>
      </div>
      {{ range .Channels }}
      <div class="list-row">
        <div class="row-main">
          <div class="row-title">{{ .ID }} <span class="badge {{ if eq (printf "%v" .State) "OPEN" }}badge-active{{ else }}badge-pending{{ end }}">{{ .State }}</span></div>
          <div class="row-sub">{{ .Share.PartyA }} → {{ .Share.PartyB }}</div>
        </div>
        <div class="row-side">
          <div class="top">{{ .Share.BalancePartyA }} / {{ .Share.BalancePartyB }}</div>
          <div class="bottom">seq {{ .Sequence }}</div>
        </div>
      </div>
      {{ else }}
      <div class="empty-row">No active state channels yet. Tap "+ Open" to create one.</div>
      {{ end }}
    </div>
  </div>

  <!-- Bridge -->
  <div id="segBridge" class="panel">
    <div class="section-card">
      <div class="section-head">
        <h2>Fiat bridge <span class="count">({{ len .UTRMappings }})</span></h2>
        <div class="head-actions">
          <button class="btn btn-sm btn-ghost" onclick="openSheet('sheetBridge')">+ Request</button>
        </div>
      </div>
      {{ range .UTRMappings }}
      <div class="list-row">
        <div class="row-main">
          <div class="row-title">{{ .UTR }} <span class="badge {{ if eq (printf "%v" .Status) "SETTLED" }}badge-settled{{ else }}badge-pending{{ end }}">{{ .Status }}</span></div>
          <div class="row-sub">Tx {{ .TransactionID }}</div>
        </div>
        <div class="row-side">
          <div class="bottom">{{ .CreatedAt.Format "02 Jan, 15:04" }}</div>
        </div>
      </div>
      {{ else }}
      <div class="empty-row">No fiat bridge requests yet.</div>
      {{ end }}
    </div>
  </div>

  <!-- Activity / Blocks -->
  <div id="segActivity" class="panel">
    <div class="section-card">
      <div class="section-head">
        <h2>Recent blocks <span class="count">({{ len .RecentBlocks }})</span></h2>
      </div>
      {{ range .RecentBlocks }}
      <div class="list-row">
        <div class="row-main">
          <div class="row-title">Block #{{ .Header.Index }}</div>
          <div class="row-sub" title="{{ .Header.Hash }}">{{ .Header.Hash }}</div>
        </div>
        <div class="row-side">
          <div class="top">{{ len .Transactions }} tx</div>
          <div class="bottom">{{ .Header.Timestamp.Format "02 Jan, 15:04" }}</div>
        </div>
      </div>
      {{ else }}
      <div class="empty-row">No blocks produced yet.</div>
      {{ end }}
    </div>
  </div>

</div>

<!-- Bottom nav -->
<nav class="bottom-nav">
  <button class="nav-btn active" onclick="scrollToTop()">
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none"><path d="M3 11.5 12 4l9 7.5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/><path d="M5 10v9a1 1 0 0 0 1 1h4v-6h4v6h4a1 1 0 0 0 1-1v-9" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
    Home
  </button>
  <button class="nav-btn" onclick="switchSegment('segActivity', document.querySelectorAll('.segment button')[2])">
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none"><path d="M3 12h4l2 7 5-14 2 7h5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
    Activity
  </button>
  <button class="nav-btn scan" onclick="openScanner('sendReceiver', 'sheetSend')">
    <span class="nav-scan-circle">
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none"><path d="M4 7V5a1 1 0 0 1 1-1h2M4 17v2a1 1 0 0 0 1 1h2M20 7V5a1 1 0 0 0-1-1h-2M20 17v2a1 1 0 0 1-1 1h-2" stroke="#ffffff" stroke-width="2.2" stroke-linecap="round"/><path d="M4 12h16" stroke="#ffffff" stroke-width="2.2" stroke-linecap="round" stroke-dasharray="1 3.4"/></svg>
    </span>
  </button>
  <button class="nav-btn" onclick="switchSegment('segChannels', document.querySelectorAll('.segment button')[0])">
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none"><path d="M7 7h10M7 12h10M7 17h6" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
    Channels
  </button>
  <button class="nav-btn" onclick="openSheet('sheetProfile')">
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none"><path d="M12 12a4.5 4.5 0 1 0 0-9 4.5 4.5 0 0 0 0 9Zm0 2c-4.2 0-7.5 2.2-7.5 5v1h15v-1c0-2.8-3.3-5-7.5-5Z" fill="currentColor"/></svg>
    Profile
  </button>
</nav>

<!-- ===================== SHEETS / MODALS ===================== -->

<!-- Send -->
<div class="overlay" id="sheetSend"><div class="sheet">
  <div class="sheet-handle"></div>
  <div class="sheet-head"><h3>Send tokens</h3><button class="sheet-close" onclick="closeSheet('sheetSend')"><svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M6 6l12 12M18 6L6 18" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg></button></div>
  <form onsubmit="submitTransfer(event)">
    <div class="form-group">
      <label for="sendReceiver">To</label>
      <div class="field-with-btn">
        <input type="text" id="sendReceiver" placeholder="Address, or tap scan" required>
        <button type="button" class="btn btn-sm btn-secondary" onclick="openScanner('sendReceiver', 'sheetSend')">Scan</button>
      </div>
    </div>
    <div class="form-group">
      <label for="sendAmount">Amount</label>
      <input type="number" id="sendAmount" min="1" placeholder="0" required>
    </div>
    <button type="submit" class="btn">Review & send</button>
  </form>
</div></div>

<!-- Receive / QR -->
<div class="overlay" id="sheetReceive"><div class="sheet">
  <div class="sheet-handle"></div>
  <div class="sheet-head"><h3>Receive</h3><button class="sheet-close" onclick="closeSheet('sheetReceive')"><svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M6 6l12 12M18 6L6 18" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg></button></div>
  <div class="qr-wrap" id="qrReceive"></div>
  <div class="qr-address">{{ if .Address }}{{ .Address }}{{ else }}No address assigned{{ end }}</div>
  <button class="btn" onclick="copyText('{{ .Address }}', this)">Copy address</button>
</div></div>

<!-- Scan -->
<div class="overlay" id="sheetScan"><div class="sheet">
  <div class="sheet-handle"></div>
  <div class="sheet-head"><h3>Scan a code</h3><button type="button" class="sheet-close" onclick="closeSheetAndStopScan(event)"><svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M6 6l12 12M18 6L6 18" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg></button></div>
  <div class="scan-frame"><div id="scannerRegion"></div></div>
  <div class="hint" id="scanHint">Point the camera at a LiquidLink QR code or choose a photo below.</div>
  <input type="file" id="qrFileInput" accept="image/*" style="display:none;" onchange="handleQrImageUpload(this)">
  <div class="btn-row" style="margin-top:10px;">
    <button type="button" class="btn btn-secondary" onclick="document.getElementById('qrFileInput').click()">📷 Upload QR Photo / Image</button>
  </div>
</div></div>

<!-- Mint -->
<div class="overlay" id="sheetMint"><div class="sheet">
  <div class="sheet-handle"></div>
  <div class="sheet-head"><h3>Mint tokens</h3><button class="sheet-close" onclick="closeSheet('sheetMint')"><svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M6 6l12 12M18 6L6 18" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg></button></div>
  <form onsubmit="submitMint(event)">
    <div class="form-group">
      <label for="mintAddress">Target address</label>
      <input type="text" id="mintAddress" placeholder="{{ .Address }} (leave blank for local wallet)">
    </div>
    <div class="form-group">
      <label for="mintAmount">Amount</label>
      <input type="number" id="mintAmount" min="1" placeholder="0" required>
    </div>
    <button type="submit" class="btn">Mint</button>
  </form>
</div></div>

<!-- Channel manage (open / update / settle) -->
<div class="overlay" id="sheetChannel"><div class="sheet">
  <div class="sheet-handle"></div>
  <div class="sheet-head"><h3>State channels</h3><button class="sheet-close" onclick="closeSheet('sheetChannel')"><svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M6 6l12 12M18 6L6 18" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg></button></div>
  <div class="sheet-tabs">
    <button class="active" onclick="switchSubpanel('subOpen', this, 'sheetChannel')">Open</button>
    <button onclick="switchSubpanel('subUpdate', this, 'sheetChannel')">Update</button>
    <button onclick="switchSubpanel('subSettle', this, 'sheetChannel')">Settle</button>
  </div>

  <div id="subOpen" class="sheet-subpanel active">
    <form onsubmit="submitOpenChannel(event)">
      <div class="form-group">
        <label for="chanCounterparty">Counterparty address</label>
        <input type="text" id="chanCounterparty" placeholder="Wallet address" required>
      </div>
      <div class="form-group">
        <label for="chanAmount">Deposit amount</label>
        <input type="number" id="chanAmount" min="1" placeholder="0" required>
      </div>
      <button type="submit" class="btn">Open channel</button>
    </form>
  </div>

  <div id="subUpdate" class="sheet-subpanel">
    <form onsubmit="submitUpdateChannel(event)">
      <div class="form-group">
        <label for="chanIdUpdate">Channel ID</label>
        <input type="text" id="chanIdUpdate" placeholder="chan-..." required>
      </div>
      <div class="form-group">
        <label for="chanUpdateAmt">Amount</label>
        <input type="number" id="chanUpdateAmt" min="1" placeholder="0">
      </div>
      <div class="form-group">
        <label for="chanDirection">Direction</label>
        <select id="chanDirection">
          <option value="true">Sending</option>
          <option value="false">Receiving</option>
        </select>
      </div>
      <button type="submit" class="btn">Update balance</button>
    </form>
  </div>

  <div id="subSettle" class="sheet-subpanel">
    <div class="form-group">
      <label for="chanIdSettle">Channel ID</label>
      <input type="text" id="chanIdSettle" placeholder="chan-..." required>
    </div>
    <div class="btn-row">
      <button class="btn btn-secondary" onclick="submitSettleChannel('initiate')">Initiate</button>
      <button class="btn" onclick="submitSettleChannel('finalize')">Finalize</button>
    </div>
  </div>
</div></div>

<!-- Bridge request / verify -->
<div class="overlay" id="sheetBridge"><div class="sheet">
  <div class="sheet-handle"></div>
  <div class="sheet-head"><h3>Fiat bridge</h3><button class="sheet-close" onclick="closeSheet('sheetBridge')"><svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M6 6l12 12M18 6L6 18" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg></button></div>
  <div class="sheet-tabs">
    <button class="active" onclick="switchSubpanel('subBridgeRequest', this, 'sheetBridge')">Request</button>
    <button onclick="switchSubpanel('subBridgeVerify', this, 'sheetBridge')">Verify receipt</button>
  </div>

  <div id="subBridgeRequest" class="sheet-subpanel active">
    <form onsubmit="submitBridgeRequest(event)">
      <div class="form-group">
        <label for="bridgeReceiver">Receiver account / address</label>
        <input type="text" id="bridgeReceiver" placeholder="Fiat account or crypto address" required>
      </div>
      <div class="form-group">
        <label for="bridgeAmount">Amount</label>
        <input type="number" id="bridgeAmount" min="1" placeholder="0" required>
      </div>
      <button type="submit" class="btn">Request bridge</button>
    </form>
  </div>

  <div id="subBridgeVerify" class="sheet-subpanel">
    <form onsubmit="submitVerifyReceipt(event)">
      <div class="form-group">
        <label for="verifyUTR">Unique transaction ref (UTR)</label>
        <input type="text" id="verifyUTR" placeholder="UTR-..." required>
      </div>
      <div class="form-group">
        <label for="verifyReceiptID">Receipt ID (optional)</label>
        <input type="text" id="verifyReceiptID" placeholder="rec-...">
      </div>
      <button type="submit" class="btn">Settle request</button>
    </form>
  </div>
</div></div>

<!-- Profile drawer -->
<div class="overlay" id="sheetProfile"><div class="sheet">
  <div class="sheet-handle"></div>
  <div class="sheet-head"><h3>Profile</h3><button class="sheet-close" onclick="closeSheet('sheetProfile')"><svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M6 6l12 12M18 6L6 18" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg></button></div>
  <div class="profile-avatar-lg" id="profileAvatarLg"></div>
  <div class="profile-item"><span class="k">Node ID</span><span class="v">{{ .NodeID }}</span></div>
  <div class="profile-item"><span class="k">Port</span><span class="v">{{ .Port }}</span></div>
  <div class="profile-item"><span class="k">Balance</span><span class="v">{{ .Balance }}</span></div>
  <div class="profile-item"><span class="k">Peers</span><span class="v">{{ .PeerCount }}</span></div>
  <div style="margin-top:16px;" class="btn-row">
    <button class="btn btn-secondary" onclick="closeSheet('sheetProfile'); openReceive();">Show my QR</button>
    <button class="btn btn-secondary" onclick="closeSheet('sheetProfile'); openScanner('sendReceiver','sheetSend');">Scan code</button>
  </div>
</div></div>

<script>
/* ---------- Toast ---------- */
let toastTimer;
function showAlert(msg, isError) {
  const box = document.getElementById('alertBox');
  box.className = (isError ? 'alert-danger' : 'alert-success');
  box.textContent = msg;
  requestAnimationFrame(() => box.classList.add('show'));
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => box.classList.remove('show'), 4200);
}

/* ---------- Sheets & Modal Management ---------- */
function openSheet(id) {
  const el = document.getElementById(id);
  if (el) {
    el.classList.add('open');
    document.body.style.overflow = 'hidden';
  }
}

function closeSheet(id) {
  const el = document.getElementById(id);
  if (el) {
    el.classList.remove('open');
  }
  if (document.querySelectorAll('.overlay.open').length === 0) {
    document.body.style.overflow = '';
  }
}

document.addEventListener('keydown', (e) => {
  if (e.key === 'Escape') {
    document.querySelectorAll('.overlay.open').forEach(ov => {
      if (ov.id === 'sheetScan') closeSheetAndStopScan();
      else closeSheet(ov.id);
    });
  }
});

function switchSegment(panelId, btn) {
  document.querySelectorAll('.segment button').forEach(b => b.classList.remove('active'));
  document.querySelectorAll('.panel').forEach(p => p.classList.remove('active'));
  btn.classList.add('active');
  document.getElementById(panelId).classList.add('active');
  document.querySelectorAll('.nav-btn').forEach(b => b.classList.remove('active'));
}

function switchSubpanel(id, btn, scope) {
  const root = document.getElementById(scope);
  root.querySelectorAll('.sheet-tabs button').forEach(b => b.classList.remove('active'));
  root.querySelectorAll('.sheet-subpanel').forEach(p => p.classList.remove('active'));
  btn.classList.add('active');
  document.getElementById(id).classList.add('active');
}

function scrollToTop() { window.scrollTo({ top: 0, behavior: 'smooth' }); }

/* ---------- Universal Clipboard Copy (Supports HTTP & Mobile) ---------- */
function copyText(text, btn) {
  if (!text || text === 'No address assigned') { showAlert('No address available to copy', true); return; }
  
  // Try modern Clipboard API if supported in current context
  if (navigator.clipboard && (window.isSecureContext || location.hostname === 'localhost' || location.hostname === '127.0.0.1')) {
    navigator.clipboard.writeText(text)
      .then(() => onCopySuccess(btn))
      .catch(() => fallbackCopyText(text, btn));
  } else {
    fallbackCopyText(text, btn);
  }
}

function fallbackCopyText(text, btn) {
  try {
    const textArea = document.createElement("textarea");
    textArea.value = text;
    textArea.style.position = "fixed";
    textArea.style.left = "-9999px";
    textArea.style.top = "-9999px";
    textArea.style.opacity = "0";
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();
    
    // For iOS Safari selection support
    textArea.setSelectionRange(0, 99999);
    
    const successful = document.execCommand('copy');
    document.body.removeChild(textArea);
    
    if (successful) {
      onCopySuccess(btn);
    } else {
      showAlert('Address: ' + text, false);
    }
  } catch (err) {
    showAlert('Address: ' + text, false);
  }
}

function onCopySuccess(btn) {
  // Show a clean green checkmark tick icon directly on the button instead of a toast message
  if (btn) {
    const origHTML = btn.innerHTML;
    const tickSVG = '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" style="display:inline-block;vertical-align:middle;color:var(--success);"><path d="M20 6L9 17l-5-5" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/></svg>';
    
    if (btn.classList.contains('icon-btn-mini')) {
      btn.innerHTML = tickSVG;
    } else {
      btn.innerHTML = '<span style="color:var(--success);display:inline-flex;align-items:center;gap:4px;">' + tickSVG + ' Copied</span>';
    }
    
    btn.style.borderColor = 'var(--success)';
    
    setTimeout(() => {
      btn.innerHTML = origHTML;
      btn.style.borderColor = '';
    }, 1800);
  }
}

/* ---------- Receive / QR Code ---------- */
function openReceive() {
  openSheet('sheetReceive');
  const target = document.getElementById('qrReceive');
  target.innerHTML = '';
  const address = {{ if .Address }}"{{ .Address }}"{{ else }}""{{ end }};
  if (address && window.QRCode) {
    new QRCode(target, { text: address, width: 200, height: 200, colorDark: '#08090d', colorLight: '#ffffff' });
  } else if (address) {
    target.innerHTML = '<div style="color:#08090d;font-size:0.85rem;padding:20px;text-align:center;word-break:break-all;"><b>' + address + '</b></div>';
  } else {
    target.innerHTML = '<div style="color:#08090d;font-size:0.85rem;padding:20px;text-align:center;">No wallet address assigned yet</div>';
  }
}

/* ---------- Scanner & QR File Upload ---------- */
let html5QrScanner = null;
let currentTargetField = null;
let currentNextSheet = null;

async function openScanner(fillFieldId, thenOpenSheet) {
  currentTargetField = fillFieldId;
  currentNextSheet = thenOpenSheet;
  openSheet('sheetScan');
  
  const hint = document.getElementById('scanHint');
  const isSecure = window.isSecureContext || location.hostname === 'localhost' || location.hostname === '127.0.0.1';
  
  if (!window.Html5Qrcode) {
    hint.innerHTML = '⚠️ Scanner engine loading. Tap <b>"📷 Upload QR Photo / Image"</b> below to pick a photo!';
    return;
  }

  // Safely stop & clear any previous scanner instance
  if (html5QrScanner) {
    const prev = html5QrScanner;
    html5QrScanner = null;
    try { await prev.stop().catch(() => {}); } catch(e) {}
    try { prev.clear(); } catch(e) {}
  }

  // Check MediaDevices API
  if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
    if (!isSecure) {
      hint.innerHTML = '💡 <b>HTTP Context Notice:</b> Mobile browsers disable live webcams over plain HTTP (<code>http://10.89.63.191</code>).<br><br><b>Solutions:</b><br>1. Tap <b>"📷 Upload QR Photo / Image"</b> below to pick/snap a photo.<br>2. Or enable HTTP camera in Chrome flags: <code>chrome://flags/#unsafely-treat-insecure-origin-as-secure</code>.';
    } else {
      hint.innerHTML = '⚠️ Live camera API not supported by browser. Tap <b>"📷 Upload QR Photo / Image"</b> below!';
    }
    return;
  }

  hint.innerHTML = '⏳ Starting camera stream... Please tap <b>Allow</b> if prompted!';

  html5QrScanner = new Html5Qrcode('scannerRegion');
  const qrConfig = {
    fps: 15,
    qrbox: { width: 220, height: 220 },
    aspectRatio: 1.0,
    videoConstraints: { facingMode: { ideal: "environment" } }
  };

  try {
    await html5QrScanner.start(
      { facingMode: { ideal: 'environment' } },
      qrConfig,
      onQrScanSuccess,
      () => {}
    );
    hint.innerHTML = 'Point camera at a LiquidLink QR code to scan.';
  } catch (err) {
    console.warn('Html5Qrcode start facingMode error, trying camera list:', err);
    try {
      const cameras = await Html5Qrcode.getCameras();
      if (cameras && cameras.length > 0) {
        const backCamera = cameras.find(c => c.label.toLowerCase().includes('back') || c.label.toLowerCase().includes('rear')) || cameras[0];
        await html5QrScanner.start(backCamera.id, qrConfig, onQrScanSuccess, () => {});
        hint.innerHTML = 'Point camera at a LiquidLink QR code to scan.';
        return;
      }
    } catch (camErr) {}
    
    if (err && (err.name === 'NotAllowedError' || err.name === 'PermissionDeniedError')) {
      hint.innerHTML = '⚠️ <b>Camera Permission Denied:</b> Please grant camera access in browser site settings.<br><br>Or tap <b>"📷 Upload QR Photo / Image"</b> below!';
    } else if (!isSecure) {
      hint.innerHTML = '💡 <b>HTTP Context:</b> Live camera stream blocked on plain HTTP.<br><b>Tap "Upload QR Photo" below</b> to select/snap a QR picture!';
    } else {
      hint.innerHTML = '💡 <b>Camera stream error:</b> Camera unavailable.<br><b>Tap "Upload QR Photo" below</b> to scan from gallery!';
    }
  }
}

function onQrScanSuccess(decodedText) {
  if (currentTargetField) {
    const field = document.getElementById(currentTargetField);
    if (field) field.value = decodedText;
  }
  closeSheetAndStopScan();
  if (currentNextSheet) openSheet(currentNextSheet);
  showAlert('Address captured: ' + decodedText.substring(0, 16) + '...');
}

function handleQrImageUpload(input) {
  if (!input.files || input.files.length === 0) return;
  const file = input.files[0];
  const hint = document.getElementById('scanHint');
  hint.textContent = 'Decoding QR image...';

  if (!window.Html5Qrcode) {
    showAlert('QR scanner engine not ready', true);
    return;
  }

  const fileScanner = new Html5Qrcode('scannerRegion');
  fileScanner.scanFile(file, true)
    .then(decodedText => {
      fileScanner.clear();
      onQrScanSuccess(decodedText);
    })
    .catch(err => {
      fileScanner.clear();
      hint.innerHTML = '⚠️ Could not detect a valid QR code in that photo.<br>Please select a clearer image or paste address directly.';
      showAlert('No QR code detected in photo', true);
    });
}

function closeSheetAndStopScan(e) {
  if (e) {
    try { e.preventDefault(); e.stopPropagation(); } catch(_) {}
  }
  // 1. Immediately close the UI modal sheet synchronously
  closeSheet('sheetScan');

  // 2. Stop camera stream safely in background
  if (html5QrScanner) {
    const scanner = html5QrScanner;
    html5QrScanner = null;
    setTimeout(() => {
      try {
        scanner.stop().then(() => {
          try { scanner.clear(); } catch(_) {}
        }).catch(() => {
          try { scanner.clear(); } catch(_) {}
        });
      } catch(_) {
        try { scanner.clear(); } catch(_) {}
      }
    }, 50);
  }
}

/* ---------- API calls & Form Submissions ---------- */
async function postData(url, data, submitBtn) {
  if (submitBtn) {
    submitBtn.disabled = true;
    submitBtn.dataset.origText = submitBtn.textContent;
    submitBtn.textContent = 'Processing...';
  }
  try {
    const resp = await fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data)
    });
    const resData = await resp.json();
    if (!resp.ok) throw new Error(resData.error || 'Request failed with status ' + resp.status);
    return resData;
  } finally {
    if (submitBtn) {
      submitBtn.disabled = false;
      submitBtn.textContent = submitBtn.dataset.origText || 'Submit';
    }
  }
}

async function submitTransfer(e) {
  e.preventDefault();
  const btn = e.target.querySelector('button[type="submit"]');
  try {
    const receiver = document.getElementById('sendReceiver').value.trim();
    const amount = parseInt(document.getElementById('sendAmount').value, 10);
    if (!receiver) { showAlert('Receiver address required', true); return; }
    if (isNaN(amount) || amount <= 0) { showAlert('Valid positive amount required', true); return; }

    const res = await postData('/api/v1/wallet/transfer', { receiver, amount }, btn);
    showAlert('Successfully sent ' + res.amount + ' LQD!');
    setTimeout(() => location.reload(), 1000);
  } catch (err) { showAlert(err.message, true); }
}

async function submitMint(e) {
  e.preventDefault();
  const btn = e.target.querySelector('button[type="submit"]');
  try {
    const address = document.getElementById('mintAddress').value.trim();
    const amount = parseInt(document.getElementById('mintAmount').value, 10);
    if (isNaN(amount) || amount <= 0) { showAlert('Valid positive amount required', true); return; }

    const res = await postData('/api/v1/wallet/mint', { address, amount }, btn);
    showAlert('Minted successfully! New balance: ' + res.new_balance);
    setTimeout(() => location.reload(), 1000);
  } catch (err) { showAlert(err.message, true); }
}

async function submitOpenChannel(e) {
  e.preventDefault();
  const btn = e.target.querySelector('button[type="submit"]');
  try {
    const counterparty = document.getElementById('chanCounterparty').value.trim();
    const amount = parseInt(document.getElementById('chanAmount').value, 10);
    if (!counterparty) { showAlert('Counterparty address required', true); return; }
    if (isNaN(amount) || amount <= 0) { showAlert('Valid deposit amount required', true); return; }

    const res = await postData('/api/v1/channels/open', { counterparty, amount }, btn);
    showAlert('State channel opened: ' + res.id);
    setTimeout(() => location.reload(), 1000);
  } catch (err) { showAlert(err.message, true); }
}

async function submitUpdateChannel(e) {
  e.preventDefault();
  const btn = e.target.querySelector('button[type="submit"]');
  try {
    const channel_id = document.getElementById('chanIdUpdate').value.trim();
    const amount = parseInt(document.getElementById('chanUpdateAmt').value, 10);
    const sending = document.getElementById('chanDirection').value === 'true';
    if (!channel_id) { showAlert('Channel ID required', true); return; }
    if (isNaN(amount) || amount <= 0) { showAlert('Valid amount required', true); return; }

    const res = await postData('/api/v1/channels/update', { channel_id, amount, sending }, btn);
    showAlert('Channel updated (Seq ' + res.sequence + ')');
    setTimeout(() => location.reload(), 1000);
  } catch (err) { showAlert(err.message, true); }
}

async function submitSettleChannel(action) {
  try {
    const channel_id = document.getElementById('chanIdSettle').value.trim();
    if (!channel_id) { showAlert('Channel ID is required', true); return; }
    await postData('/api/v1/channels/settle', { channel_id, action });
    showAlert('Settlement ' + action + ' executed');
    setTimeout(() => location.reload(), 1000);
  } catch (err) { showAlert(err.message, true); }
}

async function submitBridgeRequest(e) {
  e.preventDefault();
  const btn = e.target.querySelector('button[type="submit"]');
  try {
    const receiver = document.getElementById('bridgeReceiver').value.trim();
    const amount = parseInt(document.getElementById('bridgeAmount').value, 10);
    if (!receiver) { showAlert('Receiver required', true); return; }
    if (isNaN(amount) || amount <= 0) { showAlert('Valid amount required', true); return; }

    const res = await postData('/api/v1/bridge/fiat-request', { receiver, amount }, btn);
    showAlert('Bridge request created! UTR: ' + res.mapping.utr);
    setTimeout(() => location.reload(), 1000);
  } catch (err) { showAlert(err.message, true); }
}

async function submitVerifyReceipt(e) {
  e.preventDefault();
  const btn = e.target.querySelector('button[type="submit"]');
  try {
    const utr = document.getElementById('verifyUTR').value.trim();
    const receipt_id = document.getElementById('verifyReceiptID').value.trim();
    if (!utr) { showAlert('UTR required', true); return; }

    const res = await postData('/api/v1/bridge/verify-receipt', { utr, receipt_id }, btn);
    showAlert('Receipt verified and settled for UTR: ' + res.utr);
    setTimeout(() => location.reload(), 1000);
  } catch (err) { showAlert(err.message, true); }
}

/* Close sheet when tapping backdrop */
document.querySelectorAll('.overlay').forEach(ov => {
  ov.addEventListener('click', (e) => {
    if (e.target === ov) {
      if (ov.id === 'sheetScan') closeSheetAndStopScan();
      else closeSheet(ov.id);
    }
  });
});

/* ---------- Enhanced Deterministic Web3 Pixel Art Avatar Generator (Phantom / MetaMask style) ---------- */
const AVATAR_PALETTES = [
  { bg: '#0d0914', main: '#9945FF', accent: '#14F195', detail: '#F9F871' },
  { bg: '#0c0f1d', main: '#ff2a85', accent: '#00f0ff', detail: '#ffb400' },
  { bg: '#0a1128', main: '#1c7ed6', accent: '#20c997', detail: '#ffd43b' },
  { bg: '#081c15', main: '#2d6a4f', accent: '#52b788', detail: '#d8f3dc' },
  { bg: '#1a0c10', main: '#e63946', accent: '#f4a261', detail: '#ffe6a7' },
  { bg: '#140a1d', main: '#7209b7', accent: '#f72585', detail: '#4cc9f0' },
  { bg: '#190019', main: '#85054c', accent: '#e2107b', detail: '#00f5d4' },
  { bg: '#0d1b1e', main: '#3a5a40', accent: '#a3b18a', detail: '#dad7cd' },
  { bg: '#0f0f1b', main: '#ff0055', accent: '#00ccff', detail: '#ffffff' },
  { bg: '#0a0a1a', main: '#5e60ce', accent: '#4ea8de', detail: '#ff70a6' },
  { bg: '#1b1409', main: '#b5838d', accent: '#e07a5f', detail: '#f4f1de' },
  { bg: '#051622', main: '#1ba098', accent: '#deb992', detail: '#80ed99' }
];

function stringToSeed(str) {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    hash = ((hash << 5) - hash) + str.charCodeAt(i);
    hash |= 0;
  }
  return Math.abs(hash);
}

function generateAvatarSVG(address, size) {
  size = size || 36;
  if (!address || address === 'No address assigned') {
    address = "0x0000000000000000000000000000000000000000";
  }
  
  const seed = stringToSeed(address.toLowerCase());
  const palette = AVATAR_PALETTES[seed % AVATAR_PALETTES.length];
  
  // 6x6 pixel grid with 3-column symmetry
  const gridSize = 6;
  const rectSize = size / gridSize;
  let rectsHTML = '';
  
  for (let row = 0; row < gridSize; row++) {
    for (let col = 0; col <= 2; col++) {
      const index = row * 3 + col;
      const val = (seed ^ (index * 2654435761)) % 100;
      
      let fillColor = null;
      
      // Feature logic for cute pixel face (Row 2 = eyes)
      if (row === 2 && col === 1) {
        fillColor = palette.detail; // eye glow
      } else if (val > 35) {
        fillColor = (val % 3 === 0) ? palette.accent : palette.main;
      }
      
      if (fillColor) {
        const x1 = col * rectSize;
        const y = row * rectSize;
        const r = rectSize * 0.15; // smooth corner pixels
        
        rectsHTML += '<rect x="' + x1 + '" y="' + y + '" width="' + (rectSize + 0.3) + '" height="' + (rectSize + 0.3) + '" rx="' + r + '" fill="' + fillColor + '" />';
        
        // Mirror horizontally across center line
        const mirrorCol = gridSize - 1 - col;
        if (col !== mirrorCol) {
          const x2 = mirrorCol * rectSize;
          rectsHTML += '<rect x="' + x2 + '" y="' + y + '" width="' + (rectSize + 0.3) + '" height="' + (rectSize + 0.3) + '" rx="' + r + '" fill="' + fillColor + '" />';
        }
      }
    }
  }
  
  return '<svg width="' + size + '" height="' + size + '" viewBox="0 0 ' + size + ' ' + size + '" xmlns="http://www.w3.org/2000/svg" style="border-radius:50%;overflow:hidden;background:' + palette.bg + ';display:block;box-shadow:0 2px 8px rgba(0,0,0,0.4);border:1px solid rgba(255,255,255,0.08);">' + rectsHTML + '</svg>';
}

async function createNewWalletUI() {
  try {
    const res = await postData('/api/v1/wallet/create', {});
    showAlert('New wallet generated! Address: ' + res.address.substring(0, 16) + '... (1000 LQD initial balance credited)');
    setTimeout(() => location.reload(), 1500);
  } catch (err) {
    showAlert('Failed to create wallet: ' + err.message, true);
  }
}

async function submitOrderbookOrder(e) {
  e.preventDefault();
  const btn = e.target.querySelector('button[type="submit"]');
  try {
    const side = document.getElementById('obOrderSide').value;
    const price = parseFloat(document.getElementById('obOrderPrice').value);
    const amount = parseInt(document.getElementById('obOrderAmount').value, 10);
    if (isNaN(price) || price <= 0) { showAlert('Valid positive limit price required', true); return; }
    if (isNaN(amount) || amount <= 0) { showAlert('Valid positive token amount required', true); return; }

    const res = await postData('/api/v1/orderbook/order', { side, price, amount }, btn);
    showAlert('Order placed successfully! Trades executed: ' + (res.trades ? res.trades.length : 0));
    setTimeout(() => location.reload(), 1200);
  } catch (err) { showAlert(err.message, true); }
}

async function submitStakeShares(e) {
  e.preventDefault();
  const btn = e.target.querySelector('button[type="submit"]');
  try {
    const amount = parseInt(document.getElementById('stakeAmount').value, 10);
    if (isNaN(amount) || amount <= 0) { showAlert('Valid positive amount required', true); return; }

    const res = await postData('/api/v1/stability/stake', { amount }, btn);
    showAlert('Successfully staked ' + amount + ' LQD into shares! Staked shares: ' + res.staked_shares);
    setTimeout(() => location.reload(), 1200);
  } catch (err) { showAlert(err.message, true); }
}

async function submitUnstakeShares() {
  try {
    const amount = parseInt(document.getElementById('stakeAmount').value, 10);
    if (isNaN(amount) || amount <= 0) { showAlert('Valid positive amount required', true); return; }

    const res = await postData('/api/v1/stability/unstake', { amount });
    showAlert('Successfully unstaked ' + amount + ' shares & claimed yield payout!');
    setTimeout(() => location.reload(), 1200);
  } catch (err) { showAlert(err.message, true); }
}

async function submitBuyBonds(e) {
  e.preventDefault();
  const btn = e.target.querySelector('button[type="submit"]');
  try {
    const amount = parseInt(document.getElementById('bondAmount').value, 10);
    if (isNaN(amount) || amount <= 0) { showAlert('Valid positive bond amount required', true); return; }

    const res = await postData('/api/v1/stability/buy-bonds', { amount }, btn);
    showAlert('Successfully purchased ' + amount + ' discounted bonds!');
    setTimeout(() => location.reload(), 1200);
  } catch (err) { showAlert(err.message, true); }
}

async function submitRedeemBonds() {
  try {
    const amount = parseInt(document.getElementById('bondAmount').value, 10);
    if (isNaN(amount) || amount <= 0) { showAlert('Valid positive bond amount required', true); return; }

    const res = await postData('/api/v1/stability/redeem-bonds', { amount });
    showAlert('Successfully redeemed ' + amount + ' bonds + bonus payout!');
    setTimeout(() => location.reload(), 1200);
  } catch (err) { showAlert(err.message, true); }
}

function initAvatars() {
  const address = {{ if .Address }}"{{ .Address }}"{{ else }}""{{ end }};
  
  const topbarBtn = document.getElementById('topbarAvatarBtn');
  if (topbarBtn) topbarBtn.innerHTML = generateAvatarSVG(address, 36);
  
  const profileAvatar = document.getElementById('profileAvatarLg');
  if (profileAvatar) profileAvatar.innerHTML = generateAvatarSVG(address, 72);

  const heroAvatar = document.getElementById('heroAddressAvatar');
  if (heroAvatar) heroAvatar.innerHTML = generateAvatarSVG(address, 22);
}

document.addEventListener('DOMContentLoaded', initAvatars);
</script>
</body>
</html>`

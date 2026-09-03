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
  padding-bottom: calc(var(--nav-h) + 60px);
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

/* ---------- Welcome Overlay & Onboarding ---------- */
.welcome-overlay {
  position: fixed; inset: 0; z-index: 100;
  background: rgba(8, 9, 13, 0.94);
  backdrop-filter: blur(24px); -webkit-backdrop-filter: blur(24px);
  display: flex; align-items: center; justify-content: center; padding: 20px;
}
.welcome-card {
  background: var(--surface); border: 1px solid var(--border);
  border-radius: 20px; width: 100%; max-width: 440px; padding: 32px 24px;
  text-align: center; box-shadow: 0 20px 50px rgba(0,0,0,0.6);
  animation: welcomePop 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}
@keyframes welcomePop { from { opacity: 0; transform: scale(0.92) translateY(20px); } to { opacity: 1; transform: scale(1) translateY(0); } }
.welcome-badge { display: inline-block; font-size: 0.72rem; font-weight: 700; color: var(--accent); background: var(--accent-soft); padding: 4px 10px; border-radius: 20px; text-transform: uppercase; letter-spacing: 0.5px; margin-bottom: 16px; }
.welcome-logo { margin: 0 auto 16px; display: flex; align-items: center; justify-content: center; }
.welcome-title { font-family: var(--font-display); font-size: 1.6rem; font-weight: 700; margin: 0 0 8px; color: var(--text); }
.welcome-sub { font-size: 0.85rem; color: var(--text-dim); margin: 0 0 20px; line-height: 1.45; }
.welcome-features { display: flex; flex-direction: column; gap: 10px; text-align: left; margin-bottom: 24px; background: var(--surface-2); padding: 14px; border-radius: 12px; border: 1px solid var(--border-soft); }
.feat-item { display: flex; align-items: center; gap: 10px; font-size: 0.82rem; color: var(--text); }
.feat-icon { font-size: 1rem; width: 24px; text-align: center; }

/* ---------- Celebration Modal & Confetti ---------- */
.celebration-overlay {
  position: fixed; inset: 0; z-index: 110;
  background: rgba(8, 9, 13, 0.92);
  backdrop-filter: blur(20px);
  display: flex; align-items: center; justify-content: center; padding: 20px;
}
.celebration-card {
  background: var(--surface); border: 1px solid var(--border);
  border-radius: 24px; width: 100%; max-width: 400px; padding: 28px 24px;
  text-align: center; box-shadow: 0 24px 60px rgba(0,0,0,0.7); position: relative; z-index: 2;
  animation: celPop 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}
@keyframes celPop { from { opacity: 0; transform: scale(0.85); } to { opacity: 1; transform: scale(1); } }
.celebration-icon { font-size: 3rem; margin-bottom: 8px; }
.cel-avatar { width: 64px; height: 64px; border-radius: 50%; margin: 12px auto; display: flex; align-items: center; justify-content: center; overflow: hidden; border: 2px solid var(--accent); }
.cel-address { font-family: var(--font-mono); font-size: 0.78rem; color: var(--text-dim); background: var(--surface-2); padding: 6px 12px; border-radius: 20px; word-break: break-all; margin-bottom: 16px; border: 1px solid var(--border-soft); }
.cel-balance-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; margin-top: 12px; }
.cel-bal-card { background: var(--surface-2); padding: 10px; border-radius: 10px; border: 1px solid var(--border-soft); }
.cel-bal-card .k { font-size: 0.72rem; color: var(--text-dim); }
.cel-bal-card .v { font-size: 0.95rem; font-weight: 700; margin-top: 2px; }

.confetti-container { position: absolute; inset: 0; overflow: hidden; pointer-events: none; z-index: 1; }
.confetti-particle { position: absolute; top: -10px; width: 8px; height: 8px; border-radius: 2px; opacity: 0.8; animation: confettiFall linear infinite; }
@keyframes confettiFall {
  0% { transform: translateY(0) rotate(0deg); opacity: 1; }
  100% { transform: translateY(100vh) rotate(720deg); opacity: 0; }
}
</style>
</head>
<body>

<div id="alertBox"></div>

<!-- Welcome Screen Overlay (MetaMask-style Onboarding) -->
<div class="welcome-overlay" id="welcomeOverlay" style="display:none;">
  <div class="welcome-card">
    <div class="welcome-badge">LiquidLink Web3</div>
    <div class="welcome-logo">
      <svg width="64" height="64" viewBox="0 0 40 40" fill="none">
        <rect width="40" height="40" rx="12" fill="url(#welcGrad)"/>
        <path d="M12 28L20 12L28 28H23L20 22L17 28H12Z" fill="white"/>
        <circle cx="20" cy="18" r="3" fill="#6d9bee"/>
        <defs><linearGradient id="welcGrad" x1="0" y1="0" x2="40" y2="40"><stop stop-color="#4F6EF7"/><stop offset="1" stop-color="#1E2B7A"/></linearGradient></defs>
      </svg>
    </div>
    <h1 class="welcome-title">Welcome to LiquidLink</h1>
    <p class="welcome-sub">Decentralized Peer-to-Peer Settlement & Seigniorage Liquidity Protocol</p>

    <div class="welcome-features">
      <div class="feat-item"><span class="feat-icon">⚡</span><span>Instant P2P State Channel Payments</span></div>
      <div class="feat-item"><span class="feat-icon">🛡️</span><span>Algorithmic Seigniorage INR Peg Stability</span></div>
      <div class="feat-item"><span class="feat-icon">📈</span><span>Real-Time Limit Orderbook Exchange</span></div>
    </div>

    <div class="welcome-actions">
      <button class="btn btn-lg btn-primary" onclick="createNewWalletFromWelcome()" style="width:100%;">Create New Wallet</button>
      <button class="btn btn-ghost" onclick="closeWelcomeOverlay()" style="width:100%; margin-top:8px;">Explore Node Default Wallet</button>
    </div>
  </div>
</div>

<!-- Celebration Animation Modal -->
<div class="celebration-overlay" id="celebrationModal" style="display:none;">
  <div class="confetti-container" id="confettiContainer"></div>
  <div class="celebration-card">
    <div class="celebration-icon">🥳</div>
    <h2 style="margin:0 0 4px; font-size:1.4rem;">Wallet Created!</h2>
    <p style="font-size:0.85rem; color:var(--text-dim); margin-bottom:14px;">Your new Web3 wallet session is active and funded on LiquidLink testnet.</p>
    
    <div class="cel-avatar" id="celAvatar"></div>
    <div class="cel-address" id="celAddressText"></div>

    <div class="cel-balance-grid">
      <div class="cel-bal-card">
        <div class="k">Starting LQD</div>
        <div class="v" id="celLqdVal">1000 LQD</div>
      </div>
      <div class="cel-bal-card">
        <div class="k">Starting Fiat</div>
        <div class="v" id="celInrVal" style="color:var(--success);">₹1,000.00 INR</div>
      </div>
    </div>

    <button class="btn btn-primary" onclick="dismissCelebrationModal()" style="width:100%; margin-top:16px;">Launch Dashboard 🚀</button>
  </div>
</div>

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
      <div class="hero-balance" id="heroTotalBalance">{{ .Balance }}<small>LQD</small></div>
      <div style="font-size:1.05rem; font-weight:600; color:var(--success); margin-top:2px;" id="heroINRBalance">₹{{ printf "%.2f" .INRBalance }} INR</div>
      <div class="addr-chip">
        <span id="heroAddressAvatar" style="display:inline-flex;align-items:center;margin-right:2px;"></span>
        <span class="a" id="walletAddress">{{ if .Address }}{{ .Address }}{{ else }}No address assigned{{ end }}</span>
        <button class="icon-btn-mini" onclick="copyText(getActiveWallet(), this)" aria-label="Copy address">
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
      </div>
    </div>
  </div>

  <!-- Secondary metrics -->
  <div class="stat-strip">
    <div class="stat-card"><div class="stat-title">Status</div><div class="status-pill"><span class="dot"></span>Online</div></div>
    <div class="stat-card"><div class="stat-title">Total supply</div><div class="stat-val" id="statTotalSupply">{{ .TotalSupply }}</div></div>
    <div class="stat-card"><div class="stat-title">INR Reserve Pool</div><div class="stat-val" style="color:var(--success);" id="statInrPool">₹{{ printf "%.2f" .TotalINRPool }} INR</div></div>
    <div class="stat-card"><div class="stat-title">Block height</div><div class="stat-val" id="statBlockHeight">{{ .BlockHeight }}</div></div>
    <div class="stat-card"><div class="stat-title">Peers</div><div class="stat-val" id="statPeerCount">{{ .PeerCount }}</div></div>
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
        <h2>LQD / INR Orderbook <span class="count" id="obClearingPrice">Clearing Price: ₹{{ if .StabilityStatus }}{{ printf "%.2f" .StabilityStatus.Price }}{{ else }}1.00{{ end }} INR</span></h2>
      </div>

      <!-- Price History Chart -->
      <div style="background:var(--surface-2); border-radius:12px; padding:16px; margin:16px 16px 0; border:1px solid var(--border);">
        <h3 style="font-size:0.95rem; margin-bottom:12px; color:var(--text);">LQD / INR Exchange Rate & Peg Stability (Price Over Time)</h3>
        <div id="priceChartWrap" style="position:relative; width:100%; height:180px;"></div>
      </div>

      <div style="display:grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap:16px; padding:16px;">
        <!-- Depth Bids / Asks -->
        <div style="background:var(--surface-2); border-radius:12px; padding:16px; border:1px solid var(--border);">
          <h3 style="font-size:0.95rem; margin-bottom:12px; color:var(--text);">Orderbook Depth (INR)</h3>
          <div style="display:grid; grid-template-columns: 1fr 1fr; gap:12px;">
            <div>
              <h4 style="font-size:0.8rem; color:var(--success); margin-bottom:6px;">Bids (Buy LQD)</h4>
              <div id="obBidsContainer" style="font-family:var(--font-mono); font-size:0.8rem;">
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
              <div id="obAsksContainer" style="font-family:var(--font-mono); font-size:0.8rem;">
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
          <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:12px; flex-wrap:wrap; gap:6px;">
            <h3 style="font-size:0.95rem; color:var(--text); margin:0;">Place Limit Order</h3>
            <span style="font-size:0.78rem; color:var(--text-dim);">Available: <strong id="obAvailableINR" style="color:var(--success);">₹{{ printf "%.2f" .INRBalance }} INR</strong> | Bank Pool: <strong id="obInrPoolVal" style="color:var(--accent-strong);">₹{{ printf "%.2f" .TotalINRPool }} INR</strong></span>
          </div>
          <form onsubmit="submitOrderbookOrder(event)">
            <div class="field">
              <label>Order Type</label>
              <select id="obOrderSide" class="input" style="width:100%;" onchange="updateSuggestedOrderPrice()">
                <option value="buy">Buy LQD (Bid in INR)</option>
                <option value="sell">Sell LQD (Ask in INR)</option>
              </select>
            </div>
            <div class="field" style="margin-top:10px;">
              <label>Limit Price in INR (₹)</label>
              <input type="number" id="obOrderPrice" step="0.01" value="1.01" class="input" placeholder="e.g. 1.01" required>
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
      <div id="obTradesContainer" style="padding:16px; border-top:1px solid var(--border);">
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

      <!-- User Active Open Orders -->
      <div style="padding:16px; border-top:1px solid var(--border);">
        <h3 style="font-size:0.9rem; color:var(--text-dim); margin-bottom:10px;">My Orders (Active & History)</h3>
        <div id="userOrdersContainer">
          <div class="empty-row">No limit orders for active wallet.</div>
        </div>
      </div>
    </div>
  </div>

  <!-- Seigniorage & Holdings -->
  <div id="segStability" class="panel">
    <div class="section-card">
      <div class="section-head">
        <h2>Seigniorage Peg & Holdings <span class="badge {{ if eq (printf "%v" .StabilityStatus.State) "expansion" }}badge-active{{ else }}badge-pending{{ end }}" id="stabilityStateBadge">{{ .StabilityStatus.State }}</span></h2>
      </div>

      <!-- Monetary Metrics Multi-Line Chart -->
      <div style="background:var(--surface-2); border-radius:12px; padding:16px; margin:16px 16px 0; border:1px solid var(--border);">
        <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:12px; flex-wrap:wrap; gap:8px;">
          <h3 style="font-size:0.95rem; color:var(--text); margin:0;">Protocol Monetary Mechanics (Supply, Shares & Bonds Over Time)</h3>
          <div style="display:flex; gap:12px; font-size:0.75rem;">
            <span style="color:#4d84e2;">● Total Supply</span>
            <span style="color:#b98f45;">● Staked Shares</span>
            <span style="color:#c96a6a;">● Bonds Held</span>
          </div>
        </div>
        <div id="metricsChartWrap" style="position:relative; width:100%; height:180px;"></div>
      </div>

      <div style="padding:16px; display:grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap:12px; border-bottom:1px solid var(--border);">
        <div style="background:var(--surface-2); padding:12px; border-radius:10px;">
          <div style="font-size:0.75rem; color:var(--text-dim);">Target Peg</div>
          <div style="font-size:1.1rem; font-weight:700; color:var(--success);">₹1.00 INR</div>
        </div>
        <div style="background:var(--surface-2); padding:12px; border-radius:10px;">
          <div style="font-size:0.75rem; color:var(--text-dim);">Current Market Price</div>
          <div style="font-size:1.1rem; font-weight:700; color:var(--accent-strong);" id="marketPriceVal">₹{{ if .StabilityStatus }}{{ printf "%.2f" .StabilityStatus.Price }}{{ else }}1.00{{ end }} INR</div>
        </div>
        <div style="background:var(--surface-2); padding:12px; border-radius:10px;">
          <div style="font-size:0.75rem; color:var(--text-dim);">Staking Yield APY</div>
          <div style="font-size:1.1rem; font-weight:700; color:var(--amber);" id="stakingApyVal">{{ if .StabilityStatus }}{{ printf "%.1f" .StabilityStatus.CurrentAPY }}{{ else }}14.5{{ end }}%</div>
        </div>
        <div style="background:var(--surface-2); padding:12px; border-radius:10px;">
          <div style="font-size:0.75rem; color:var(--text-dim);">Protocol Staked Shares</div>
          <div style="font-size:1.1rem; font-weight:700; color:#b98f45;" id="totalStakedSharesVal">{{ if .StabilityStatus }}{{ .StabilityStatus.TotalStakedShares }}{{ else }}50000{{ end }} Shares</div>
        </div>
        <div style="background:var(--surface-2); padding:12px; border-radius:10px;">
          <div style="font-size:0.75rem; color:var(--text-dim);">Protocol Outstanding Bonds</div>
          <div style="font-size:1.1rem; font-weight:700; color:#c96a6a;" id="totalOutstandingBondsVal">{{ if .StabilityStatus }}{{ .StabilityStatus.TotalOutstandingBonds }}{{ else }}12000{{ end }} Bonds</div>
        </div>
        <div style="background:var(--surface-2); padding:12px; border-radius:10px;">
          <div style="font-size:0.75rem; color:var(--text-dim);">Supply Converted to Bonds & Shares</div>
          <div style="font-size:1.1rem; font-weight:700; color:var(--primary);" id="totalSupplyConvertedVal">{{ if .StabilityStatus }}{{ .StabilityStatus.TotalStakedShares }}{{ else }}62000{{ end }} LQD</div>
        </div>
        <div style="background:var(--surface-2); padding:12px; border-radius:10px;">
          <div style="font-size:0.75rem; color:var(--text-dim);">My Staked Shares</div>
          <div style="font-size:1.1rem; font-weight:700;" id="userStakedShares">{{ .StakedShares }} Shares</div>
        </div>
        <div style="background:var(--surface-2); padding:12px; border-radius:10px;">
          <div style="font-size:0.75rem; color:var(--text-dim);">My Bonds Held</div>
          <div style="font-size:1.1rem; font-weight:700;" id="userBondsHeld">{{ .BondsHeld }} Bonds</div>
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
        <h2>State channels <span class="count" id="channelsCount">({{ len .Channels }})</span></h2>
        <div class="head-actions">
          <button class="btn btn-sm btn-secondary" onclick="openSheet('sheetChannel')">Manage</button>
          <button class="btn btn-sm btn-ghost" onclick="openSheet('sheetChannel')">+ Open</button>
        </div>
      </div>
      <div id="channelsListContainer">
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
  </div>

  <!-- Bridge -->
  <div id="segBridge" class="panel">
    <div class="section-card">
      <div class="section-head">
        <h2>Fiat bridge <span class="count" id="bridgeCount">({{ len .UTRMappings }})</span></h2>
        <div class="head-actions">
          <button class="btn btn-sm btn-ghost" onclick="openSheet('sheetBridge')">+ Request</button>
        </div>
      </div>
      <div id="bridgeListContainer">
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
  </div>

  <!-- Activity / Blocks -->
  <div id="segActivity" class="panel">
    <div class="section-card">
      <div class="section-head">
        <h2>Recent blocks <span class="count" id="blocksCount">({{ len .RecentBlocks }})</span></h2>
      </div>
      <div id="blocksListContainer">
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
  <div class="qr-address" id="sheetReceiveAddress">{{ if .Address }}{{ .Address }}{{ else }}No address assigned{{ end }}</div>
  <button class="btn" onclick="copyText(getActiveWallet(), this)">Copy address</button>
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
  <div class="sheet-head"><h3>Devnet Faucet / Mint</h3><button class="sheet-close" onclick="closeSheet('sheetMint')"><svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M6 6l12 12M18 6L6 18" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg></button></div>
  <div style="font-size:0.8rem; color:#f59e0b; margin-bottom:12px; padding:8px 12px; background:rgba(245,158,11,0.1); border:1px solid rgba(245,158,11,0.3); border-radius:6px; display:flex; align-items:center; gap:8px;">
    <span>⚡ <strong>Genesis / Devnet Faucet Tool:</strong> Minting is an administrative tool for testnet genesis allocation. Real end-users acquire tokens via transfers, trades, or fiat bridging.</span>
  </div>
  <form onsubmit="submitMint(event)">
    <div class="form-group">
      <label for="mintAddress">Target address</label>
      <input type="text" id="mintAddress" placeholder="Target wallet address (blank for active wallet)">
    </div>
    <div class="form-group">
      <label for="mintAmount">Amount</label>
      <input type="number" id="mintAmount" min="1" placeholder="0" required>
    </div>
    <button type="submit" class="btn">Mint tokens</button>
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
    <div class="btn-row" style="flex-wrap:wrap; gap:8px;">
      <button class="btn btn-secondary" onclick="submitSettleChannel('initiate')" style="flex:1;">Initiate</button>
      <button class="btn btn-secondary" onclick="submitChallengeChannel()" style="flex:1;">Challenge (S+1)</button>
      <button class="btn" onclick="submitSettleChannel('finalize')" style="flex:1;">Finalize</button>
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
  <div class="sheet-head"><h3>Profile & Accounts</h3><button class="sheet-close" onclick="closeSheet('sheetProfile')"><svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M6 6l12 12M18 6L6 18" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg></button></div>
  <div class="profile-avatar-lg" id="profileAvatarLg"></div>
  <div class="profile-item"><span class="k">Node ID</span><span class="v">{{ .NodeID }}</span></div>
  <div class="profile-item"><span class="k">Port</span><span class="v">{{ .Port }}</span></div>
  <div class="profile-item"><span class="k">Peers</span><span class="v">{{ .PeerCount }}</span></div>

  <div style="margin-top:16px; font-size:0.85rem; font-weight:600; color:var(--text); text-transform:uppercase; letter-spacing:0.5px;">Active Account / Local Wallets</div>
  <div id="profileWalletsList" style="margin-top:8px; display:flex; flex-direction:column; gap:8px; max-height:180px; overflow-y:auto;"></div>

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
  const address = getActiveWallet();
  const addrEl = document.getElementById('sheetReceiveAddress');
  if (addrEl) addrEl.textContent = address || 'No address assigned';
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
async function fetchData(url) {
  const resp = await fetch(url);
  const resData = await resp.json();
  if (!resp.ok) throw new Error(resData.error || 'Request failed with status ' + resp.status);
  return resData;
}

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

    const sender = getActiveWallet();
    const res = await postData('/api/v1/wallet/transfer', { sender, receiver, amount }, btn);
    showAlert('Successfully sent ' + res.amount + ' LQD!');
    refreshDashboardData();
  } catch (err) { showAlert(err.message, true); }
}

async function submitMint(e) {
  e.preventDefault();
  const btn = e.target.querySelector('button[type="submit"]');
  try {
    let address = document.getElementById('mintAddress').value.trim();
    if (!address) address = getActiveWallet();
    const amount = parseInt(document.getElementById('mintAmount').value, 10);
    if (isNaN(amount) || amount <= 0) { showAlert('Valid positive amount required', true); return; }

    const res = await postData('/api/v1/wallet/mint', { address, amount }, btn);
    showAlert('Minted successfully! New balance: ' + res.new_balance);
    refreshDashboardData();
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
    refreshDashboardData();
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
    refreshDashboardData();
  } catch (err) { showAlert(err.message, true); }
}

async function submitSettleChannel(action) {
  try {
    const channel_id = document.getElementById('chanIdSettle').value.trim();
    if (!channel_id) { showAlert('Channel ID is required', true); return; }
    await postData('/api/v1/channels/settle', { channel_id, action });
    showAlert('Settlement ' + action + ' executed');
    refreshDashboardData();
  } catch (err) { showAlert(err.message, true); }
}

async function submitChallengeChannel() {
  try {
    const channel_id = document.getElementById('chanIdSettle').value.trim();
    if (!channel_id) { showAlert('Channel ID is required', true); return; }
    await postData('/api/v1/channels/challenge', { channel_id });
    showAlert('Dispute challenged with higher sequence co-signature!');
    refreshDashboardData();
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
    refreshDashboardData();
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
    refreshDashboardData();
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
  const currentWallet = localStorage.getItem('liquidlink_active_wallet');
  if (currentWallet) {
    const confirmReplace = confirm(
      'A wallet session already exists in browser storage:\n' +
      currentWallet + '\n\n' +
      'Creating a new wallet will delete/reset your current active wallet session. Do you want to delete it and create a new wallet?'
    );
    if (!confirmReplace) {
      return;
    }
  }

  try {
    const res = await postData('/api/v1/wallet/create', {});
    localStorage.setItem('liquidlink_active_wallet', res.address);
    localStorage.setItem('liquidlink_known_wallets', JSON.stringify([res.address]));
    await loadActiveWalletUI(res.address, res.balance || 1000, res.inr_balance || 1000);
    refreshDashboardData();
    celebrateWalletCreation(res.address, res.balance || 1000, res.inr_balance || 1000);
  } catch (err) {
    showAlert('Failed to create wallet: ' + err.message, true);
  }
}

async function submitOrderbookOrder(e) {
  e.preventDefault();
  const btn = e.target.querySelector('button[type="submit"]');
  try {
    const user = getActiveWallet();
    const side = document.getElementById('obOrderSide').value;
    const price = parseFloat(document.getElementById('obOrderPrice').value);
    const amount = parseInt(document.getElementById('obOrderAmount').value, 10);
    if (isNaN(price) || price <= 0) { showAlert('Valid positive limit price required', true); return; }
    if (isNaN(amount) || amount <= 0) { showAlert('Valid positive token amount required', true); return; }

    const res = await postData('/api/v1/orderbook/order', { user, side, price, amount }, btn);
    const tradesCount = (res.trades && Array.isArray(res.trades)) ? res.trades.length : 0;
    if (tradesCount > 0) {
      showAlert('Trade Executed! Executed ' + tradesCount + ' trade(s). LQD balance updated!');
    } else {
      showAlert('Limit order placed! Resting on orderbook at ₹' + price.toFixed(2) + ' INR (INR reserved until matched).');
    }
    refreshDashboardData();
  } catch (err) { showAlert(err.message, true); }
}

function updateSuggestedOrderPrice() {
  const side = document.getElementById('obOrderSide').value;
  const priceInput = document.getElementById('obOrderPrice');
  if (side === 'buy') {
    priceInput.value = '1.01';
  } else {
    priceInput.value = '0.99';
  }
}

async function submitStakeShares(e) {
  e.preventDefault();
  const btn = e.target.querySelector('button[type="submit"]');
  try {
    const user = getActiveWallet();
    const amount = parseInt(document.getElementById('stakeAmount').value, 10);
    if (isNaN(amount) || amount <= 0) { showAlert('Valid positive amount required', true); return; }

    const res = await postData('/api/v1/stability/stake', { user, amount }, btn);
    showAlert('Successfully staked ' + amount + ' LQD into shares! Staked shares: ' + res.staked_shares);
    refreshDashboardData();
  } catch (err) { showAlert(err.message, true); }
}

async function submitUnstakeShares() {
  try {
    const user = getActiveWallet();
    const amount = parseInt(document.getElementById('stakeAmount').value, 10);
    if (isNaN(amount) || amount <= 0) { showAlert('Valid positive amount required', true); return; }

    const res = await postData('/api/v1/stability/unstake', { user, amount });
    showAlert('Successfully unstaked ' + amount + ' shares & claimed yield payout!');
    refreshDashboardData();
  } catch (err) { showAlert(err.message, true); }
}

async function submitBuyBonds(e) {
  e.preventDefault();
  const btn = e.target.querySelector('button[type="submit"]');
  try {
    const user = getActiveWallet();
    const amount = parseInt(document.getElementById('bondAmount').value, 10);
    if (isNaN(amount) || amount <= 0) { showAlert('Valid positive bond amount required', true); return; }

    const res = await postData('/api/v1/stability/buy-bonds', { user, amount }, btn);
    showAlert('Successfully purchased ' + amount + ' discounted bonds!');
    refreshDashboardData();
  } catch (err) { showAlert(err.message, true); }
}

async function submitRedeemBonds() {
  try {
    const user = getActiveWallet();
    const amount = parseInt(document.getElementById('bondAmount').value, 10);
    if (isNaN(amount) || amount <= 0) { showAlert('Valid positive bond amount required', true); return; }

    const res = await postData('/api/v1/stability/redeem-bonds', { user, amount });
    showAlert('Successfully redeemed ' + amount + ' bonds + bonus payout!');
    refreshDashboardData();
  } catch (err) { showAlert(err.message, true); }
}

/* ---------- Onboarding Welcome & Celebration Modal ---------- */
function checkFirstTimeWelcome() {
  const active = localStorage.getItem('liquidlink_active_wallet');
  if (!active) {
    const welcome = document.getElementById('welcomeOverlay');
    if (welcome) welcome.style.display = 'flex';
  }
}

function closeWelcomeOverlay() {
  const welcome = document.getElementById('welcomeOverlay');
  if (welcome) welcome.style.display = 'none';
}

async function createNewWalletFromWelcome() {
  closeWelcomeOverlay();
  try {
    const res = await postData('/api/v1/wallet/create', {});
    localStorage.setItem('liquidlink_active_wallet', res.address);
    localStorage.setItem('liquidlink_known_wallets', JSON.stringify([res.address]));
    await loadActiveWalletUI(res.address, res.balance || 1000, res.inr_balance || 1000);
    refreshDashboardData();
    celebrateWalletCreation(res.address, res.balance || 1000, res.inr_balance || 1000);
  } catch (err) {
    showAlert('Failed to create wallet: ' + err.message, true);
  }
}

function celebrateWalletCreation(address, lqdBal, inrBal) {
  const modal = document.getElementById('celebrationModal');
  if (!modal) return;

  const addrEl = document.getElementById('celAddressText');
  if (addrEl) addrEl.textContent = address;

  const avEl = document.getElementById('celAvatar');
  if (avEl) avEl.innerHTML = generateAvatarSVG(address, 64);

  const lqdEl = document.getElementById('celLqdVal');
  if (lqdEl) lqdEl.textContent = (lqdBal || 1000) + ' LQD';

  const inrEl = document.getElementById('celInrVal');
  if (inrEl) inrEl.textContent = '₹' + ((inrBal || 1000).toFixed(2)) + ' INR';

  modal.style.display = 'flex';

  const container = document.getElementById('confettiContainer');
  if (container) {
    container.innerHTML = '';
    const colors = ['#4F6EF7', '#45a374', '#b98f45', '#6d9bee', '#e85d04', '#9d4edd'];
    for (let i = 0; i < 40; i++) {
      const p = document.createElement('div');
      p.className = 'confetti-particle';
      p.style.left = Math.random() * 100 + '%';
      p.style.backgroundColor = colors[Math.floor(Math.random() * colors.length)];
      p.style.animationDelay = (Math.random() * 0.8) + 's';
      p.style.animationDuration = (1.5 + Math.random() * 1.5) + 's';
      container.appendChild(p);
    }
  }
}

function dismissCelebrationModal() {
  const modal = document.getElementById('celebrationModal');
  if (modal) modal.style.display = 'none';
}

/* ---------- Standalone Interactive SVG Line Charts ---------- */
function renderPriceChart(history) {
  const container = document.getElementById('priceChartWrap');
  if (!container) return;

  if (!history || history.length === 0) {
    container.innerHTML = '<div style="display:flex;align-items:center;justify-content:center;height:100%;color:var(--text-faint);font-size:0.85rem;">Collecting historical price points...</div>';
    return;
  }

  const width = container.clientWidth || 500;
  const height = 180;
  const padding = { top: 20, right: 20, bottom: 25, left: 45 };
  const graphW = width - padding.left - padding.right;
  const graphH = height - padding.top - padding.bottom;

  const prices = history.map(h => h.price);
  let minP = Math.min(...prices, 0.95);
  let maxP = Math.max(...prices, 1.05);
  if (minP === maxP) { minP -= 0.05; maxP += 0.05; }

  const getX = (idx) => padding.left + (idx / Math.max(history.length - 1, 1)) * graphW;
  const getY = (val) => padding.top + graphH - ((val - minP) / (maxP - minP)) * graphH;

  let pathD = '';
  history.forEach((h, idx) => {
    const x = getX(idx);
    const y = getY(h.price);
    pathD += (idx === 0 ? 'M' : 'L') + x.toFixed(1) + ' ' + y.toFixed(1);
  });

  const areaD = pathD + ' L' + getX(history.length - 1).toFixed(1) + ' ' + (padding.top + graphH) + ' L' + getX(0).toFixed(1) + ' ' + (padding.top + graphH) + ' Z';
  const pegY = getY(1.00);

  let dotsHTML = '';
  history.forEach((h, idx) => {
    const x = getX(idx).toFixed(1);
    const y = getY(h.price).toFixed(1);
    const priceStr = h.price.toFixed(2);
    dotsHTML += '<circle cx="' + x + '" cy="' + y + '" r="4" fill="var(--accent-strong)" stroke="var(--bg)" stroke-width="2"><title>₹' + priceStr + ' INR</title></circle>';
  });

  const svg = '<svg width="100%" height="' + height + '" viewBox="0 0 ' + width + ' ' + height + '" preserveAspectRatio="none" style="overflow:visible;">' +
    '<defs><linearGradient id="pGrad" x1="0" y1="0" x2="0" y2="1"><stop offset="0%" stop-color="#4d84e2" stop-opacity="0.3"/><stop offset="100%" stop-color="#4d84e2" stop-opacity="0.0"/></linearGradient></defs>' +
    '<line x1="' + padding.left + '" y1="' + pegY.toFixed(1) + '" x2="' + (width - padding.right) + '" y2="' + pegY.toFixed(1) + '" stroke="#b98f45" stroke-dasharray="4 4" stroke-width="1.5"/>' +
    '<text x="' + (width - padding.right) + '" y="' + (pegY - 4).toFixed(1) + '" fill="#b98f45" font-size="10" text-anchor="end">Target Peg ₹1.00</text>' +
    '<path d="' + areaD + '" fill="url(#pGrad)"/>' +
    '<path d="' + pathD + '" fill="none" stroke="var(--accent-strong)" stroke-width="2.5" stroke-linecap="round"/>' +
    dotsHTML +
    '<text x="' + (padding.left - 6) + '" y="' + (padding.top + 10) + '" fill="var(--text-dim)" font-size="10" text-anchor="end">₹' + maxP.toFixed(2) + '</text>' +
    '<text x="' + (padding.left - 6) + '" y="' + (padding.top + graphH) + '" fill="var(--text-dim)" font-size="10" text-anchor="end">₹' + minP.toFixed(2) + '</text>' +
    '</svg>';

  container.innerHTML = svg;
}

function renderMetricsChart(history) {
  const container = document.getElementById('metricsChartWrap');
  if (!container) return;

  if (!history || history.length === 0) {
    container.innerHTML = '<div style="display:flex;align-items:center;justify-content:center;height:100%;color:var(--text-faint);font-size:0.85rem;">Collecting monetary metric points...</div>';
    return;
  }

  const width = container.clientWidth || 500;
  const height = 180;
  const padding = { top: 20, right: 20, bottom: 25, left: 55 };
  const graphW = width - padding.left - padding.right;
  const graphH = height - padding.top - padding.bottom;

  const supplies = history.map(h => (typeof h.total_supply === 'number' ? h.total_supply : 1000000));
  const shares = history.map(h => (typeof h.total_shares === 'number' ? h.total_shares : 50000));
  const bonds = history.map(h => (typeof h.total_bonds === 'number' ? h.total_bonds : 12000));

  const validVals = [...supplies, ...shares, ...bonds].filter(v => typeof v === 'number' && Number.isFinite(v));
  let minV = validVals.length > 0 ? Math.min(...validVals) : 0;
  let maxV = validVals.length > 0 ? Math.max(...validVals) : 100;
  if (minV === maxV) { minV = 0; maxV = Math.max(maxV, 100); }

  const getX = (idx) => padding.left + (idx / Math.max(history.length - 1, 1)) * graphW;
  const getY = (val) => padding.top + graphH - ((val - minV) / (maxV - minV)) * graphH;

  const buildPath = (vals) => {
    let d = '';
    vals.forEach((v, idx) => {
      d += (idx === 0 ? 'M' : 'L') + getX(idx).toFixed(1) + ' ' + getY(v).toFixed(1);
    });
    return d;
  };

  const pathSupply = buildPath(supplies);
  const pathShares = buildPath(shares);
  const pathBonds = buildPath(bonds);

  const svg = '<svg width="100%" height="' + height + '" viewBox="0 0 ' + width + ' ' + height + '" preserveAspectRatio="none" style="overflow:visible;">' +
    '<path d="' + pathSupply + '" fill="none" stroke="#4d84e2" stroke-width="2" stroke-linecap="round"/>' +
    '<path d="' + pathShares + '" fill="none" stroke="#b98f45" stroke-width="2" stroke-dasharray="3 3"/>' +
    '<path d="' + pathBonds + '" fill="none" stroke="#c96a6a" stroke-width="2"/>' +
    '<text x="' + (padding.left - 6) + '" y="' + (padding.top + 10) + '" fill="var(--text-dim)" font-size="10" text-anchor="end">' + Math.round(maxV) + '</text>' +
    '<text x="' + (padding.left - 6) + '" y="' + (padding.top + graphH) + '" fill="var(--text-dim)" font-size="10" text-anchor="end">' + Math.round(minV) + '</text>' +
    '</svg>';

  container.innerHTML = svg;
}

function getActiveWallet() {
  return localStorage.getItem('liquidlink_active_wallet') || '{{ .Address }}';
}

function getKnownWallets() {
  try {
    const list = JSON.parse(localStorage.getItem('liquidlink_known_wallets'));
    if (Array.isArray(list) && list.length > 0) return list;
  } catch (e) {}
  return ['{{ .Address }}'];
}

function saveKnownWallet(addr) {
  if (!addr) return;
  const known = getKnownWallets();
  if (!known.includes(addr)) {
    known.push(addr);
    localStorage.setItem('liquidlink_known_wallets', JSON.stringify(known));
  }
}

async function loadActiveWalletUI(targetAddr, initialBalance, initialINRBalance) {
  const addr = targetAddr || getActiveWallet();
  localStorage.setItem('liquidlink_active_wallet', addr);
  saveKnownWallet(addr);

  const walletAddrEl = document.getElementById('walletAddress');
  if (walletAddrEl) walletAddrEl.textContent = addr;

  const heroBalEl = document.getElementById('heroTotalBalance');
  if (heroBalEl && initialBalance !== undefined) {
    heroBalEl.innerHTML = initialBalance + '<small>LQD</small>';
  }

  const inrBalEl = document.getElementById('heroINRBalance');
  const obInrEl = document.getElementById('obAvailableINR');
  if (initialINRBalance !== undefined) {
    const inrStr = (typeof initialINRBalance === 'number') ? initialINRBalance.toFixed(2) : initialINRBalance;
    if (inrBalEl) inrBalEl.textContent = '₹' + inrStr + ' INR';
    if (obInrEl) obInrEl.textContent = '₹' + inrStr + ' INR';
  }

  const topbarBtn = document.getElementById('topbarAvatarBtn');
  if (topbarBtn) topbarBtn.innerHTML = generateAvatarSVG(addr, 36);

  const profileAvatar = document.getElementById('profileAvatarLg');
  if (profileAvatar) profileAvatar.innerHTML = generateAvatarSVG(addr, 72);

  const heroAvatar = document.getElementById('heroAddressAvatar');
  if (heroAvatar) heroAvatar.innerHTML = generateAvatarSVG(addr, 22);

  try {
    const holdings = await fetchData('/api/v1/wallet/holdings?address=' + encodeURIComponent(addr));
    if (holdings) {
      if (holdings.balance !== undefined && heroBalEl) {
        heroBalEl.innerHTML = holdings.balance + '<small>LQD</small>';
      }
      if (holdings.inr_balance !== undefined) {
        const inrStr = (typeof holdings.inr_balance === 'number') ? holdings.inr_balance.toFixed(2) : holdings.inr_balance;
        const resVal = (typeof holdings.inr_reserved === 'number') ? holdings.inr_reserved : parseFloat(holdings.inr_reserved || 0);
        const displayINR = '₹' + inrStr + ' INR' + (resVal > 0 ? ' (' + resVal.toFixed(2) + ' Reserved)' : '');
        if (inrBalEl) inrBalEl.textContent = displayINR;
        if (obInrEl) obInrEl.textContent = displayINR;
      }
      const stEl = document.getElementById('userStakedShares');
      if (stEl && holdings.staked_shares !== undefined) {
        stEl.textContent = holdings.staked_shares + ' Shares';
      }
      const bdEl = document.getElementById('userBondsHeld');
      if (bdEl && holdings.bonds_held !== undefined) {
        bdEl.textContent = holdings.bonds_held + ' Bonds';
      }
    }
  } catch (err) {}

  renderWalletsListUI();
}

function renderWalletsListUI() {
  const container = document.getElementById('profileWalletsList');
  if (!container) return;
  const known = getKnownWallets();
  const active = getActiveWallet();

  let html = '';
  known.forEach((w, idx) => {
    const isActive = (w === active);
    const label = w === '{{ .Address }}' ? 'Node Default' : 'User Wallet ' + (idx);
    html += '<div style="display:flex;align-items:center;justify-content:space-between;padding:8px 12px;background:' + (isActive ? 'rgba(79,110,247,0.12)' : 'rgba(255,255,255,0.03)') + ';border:1px solid ' + (isActive ? 'var(--primary)' : 'var(--border)') + ';border-radius:8px;">' +
      '<div style="display:flex;align-items:center;gap:10px;">' +
      generateAvatarSVG(w, 24) +
      '<div style="display:flex;flex-direction:column;"><span style="font-size:0.82rem;font-weight:600;color:var(--text);">' + label + '</span><span style="font-family:var(--font-mono);font-size:0.75rem;color:var(--text-dim);">' + w.substring(0,10) + '...' + w.substring(w.length-6) + '</span></div>' +
      '</div>' +
      (isActive ? '<span class="badge badge-settled">Active</span>' : '<button class="btn btn-sm btn-ghost" onclick="switchActiveWallet(\'' + w + '\')">Switch</button>') +
      '</div>';
  });
  container.innerHTML = html;
}

function switchActiveWallet(addr) {
  loadActiveWalletUI(addr);
  refreshDashboardData();
  showAlert('Switched active wallet to ' + addr.substring(0, 14) + '...');
}

async function cancelOrderbookOrder(orderId) {
  try {
    const user = getActiveWallet();
    await postData('/api/v1/orderbook/cancel', { order_id: orderId, user });
    showAlert('Order cancelled successfully!');
    refreshDashboardData();
  } catch (err) { showAlert(err.message, true); }
}

/* ---------- Live Dynamic Data Refetching ---------- */
async function refreshOrderbookUI() {
  try {
    const depth = await fetchData('/api/v1/orderbook/depth');
    if (depth) {
      const bidsEl = document.getElementById('obBidsContainer');
      if (bidsEl) {
        if (depth.bids && depth.bids.length > 0) {
          let html = '';
          depth.bids.forEach(b => {
            const priceStr = (typeof b.price === 'number') ? b.price.toFixed(2) : b.price;
            html += '<div style="display:flex; justify-content:space-between; margin-bottom:4px;">' +
              '<span style="color:var(--success);">₹' + priceStr + '</span>' +
              '<span>' + b.quantity + ' LQD</span></div>';
          });
          bidsEl.innerHTML = html;
        } else {
          bidsEl.innerHTML = '<div style="color:var(--text-faint);">No open bids</div>';
        }
      }
      const asksEl = document.getElementById('obAsksContainer');
      if (asksEl) {
        if (depth.asks && depth.asks.length > 0) {
          let html = '';
          depth.asks.forEach(a => {
            const priceStr = (typeof a.price === 'number') ? a.price.toFixed(2) : a.price;
            html += '<div style="display:flex; justify-content:space-between; margin-bottom:4px;">' +
              '<span style="color:var(--red);">₹' + priceStr + '</span>' +
              '<span>' + a.quantity + ' LQD</span></div>';
          });
          asksEl.innerHTML = html;
        } else {
          asksEl.innerHTML = '<div style="color:var(--text-faint);">No open asks</div>';
        }
      }
    }
  } catch (err) {}

  try {
    const res = await fetchData('/api/v1/orderbook/trades');
    const trades = res.trades || res;
    const tradesEl = document.getElementById('obTradesContainer');
    if (tradesEl && Array.isArray(trades)) {
      if (trades.length > 0) {
        let html = '<h3 style="font-size:0.9rem; color:var(--text-dim); margin-bottom:10px;">Recent Trades</h3>';
        trades.forEach(t => {
          const priceStr = (typeof t.price === 'number') ? t.price.toFixed(2) : t.price;
          html += '<div class="list-row"><div class="row-main"><div class="row-title">Trade #' + t.id + '</div>' +
            '<div class="row-sub">' + (t.buyer || '') + ' ← ' + (t.seller || '') + '</div></div>' +
            '<div class="row-side"><div class="top" style="color:var(--accent-strong);">₹' + priceStr + ' INR</div>' +
            '<div class="bottom">' + t.amount + ' LQD</div></div></div>';
        });
        tradesEl.innerHTML = html;
      } else {
        tradesEl.innerHTML = '<h3 style="font-size:0.9rem; color:var(--text-dim); margin-bottom:10px;">Recent Trades</h3><div class="empty-row">No trades executed yet on the orderbook.</div>';
      }
    }
  } catch (err) {}

  try {
    const activeWallet = getActiveWallet();
    const res = await fetchData('/api/v1/orderbook/user-orders?address=' + encodeURIComponent(activeWallet));
    const userOrders = res.orders || res || [];
    const container = document.getElementById('userOrdersContainer');
    if (container && Array.isArray(userOrders)) {
      if (userOrders.length > 0) {
        let html = '';
        userOrders.forEach(o => {
          const isBuy = o.side === 'buy';
          const sideColor = isBuy ? 'var(--success)' : 'var(--red)';
          const priceStr = (typeof o.price === 'number') ? o.price.toFixed(2) : o.price;
          const isCanCancel = (o.status === 'open' || o.status === 'partial');
          let badgeClass = 'badge-pending';
          if (o.status === 'open' || o.status === 'partial') badgeClass = 'badge-active';
          else if (o.status === 'filled') badgeClass = 'badge-settled';

          const actionBtn = isCanCancel 
            ? '<button class="btn btn-sm btn-ghost" onclick="cancelOrderbookOrder(\'' + o.id + '\')">Cancel</button>' 
            : '<span style="font-size:0.75rem; color:var(--text-faint); text-transform:uppercase;">Closed</span>';

          html += '<div class="list-row"><div class="row-main">' +
            '<div class="row-title"><span style="color:' + sideColor + '; font-weight:700; text-transform:uppercase;">' + o.side + '</span> ' + o.id + ' <span class="badge ' + badgeClass + '">' + o.status + '</span></div>' +
            '<div class="row-sub">Price: ₹' + priceStr + ' | Filled: ' + o.filled + '/' + o.amount + ' LQD</div></div>' +
            '<div class="row-side">' + actionBtn + '</div></div>';
        });
        container.innerHTML = html;
      } else {
        container.innerHTML = '<div class="empty-row">No limit orders for active wallet.</div>';
      }
    }
  } catch (err) {}
}

async function refreshChannelsUI() {
  try {
    const res = await fetchData('/api/v1/channels/list');
    const channels = res.channels || [];
    const countEl = document.getElementById('channelsCount');
    if (countEl) countEl.textContent = '(' + channels.length + ')';

    const container = document.getElementById('channelsListContainer');
    if (container) {
      if (channels.length > 0) {
        let html = '';
        channels.forEach(ch => {
          const isOpened = ch.state === 'OPEN' || ch.state === 0;
          const stateStr = isOpened ? 'OPEN' : ch.state;
          const badgeClass = isOpened ? 'badge-active' : 'badge-pending';
          const partyA = (ch.share && ch.share.party_a) || (ch.share && ch.share.PartyA) || '';
          const partyB = (ch.share && ch.share.party_b) || (ch.share && ch.share.PartyB) || '';
          const balA = (ch.share && ch.share.balance_party_a) !== undefined ? ch.share.balance_party_a : (ch.share && ch.share.BalancePartyA);
          const balB = (ch.share && ch.share.balance_party_b) !== undefined ? ch.share.balance_party_b : (ch.share && ch.share.BalancePartyB);

          html += '<div class="list-row"><div class="row-main">' +
            '<div class="row-title">' + ch.id + ' <span class="badge ' + badgeClass + '">' + stateStr + '</span></div>' +
            '<div class="row-sub">' + partyA + ' → ' + partyB + '</div></div>' +
            '<div class="row-side"><div class="top">' + balA + ' / ' + balB + '</div>' +
            '<div class="bottom">seq ' + ch.sequence + '</div></div></div>';
        });
        container.innerHTML = html;
      } else {
        container.innerHTML = '<div class="empty-row">No active state channels yet. Tap "+ Open" to create one.</div>';
      }
    }
  } catch (err) {}
}

async function refreshBridgeUI() {
  try {
    const res = await fetchData('/api/v1/bridge/list');
    const mappings = res.mappings || [];
    const countEl = document.getElementById('bridgeCount');
    if (countEl) countEl.textContent = '(' + mappings.length + ')';

    const container = document.getElementById('bridgeListContainer');
    if (container) {
      if (mappings.length > 0) {
        let html = '';
        mappings.forEach(m => {
          const isSettled = m.status === 'SETTLED' || m.status === 2;
          const statusStr = isSettled ? 'SETTLED' : m.status;
          const badgeClass = isSettled ? 'badge-settled' : 'badge-pending';
          const txId = m.transaction_id || m.TransactionID || '';

          html += '<div class="list-row"><div class="row-main">' +
            '<div class="row-title">' + m.utr + ' <span class="badge ' + badgeClass + '">' + statusStr + '</span></div>' +
            '<div class="row-sub">Tx ' + txId + '</div></div>' +
            '<div class="row-side"><div class="bottom">' + (m.created_at || '') + '</div></div></div>';
        });
        container.innerHTML = html;
      } else {
        container.innerHTML = '<div class="empty-row">No fiat bridge requests yet.</div>';
      }
    }
  } catch (err) {}
}

async function refreshStabilityUI() {
  try {
    const status = await fetchData('/api/v1/stability/status');
    if (status) {
      const priceStr = (typeof status.price === 'number') ? status.price.toFixed(2) : status.price;
      const apyStr = (typeof status.current_apy === 'number') ? status.current_apy.toFixed(1) : status.current_apy;

      const pEl = document.getElementById('marketPriceVal');
      if (pEl) pEl.textContent = '₹' + priceStr + ' INR';

      const obClrEl = document.getElementById('obClearingPrice');
      if (obClrEl) obClrEl.textContent = 'Clearing Price: ₹' + priceStr + ' INR';

      const apyEl = document.getElementById('stakingApyVal');
      if (apyEl) apyEl.textContent = apyStr + '%';

      const stBadge = document.getElementById('stabilityStateBadge');
      if (stBadge) {
        stBadge.textContent = status.state;
        stBadge.className = 'badge ' + (status.state === 'expansion' ? 'badge-active' : 'badge-pending');
      }

      const totStEl = document.getElementById('totalStakedSharesVal');
      if (totStEl && status.total_staked_shares !== undefined) {
        totStEl.textContent = status.total_staked_shares + ' Shares';
      }

      const totBdEl = document.getElementById('totalOutstandingBondsVal');
      if (totBdEl && status.total_outstanding_bonds !== undefined) {
        totBdEl.textContent = status.total_outstanding_bonds + ' Bonds';
      }

      const totConvEl = document.getElementById('totalSupplyConvertedVal');
      if (totConvEl && status.total_staked_shares !== undefined && status.total_outstanding_bonds !== undefined) {
        const conv = (status.total_staked_shares || 0) + (status.total_outstanding_bonds || 0);
        totConvEl.textContent = conv + ' LQD';
      }
    }
  } catch (err) {}

  try {
    const res = await fetchData('/api/v1/stability/history');
    if (res && res.history) {
      renderPriceChart(res.history);
      renderMetricsChart(res.history);
    }
  } catch (err) {}
}

async function refreshFiatPoolUI() {
  try {
    const res = await fetchData('/api/v1/fiat/pool');
    if (res && res.total_inr_pool !== undefined) {
      const poolStr = (typeof res.total_inr_pool === 'number') ? res.total_inr_pool.toFixed(2) : res.total_inr_pool;
      const el = document.getElementById('statInrPool');
      if (el) el.textContent = '₹' + poolStr + ' INR';
      const obPoolEl = document.getElementById('obInrPoolVal');
      if (obPoolEl) obPoolEl.textContent = '₹' + poolStr + ' INR';
    }
  } catch (err) {}
}

async function refreshBlockchainUI() {
  try {
    const res = await fetchData('/api/v1/blockchain/height');
    if (res && res.height !== undefined) {
      const hEl = document.getElementById('statBlockHeight');
      if (hEl) hEl.textContent = res.height;
    }
  } catch (err) {}
}

async function refreshDashboardData() {
  await Promise.allSettled([
    loadActiveWalletUI(),
    refreshOrderbookUI(),
    refreshChannelsUI(),
    refreshBridgeUI(),
    refreshStabilityUI(),
    refreshBlockchainUI(),
    refreshFiatPoolUI()
  ]);
}

function initAvatars() {
  checkFirstTimeWelcome();
  refreshDashboardData();
  setInterval(refreshDashboardData, 3000);
}

document.addEventListener('DOMContentLoaded', initAvatars);
</script>
</body>
</html>`

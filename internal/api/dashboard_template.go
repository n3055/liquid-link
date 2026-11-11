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
<link rel="manifest" href="/manifest.json">
<meta name="theme-color" content="#090d16">
<meta name="mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-status-bar-style" content="black-translucent">
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=IBM+Plex+Mono:wght@400;500;600&display=swap" rel="stylesheet">
<script src="https://cdnjs.cloudflare.com/ajax/libs/qrcodejs/1.0.0/qrcode.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/html5-qrcode/2.3.8/html5-qrcode.min.js"></script>
<script src="https://cdn.jsdelivr.net/npm/tweetnacl@1.0.3/nacl-fast.min.js"></script>
<script src="https://cdn.jsdelivr.net/npm/tweetnacl-util@0.15.1/nacl-util.min.js"></script>
<style>
:root {
  --bg: #090d16;
  --bg-elevated: #0f172a;
  --surface: #131b2e;
  --surface-2: #1e293b;
  --border: #243048;
  --border-soft: #172136;
  --text: #f8fafc;
  --text-dim: #94a3b8;
  --text-faint: #64748b;
  --accent: #38bdf8;
  --accent-dim: #0284c7;
  --accent-strong: #7dd3fc;
  --accent-ink: #041019;
  --accent-soft: rgba(56, 189, 248, 0.12);
  --success: #10b981;
  --success-soft: rgba(16, 185, 129, 0.12);
  --amber: #f59e0b;
  --amber-soft: rgba(245, 158, 11, 0.14);
  --red: #ef4444;
  --red-soft: rgba(239, 68, 68, 0.14);
  --radius-lg: 16px;
  --radius: 12px;
  --radius-sm: 8px;
  --font-display: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  --font-body: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  --font-mono: 'IBM Plex Mono', ui-monospace, SFMono-Regular, 'SF Mono', Consolas, monospace;
  --nav-h: 68px;
}
* { box-sizing: border-box; margin: 0; padding: 0; }
html {
  height: 100%;
  overflow-x: hidden;
  max-width: 100vw;
  width: 100%;
}
body {
  height: 100%;
  overflow-x: hidden;
  max-width: 100vw;
  width: 100%;
  font-family: var(--font-body);
  background: var(--bg);
  color: var(--text);
  line-height: 1.5;
  padding-bottom: calc(var(--nav-h) + 80px + env(safe-area-inset-bottom));
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
.container {
  width: 100%;
  max-width: 580px;
  margin: 0 auto;
  padding: 16px 16px calc(var(--nav-h) + 60px + env(safe-area-inset-bottom));
  box-sizing: border-box;
  overflow-x: hidden;
}

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
.segment {
  display: flex;
  gap: 6px;
  background: var(--surface);
  border: 1px solid var(--border-soft);
  border-radius: 12px;
  padding: 4px;
  margin-bottom: 16px;
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
  scrollbar-width: none;
  max-width: 100%;
}
.segment::-webkit-scrollbar { display: none; }
.segment button {
  flex: 0 0 auto;
  white-space: nowrap;
  border: none;
  background: none;
  padding: 8px 14px;
  border-radius: 8px;
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--text-dim);
  cursor: pointer;
  transition: background 0.15s ease, color 0.15s ease;
}
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
.sheet {
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  width: 100%;
  max-width: 500px;
  border-radius: 22px 22px 0 0;
  padding: 20px 20px calc(48px + env(safe-area-inset-bottom));
  max-height: 86vh;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
  box-sizing: border-box;
  animation: rise 0.22s ease;
}
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
#scannerRegion, #offlineScannerRegion { width: 100% !important; height: 100% !important; min-height: 260px; display: flex; align-items: center; justify-content: center; }
#scannerRegion video, #offlineScannerRegion video { width: 100% !important; height: 100% !important; object-fit: cover !important; border-radius: 16px; display: block !important; }
#scannerRegion canvas, #offlineScannerRegion canvas { display: none !important; }
#scannerRegion img, #offlineScannerRegion img { width: 100% !important; height: 100% !important; object-fit: contain !important; }
.scan-frame::after { content: ""; position: absolute; inset: 10%; border: 2px solid rgba(56,189,248,0.85); border-radius: 16px; box-shadow: 0 0 0 999px rgba(0,0,0,0.45); pointer-events: none; }

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
      <div class="feat-item"><span class="feat-icon"><svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/></svg></span><span>Instant P2P State Channel Payments</span></div>
      <div class="feat-item"><span class="feat-icon"><svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg></span><span>Algorithmic Seigniorage INR Peg Stability</span></div>
      <div class="feat-item"><span class="feat-icon"><svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg></span><span>Real-Time Limit Orderbook Exchange</span></div>
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
    <div class="celebration-icon" style="display:flex; justify-content:center; align-items:center; margin-bottom:10px;">
      <svg width="44" height="44" viewBox="0 0 24 24" fill="none" stroke="var(--success)" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M9 12l2 2 4-4"/></svg>
    </div>
    <h2 style="margin:0 0 4px; font-size:1.35rem; font-weight:700;">Wallet Created</h2>
    <p style="font-size:0.85rem; color:var(--text-dim); margin-bottom:14px;">Your Web3 wallet session is active and funded on LiquidLink.</p>
    
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

    <button class="btn btn-primary" onclick="dismissCelebrationModal()" style="width:100%; margin-top:16px;">Enter Dashboard</button>
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

      <div class="quick-actions" style="display:grid; grid-template-columns: repeat(5, 1fr); gap:8px;">
        <button class="qa-btn" onclick="createNewWalletUI()">
          <span class="qa-circle"><svg width="18" height="18" viewBox="0 0 24 24" fill="none"><path d="M12 5v14M5 12h14" stroke="currentColor" stroke-width="2.2" stroke-linecap="round"/></svg></span>
          <span>New</span>
        </button>
        <button class="qa-btn" onclick="openSheet('sheetSend')">
          <span class="qa-circle"><svg width="18" height="18" viewBox="0 0 24 24" fill="none"><path d="M4 12h16M14 6l6 6-6 6" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"/></svg></span>
          <span>Send</span>
        </button>
        <button class="qa-btn" onclick="openReceive()">
          <span class="qa-circle"><svg width="18" height="18" viewBox="0 0 24 24" fill="none"><rect x="3" y="3" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="2"/><rect x="14" y="3" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="2"/><rect x="3" y="14" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="2"/><path d="M14 14h3v3h-3zM19 14h2v2h-2zM14 19h2v2h-2zM19 19h2v2h-2z" fill="currentColor"/></svg></span>
          <span>Receive</span>
        </button>
        <button class="qa-btn" onclick="openSheet('sheetBluetooth')">
          <span class="qa-circle"><svg width="18" height="18" viewBox="0 0 24 24" fill="none"><path d="M5 12.55a11 11 0 0 1 14.08 0M1.42 9a16 16 0 0 1 21.16 0M8.53 16.11a6 6 0 0 1 6.95 0M12 20h.01" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"/></svg></span>
          <span>Offline</span>
        </button>
        <button class="qa-btn" onclick="openScanner('sendReceiver', 'sheetSend')">
          <span class="qa-circle"><svg width="18" height="18" viewBox="0 0 24 24" fill="none"><path d="M4 7V5a1 1 0 0 1 1-1h2M4 17v2a1 1 0 0 0 1 1h2M20 7V5a1 1 0 0 0-1-1h-2M20 17v2a1 1 0 0 1-1 1h-2" stroke="currentColor" stroke-width="2.2" stroke-linecap="round"/><path d="M4 12h16" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-dasharray="1 3.4"/></svg></span>
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
        <div style="grid-column: 1 / -1; background:var(--surface-2); padding:12px; border-radius:10px;">
          <div style="display:flex; justify-content:space-between; font-size:0.75rem; color:var(--text-dim); margin-bottom:4px;">
            <span>Seigniorage Peg Equilibrium Progress (Target: ₹1.00 INR)</span>
            <span id="pegProgressPercent">100%</span>
          </div>
          <div style="width:100%; height:8px; background:var(--border); border-radius:4px; overflow:hidden;">
            <div id="pegProgressBar" style="width:100%; height:100%; background:linear-gradient(90deg, var(--accent) 0%, var(--success) 100%); transition:width 0.3s ease;"></div>
          </div>
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
              <input type="number" id="stakeAmount" min="1" value="100" class="input" placeholder="LQD Amount" oninput="calcStakeYield()" required>
            </div>
            <div id="stakeCalcPreview" style="font-size:0.75rem; color:var(--success); margin-top:6px; font-weight:600;">Estimated Annual Yield: +14.50 LQD / year</div>
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
              <input type="number" id="bondAmount" min="1" value="50" class="input" placeholder="LQD/Bond Amount" oninput="calcBondProfit()" required>
            </div>
            <div id="bondCalcPreview" style="font-size:0.75rem; color:var(--amber); margin-top:6px; font-weight:600;">Expected Redemption Profit: +0.00 LQD at ₹1.00 peg</div>
            <div style="display:flex; gap:10px; margin-top:12px;">
              <button type="submit" class="btn btn-primary" style="flex:1;">Buy Discounted Bonds</button>
              <button type="button" onclick="submitRedeemBonds()" class="btn btn-secondary" style="flex:1;">Redeem Bonds</button>
            </div>
          </form>
        </div>
      </div>

      <!-- Economist Design Explanation -->
      <div style="margin:16px; background:linear-gradient(135deg, rgba(77,132,226,0.1), rgba(129,140,248,0.1)); border:1px solid var(--accent-soft); border-radius:12px; padding:16px;">
        <h4 style="font-size:0.9rem; color:var(--accent-strong); margin-bottom:6px;">Seigniorage Stability Protocol</h4>
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
          <button class="btn btn-sm btn-secondary" onclick="openReceiveOfflineModal()">Receive Offline</button>
          <button class="btn btn-sm btn-secondary" onclick="openSheet('sheetBluetooth')">Offline P2P Sync</button>
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
      <div class="list-row" onclick="openVerifySheet('{{ .UTR }}')" style="cursor:pointer;">
        <div class="row-main">
          <div class="row-title">{{ .UTR }} <span class="badge {{ if eq (printf "%v" .Status) "SETTLED" }}badge-settled{{ else }}badge-pending{{ end }}">{{ .Status }}</span></div>
          <div class="row-sub">Tx {{ .TransactionID }}</div>
        </div>
        <div class="row-side" style="display:flex; flex-direction:column; align-items:flex-end; gap:4px;">
          {{ if ne (printf "%v" .Status) "SETTLED" }}
          <button class="btn btn-sm btn-secondary" onclick="event.stopPropagation(); openVerifySheet('{{ .UTR }}');">Verify & Settle</button>
          {{ end }}
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
        <h2>Activity & Event Feed <span class="count" id="blocksCount">({{ len .ActivityLog }})</span></h2>
      </div>
      <div id="blocksListContainer">
      {{ range .ActivityLog }}
      <div class="list-row">
        <div class="row-main">
          <div class="row-title">{{ .Title }}</div>
          <div class="row-sub">{{ .Subtitle }}</div>
        </div>
        <div class="row-side">
          <div class="top">{{ .Detail }}</div>
          <div class="bottom">{{ .Timestamp }}</div>
        </div>
      </div>
      {{ else }}
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
      <div class="empty-row">No activity events recorded yet.</div>
      {{ end }}
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
  <div class="sheet-head"><h3>Scan QR Code</h3><button type="button" class="sheet-close" onclick="closeSheetAndStopScan(event)"><svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M6 6l12 12M18 6L6 18" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg></button></div>
  <div class="scan-frame"><div id="scannerRegion"></div></div>
  <div class="hint" id="scanHint" style="margin-top:10px;">Point camera at a LiquidLink QR code or snap with device camera.</div>
  <input type="file" id="qrCameraDirect" accept="image/*" capture="environment" style="display:none;" onchange="handleQrImageUpload(this)">
  <input type="file" id="qrFileInput" accept="image/*" style="display:none;" onchange="handleQrImageUpload(this)">
  <div class="btn-row" style="margin-top:14px; gap:8px;">
    <button type="button" class="btn btn-primary" onclick="document.getElementById('qrCameraDirect').click()" style="flex:1;">Snap Camera</button>
    <button type="button" class="btn btn-secondary" onclick="document.getElementById('qrFileInput').click()" style="flex:1;">Upload Image</button>
  </div>
</div></div>

<!-- Mint -->
<div class="overlay" id="sheetMint"><div class="sheet">
  <div class="sheet-handle"></div>
  <div class="sheet-head"><h3>Devnet Faucet / Mint</h3><button class="sheet-close" onclick="closeSheet('sheetMint')"><svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M6 6l12 12M18 6L6 18" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg></button></div>
  <div style="font-size:0.8rem; color:#f59e0b; margin-bottom:12px; padding:8px 12px; background:rgba(245,158,11,0.1); border:1px solid rgba(245,158,11,0.3); border-radius:6px; display:flex; align-items:center; gap:8px;">
    <span><strong>Genesis Faucet:</strong> Minting is an administrative tool for testnet genesis allocation. Real end-users acquire tokens via transfers, trades, or fiat bridging.</span>
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
      <div style="font-size:0.75rem; color:var(--text-dim); margin-bottom:12px;">Opening from active wallet: <strong id="openChannelSenderAddr" style="font-family:var(--font-mono); color:var(--accent-strong);"></strong></div>
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
        <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:6px;">
          <label for="chanIdUpdate" style="margin:0;">Select Channel to Update</label>
          <button type="button" onclick="toggleCustomUpdateChan()" id="btnToggleCustomUpdate" style="font-size:0.75rem; color:var(--accent); background:none; border:none; cursor:pointer;">+ Custom ID</button>
        </div>
        <select id="chanIdUpdate" style="width:100%; padding:10px 12px; background:var(--bg-elevated); color:var(--text); border:1px solid var(--border); border-radius:8px;">
          <option value="">-- Choose active channel --</option>
        </select>
        <input type="text" id="chanIdUpdateCustom" placeholder="Channel hash ID" style="display:none; margin-top:8px;">
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
      <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:6px;">
        <label for="chanIdSettle" style="margin:0;">Select Channel to Settle</label>
        <button type="button" onclick="toggleCustomSettleChan()" id="btnToggleCustomSettle" style="font-size:0.75rem; color:var(--accent); background:none; border:none; cursor:pointer;">+ Custom ID</button>
      </div>
      <select id="chanIdSettle" style="width:100%; padding:10px 12px; background:var(--bg-elevated); color:var(--text); border:1px solid var(--border); border-radius:8px;">
        <option value="">-- Choose active channel --</option>
      </select>
      <input type="text" id="chanIdSettleCustom" placeholder="Channel hash ID" style="display:none; margin-top:8px;">
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

  <div style="margin-top:16px; font-size:0.85rem; font-weight:600; color:var(--text); text-transform:uppercase; letter-spacing:0.5px;">Bluetooth Offline Mesh</div>
  <div style="margin-top:8px; display:flex; gap:8px;">
    <button class="btn btn-sm btn-secondary" onclick="openSheet('sheetBluetooth')" style="flex:1;">Offline P2P Sync</button>
    <button class="btn btn-sm btn-ghost" onclick="sendOfflineBluetoothTx()" style="flex:1;">Relay BLE Tx</button>
  </div>

  <div style="margin-top:16px;" class="btn-row">
    <button class="btn btn-secondary" onclick="closeSheet('sheetProfile'); openReceive();">Show my QR</button>
    <button class="btn btn-secondary" onclick="closeSheet('sheetProfile'); openScanner('sendReceiver','sheetSend');">Scan code</button>
  </div>
</div></div>

<!-- Bluetooth P2P & Offline Sync Modal -->
<div class="overlay" id="sheetBluetooth"><div class="sheet">
  <div class="sheet-handle"></div>
  <div class="sheet-head">
    <h3>Offline P2P & BLE Mesh</h3>
    <button class="sheet-close" onclick="closeSheet('sheetBluetooth')">
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M6 6l12 12M18 6L6 18" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
    </button>
  </div>

  <div class="sheet-tabs" style="display:grid; grid-template-columns: repeat(4, 1fr); gap:4px;">
    <button class="active" onclick="switchSubpanel('subBlePayment', this, 'sheetBluetooth')" style="font-size:0.74rem; padding:8px 4px;">Send Offline</button>
    <button onclick="switchSubpanel('subBleReceive', this, 'sheetBluetooth')" style="font-size:0.74rem; padding:8px 4px;">Receive QR</button>
    <button onclick="switchSubpanel('subBlePair', this, 'sheetBluetooth')" style="font-size:0.74rem; padding:8px 4px;">BLE Device</button>
    <button onclick="switchSubpanel('subBleQueue', this, 'sheetBluetooth')" style="font-size:0.74rem; padding:8px 4px;">Outbox (<span id="bleQueueCount">0</span>)</button>
  </div>

  <!-- Subpanel 2: Make Offline Payment -->
  <div id="subBlePayment" class="sheet-subpanel active">
    <form onsubmit="submitOfflinePaymentForm(event)">
      <div class="form-group">
        <label>Payment Medium</label>
        <select id="blePayType" onchange="toggleBlePaymentInputs()">
          <option value="pcn">State Channel Payment (PCN Off-Chain)</option>
          <option value="direct">Direct Signed Transfer (Ed25519)</option>
        </select>
      </div>

      <div class="form-group" id="grpBleChannel">
        <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:6px;">
          <label for="bleChannelID" style="margin:0;">Select State Channel</label>
          <button type="button" onclick="toggleCustomChannelInput()" id="btnToggleCustomChan" style="font-size:0.75rem; color:var(--accent); background:none; border:none; cursor:pointer;">+ Custom ID</button>
        </div>
        <select id="bleChannelID" style="width:100%; padding:10px 12px; background:var(--bg-elevated); color:var(--text); border:1px solid var(--border); border-radius:8px;">
          <option value="">-- Choose active channel --</option>
        </select>
        <input type="text" id="bleChannelCustomID" placeholder="Paste custom channel hash" style="display:none; margin-top:8px;">
      </div>

      <div class="form-group" id="grpBleReceiver" style="display:none;">
        <label for="bleReceiver">Receiver Wallet Address</label>
        <input type="text" id="bleReceiver" placeholder="0x...">
      </div>

      <div class="form-group">
        <label for="bleAmount">Amount (LQD)</label>
        <input type="number" id="bleAmount" min="1" placeholder="10" required>
      </div>

      <button type="submit" class="btn btn-primary" style="width:100%;">Sign & Show Payment QR</button>
    </form>

    <!-- Signed Offline QR Display Card -->
    <div id="offlinePaymentQrCard" style="display:none; background:var(--surface-2); padding:16px; border-radius:12px; border:1px solid var(--border); margin-top:14px; text-align:center;">
      <div style="font-size:0.85rem; font-weight:600; color:var(--accent); margin-bottom:4px;">Signed Offline Payment</div>
      <div style="font-size:0.75rem; color:var(--text-dim); margin-bottom:10px;" id="offlineQrDetails">Channel Payment (Seq 3)</div>
      <div id="offlinePaymentQrCanvas" style="display:inline-block; padding:10px; background:white; border-radius:8px; margin-bottom:10px;"></div>
      <div style="font-size:0.74rem; color:var(--text-faint); margin-bottom:12px;">Present this QR code to the receiver's phone to transfer value offline.</div>
      <button type="button" class="btn btn-sm btn-secondary" onclick="dismissOfflinePaymentQr()" style="width:100%;">Done</button>
    </div>
  </div>

  <!-- Subpanel 4: Receive Offline Payment via Camera -->
  <div id="subBleReceive" class="sheet-subpanel">
    <div style="background:var(--surface-2); padding:16px; border-radius:12px; border:1px solid var(--border); text-align:center;">
      <div style="font-size:0.95rem; font-weight:600; color:var(--text); margin-bottom:4px;">Scan Offline Payment</div>
      <p class="hint" id="offlineScanHint" style="font-size:0.78rem; color:var(--text-dim); margin-bottom:12px;">Scan the payer's offline payment QR code to verify Ed25519 signature and credit channel balance with zero internet.</p>
      <div class="scan-frame" style="width:100%; max-width:280px; margin:0 auto 14px;"><div id="offlineScannerRegion"></div></div>
      <input type="file" id="offlineCameraDirect" accept="image/*" capture="environment" style="display:none;" onchange="handleOfflineQrImageUpload(this)">
      <div style="display:flex; gap:8px;">
        <button type="button" class="btn btn-primary" onclick="startOfflinePaymentScan()" id="btnStartOfflineScan" style="flex:1;">Live Camera</button>
        <button type="button" class="btn btn-secondary" onclick="stopOfflinePaymentScan()" id="btnStopOfflineScan" style="flex:1; display:none;">Stop Camera</button>
        <button type="button" class="btn btn-primary" onclick="document.getElementById('offlineCameraDirect').click()" id="btnDirectOfflineSnap" style="flex:1;">Snap Camera</button>
      </div>
      <div style="margin-top:10px;">
        <button type="button" class="btn btn-sm btn-ghost" onclick="document.getElementById('offlineQrFileInput').click()" style="width:100%;">Upload QR Image File</button>
        <input type="file" id="offlineQrFileInput" accept="image/*" style="display:none;" onchange="handleOfflineQrImageUpload(this)">
      </div>
    </div>
  </div>

  <!-- Subpanel 1: BLE Device Pairing -->
  <div id="subBlePair" class="sheet-subpanel">
    <div style="background:var(--surface-2); padding:14px; border-radius:12px; border:1px solid var(--border); margin-bottom:14px;">
      <div style="font-size:0.75rem; color:var(--text-dim); margin-bottom:4px;">BLE Pairing Status</div>
      <div id="bleConnStatus" style="font-weight:700; font-size:0.95rem; color:var(--text); display:flex; align-items:center; gap:8px;">
        <span style="width:8px; height:8px; border-radius:50%; background:#989ba3; display:inline-block;" id="bleDot"></span>
        <span id="bleStatusText">Disconnected</span>
      </div>
      <div id="bleDeviceInfo" style="font-size:0.78rem; font-family:var(--font-mono); color:var(--accent-strong); margin-top:6px;">No device paired</div>
    </div>
    <div style="display:flex; gap:10px;">
      <button class="btn btn-primary" onclick="pairWebBluetooth()" style="flex:1;">Pair BLE Device</button>
      <button class="btn btn-secondary" onclick="disconnectWebBluetooth()" style="flex:1;">Disconnect</button>
    </div>
    <p class="hint" style="margin-top:12px; font-size:0.78rem; color:var(--text-faint);">
      Connects directly to nearby devices running LiquidLink BLE GATT service. Payments made while disconnected will buffer locally with Ed25519 signatures.
    </p>
  </div>

  <!-- Subpanel 3: Queued Offline Payments & Host Push -->
  <div id="subBleQueue" class="sheet-subpanel">
    <div style="margin-bottom:12px;">
      <label for="bleHostUrl">Host Node Sync URL</label>
      <div style="display:flex; gap:8px;">
        <input type="text" id="bleHostUrl" placeholder="http://192.168.1.X:8080" style="font-family:var(--font-mono); font-size:0.85rem;">
        <button class="btn btn-sm btn-secondary" onclick="resetHostUrl()" style="white-space:nowrap;">This Node</button>
      </div>
      <p class="hint" style="margin-top:4px;">Endpoint of the validator node to commit offline state transitions.</p>
    </div>

    <div id="bleQueueList" style="display:flex; flex-direction:column; gap:8px; max-height:220px; overflow-y:auto; margin-bottom:14px;">
      <div class="empty-row" style="padding:16px;">No queued offline payments.</div>
    </div>

    <div style="display:flex; gap:10px;">
      <button class="btn btn-primary" onclick="pushOfflineQueueToHost()" id="btnPushHost" style="flex:2;">Push to Validator</button>
      <button class="btn btn-secondary" onclick="clearOfflineQueueUI()" style="flex:1;">Clear</button>
    </div>
  </div>
</div></div>

<script>
/* ---------- PWA Service Worker Registration ---------- */
if ('serviceWorker' in navigator) {
  window.addEventListener('load', () => {
    navigator.serviceWorker.register('/sw.js').then((reg) => {
      console.log('[LiquidLink PWA] Service Worker registered:', reg.scope);
    }).catch((err) => {
      console.warn('[LiquidLink PWA] Service Worker registration failed:', err);
    });
  });
}

/* ---------- In-Browser Cryptographic Utilities (TweetNaCl & WebCrypto) ---------- */
function bytesToHex(bytes) {
  return Array.from(bytes).map(b => b.toString(16).padStart(2, '0')).join('');
}

function hexToBytes(hex) {
  const bytes = new Uint8Array(hex.length / 2);
  for (let i = 0; i < bytes.length; i++) {
    bytes[i] = parseInt(hex.substr(i * 2, 2), 16);
  }
  return bytes;
}

// Synchronous SHA-256 digest computation (RFC 6234 compliant)
function sha256Sync(data) {
  const K = [
    0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
    0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3, 0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
    0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
    0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
    0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13, 0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
    0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
    0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
    0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208, 0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2
  ];

  let H0 = 0x6a09e667, H1 = 0xbb67ae85, H2 = 0x3c6ef372, H3 = 0xa54ff53a;
  let H4 = 0x510e527f, H5 = 0x9b05688c, H6 = 0x1f83d9ab, H7 = 0x5be0cd19;

  const input = typeof data === 'string' ? new TextEncoder().encode(data) : data;
  const l = input.length;
  const bitLen = l * 8;
  const newLen = ((l + 8) >> 6 << 6) + 64;
  const buf = new Uint8Array(newLen);
  buf.set(input);
  buf[l] = 0x80;

  const view = new DataView(buf.buffer);
  view.setUint32(newLen - 4, bitLen, false);

  const W = new Int32Array(64);
  for (let i = 0; i < newLen; i += 64) {
    for (let t = 0; t < 16; t++) {
      W[t] = view.getInt32(i + t * 4, false);
    }
    for (let t = 16; t < 64; t++) {
      const s0 = ((W[t-15] >>> 7) | (W[t-15] << 25)) ^ ((W[t-15] >>> 18) | (W[t-15] << 14)) ^ (W[t-15] >>> 3);
      const s1 = ((W[t-2] >>> 17) | (W[t-2] << 15)) ^ ((W[t-2] >>> 19) | (W[t-2] << 13)) ^ (W[t-2] >>> 10);
      W[t] = (W[t-16] + s0 + W[t-7] + s1) | 0;
    }

    let a = H0, b = H1, c = H2, d = H3, e = H4, f = H5, g = H6, h = H7;
    for (let t = 0; t < 64; t++) {
      const S1 = ((e >>> 6) | (e << 26)) ^ ((e >>> 11) | (e << 21)) ^ ((e >>> 25) | (e << 7));
      const ch = (e & f) ^ ((~e) & g);
      const temp1 = (h + S1 + ch + K[t] + W[t]) | 0;
      const S0 = ((a >>> 2) | (a << 30)) ^ ((a >>> 13) | (a << 19)) ^ ((a >>> 22) | (a << 10));
      const maj = (a & b) ^ (a & c) ^ (b & c);
      const temp2 = (S0 + maj) | 0;

      h = g;
      g = f;
      f = e;
      e = (d + temp1) | 0;
      d = c;
      c = b;
      b = a;
      a = (temp1 + temp2) | 0;
    }

    H0 = (H0 + a) | 0;
    H1 = (H1 + b) | 0;
    H2 = (H2 + c) | 0;
    H3 = (H3 + d) | 0;
    H4 = (H4 + e) | 0;
    H5 = (H5 + f) | 0;
    H6 = (H6 + g) | 0;
    H7 = (H7 + h) | 0;
  }

  const out = new Uint8Array(32);
  const outView = new DataView(out.buffer);
  outView.setUint32(0, H0, false);
  outView.setUint32(4, H1, false);
  outView.setUint32(8, H2, false);
  outView.setUint32(12, H3, false);
  outView.setUint32(16, H4, false);
  outView.setUint32(20, H5, false);
  outView.setUint32(24, H6, false);
  outView.setUint32(28, H7, false);
  return out;
}

async function deriveAddressFromPub(pubBytes) {
  const hashBytes = sha256Sync(pubBytes);
  return '0x' + bytesToHex(hashBytes.slice(0, 20));
}

async function getOrCreateClientKeypair() {
  let privHex = localStorage.getItem('liquidlink_client_privkey');
  let pubHex = localStorage.getItem('liquidlink_client_pubkey');
  if (privHex && pubHex) {
    const pubBytes = hexToBytes(pubHex);
    const addr = await deriveAddressFromPub(pubBytes);
    return {
      secretKey: hexToBytes(privHex),
      publicKey: pubBytes,
      address: addr
    };
  }
  if (window.nacl) {
    const kp = nacl.sign.keyPair();
    privHex = bytesToHex(kp.secretKey);
    pubHex = bytesToHex(kp.publicKey);
    localStorage.setItem('liquidlink_client_privkey', privHex);
    localStorage.setItem('liquidlink_client_pubkey', pubHex);
    const addr = await deriveAddressFromPub(kp.publicKey);
    localStorage.setItem('liquidlink_client_address', addr);
    return { secretKey: kp.secretKey, publicKey: kp.publicKey, address: addr };
  }
  return null;
}

/* ---------- Toast ---------- */
let toastTimer;
function showAlert(msg, isError) {
  const box = document.getElementById('alertBox');
  box.className = (isError ? 'alert-danger' : 'alert-success');
  box.textContent = msg;
  if (typeof requestAnimationFrame === 'function') {
    requestAnimationFrame(() => box.classList.add('show'));
  } else {
    box.classList.add('show');
  }
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => box.classList.remove('show'), 4200);
}

/* ---------- Sheets & Modal Management ---------- */
function openSheet(id) {
  const el = document.getElementById(id);
  if (el) {
    if (id === 'sheetBridge') {
      const rec = document.getElementById('bridgeReceiver');
      if (rec && !rec.value) rec.value = getActiveWallet();
    }
    if (id === 'sheetChannel') {
      const senderEl = document.getElementById('openChannelSenderAddr');
      if (senderEl) senderEl.textContent = getActiveWallet();
    }
    el.classList.add('open');
    document.body.style.overflow = 'hidden';
  }
}

function closeSheet(id) {
  if (id === 'sheetBluetooth') {
    stopOfflinePaymentScan();
  }
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
  if (scope === 'sheetBluetooth') {
    if (id === 'subBleReceive') {
      startOfflinePaymentScan();
    } else {
      stopOfflinePaymentScan();
    }
  }
}

function openVerifySheet(utr) {
  openSheet('sheetBridge');
  const navBtns = document.querySelectorAll('#sheetBridge .sheet-tabs button');
  if (navBtns && navBtns[1]) switchSubpanel('subBridgeVerify', navBtns[1], 'sheetBridge');
  const input = document.getElementById('verifyUTR');
  if (input) input.value = utr;
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

  // Safely stop & clear any previous scanner instance
  if (html5QrScanner) {
    const prev = html5QrScanner;
    html5QrScanner = null;
    try { await prev.stop().catch(() => {}); } catch(e) {}
    try { prev.clear(); } catch(e) {}
  }

  // Explicitly prompt for camera permission immediately on user interaction
  if (navigator.mediaDevices && navigator.mediaDevices.getUserMedia) {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ video: { facingMode: 'environment' } });
      stream.getTracks().forEach(t => t.stop());
    } catch (pErr) {
      console.warn('[Camera] Permission request:', pErr);
    }
  }

  if (!window.Html5Qrcode) {
    if (hint) hint.innerHTML = 'Camera engine loading... Tap <b>"Snap Camera"</b> below!';
    return;
  }

  if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
    if (hint) {
      hint.innerHTML = '<b>Plain HTTP Note:</b> Mobile browsers restrict live webcam streaming over insecure HTTP. Tap <b>"Snap Camera"</b> below to capture with native camera!';
    }
    return;
  }

  if (hint) hint.innerHTML = 'Starting camera stream... Please tap <b>Allow</b> if prompted.';

  html5QrScanner = new Html5Qrcode('scannerRegion', {
    formatsToSupport: window.Html5QrcodeSupportedFormats ? [ window.Html5QrcodeSupportedFormats.QR_CODE ] : undefined,
    experimentalFeatures: { useBarCodeDetectorIfSupported: true }
  });

  const qrConfig = {
    fps: 10,
    qrbox: function(w, h) {
      const minEdge = Math.min(w, h);
      const edge = Math.max(160, Math.floor(minEdge * 0.85));
      return { width: edge, height: edge };
    },
    videoConstraints: {
      facingMode: { ideal: "environment" },
      width: { min: 640, ideal: 1280 },
      height: { min: 480, ideal: 720 }
    }
  };

  try {
    await html5QrScanner.start(
      { facingMode: { ideal: 'environment' } },
      qrConfig,
      onQrScanSuccess,
      () => {}
    );
    if (hint) hint.innerHTML = 'Point camera at a LiquidLink QR code to scan.';
  } catch (err) {
    console.warn('Html5Qrcode start error, trying camera list:', err);
    try {
      const cameras = await Html5Qrcode.getCameras();
      if (cameras && cameras.length > 0) {
        const backCamera = cameras.find(c => c.label.toLowerCase().includes('back') || c.label.toLowerCase().includes('rear')) || cameras[0];
        await html5QrScanner.start(backCamera.id, qrConfig, onQrScanSuccess, () => {});
        if (hint) hint.innerHTML = 'Point camera at a LiquidLink QR code to scan.';
        return;
      }
    } catch (camErr) {}
    
    if (hint) {
      hint.innerHTML = 'Live camera stream paused. Tap <b>"Snap Camera"</b> below to capture directly!';
    }
  }
}

function onQrScanSuccess(decodedText) {
  // Check if scanned QR is an offline payment payload
  if (decodedText) {
    try {
      let raw = typeof decodedText === 'string' ? decodedText.trim() : '';
      if (raw.startsWith('%7B')) raw = decodeURIComponent(raw);
      if ((raw.startsWith('{') && raw.endsWith('}')) || typeof decodedText === 'object') {
        const parsed = typeof decodedText === 'object' ? decodedText : JSON.parse(raw);
        if (parsed && parsed.signature && (parsed.channel_id || parsed.type === 'pcn_offline_payment' || parsed.type === 'direct_tx_offline' || parsed.type === 'direct_transfer' || parsed.receiver)) {
          closeSheetAndStopScan();
          onOfflineQrScanSuccess(parsed);
          return;
        }
      }
    } catch (_) {}
  }
  if (currentTargetField) {
    const field = document.getElementById(currentTargetField);
    if (field) field.value = decodedText;
  }
  closeSheetAndStopScan();
  if (currentNextSheet) openSheet(currentNextSheet);
  showAlert('Address captured: ' + String(decodedText).substring(0, 16) + '...');
}

async function handleQrImageUpload(input) {
  if (!input.files || input.files.length === 0) return;
  const file = input.files[0];
  const hint = document.getElementById('scanHint');
  hint.textContent = 'Decoding QR image...';

  // Stop live scanner if currently running
  if (html5QrScanner) {
    try { await html5QrScanner.stop(); } catch (_) {}
    try { html5QrScanner.clear(); } catch (_) {}
    html5QrScanner = null;
  }

  if (!window.Html5Qrcode) {
    showAlert('QR scanner engine not ready', true);
    return;
  }

  const fileScanner = new Html5Qrcode('scannerRegion', {
    formatsToSupport: window.Html5QrcodeSupportedFormats ? [ window.Html5QrcodeSupportedFormats.QR_CODE ] : undefined,
    experimentalFeatures: { useBarCodeDetectorIfSupported: true }
  });
  fileScanner.scanFile(file, true)
    .then(decodedText => {
      fileScanner.clear();
      onQrScanSuccess(decodedText);
    })
    .catch(err => {
      fileScanner.clear();
      hint.innerHTML = ' Could not detect a valid QR code in that photo.<br>Please select a clearer image or paste address directly.';
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

function getStorageKey(key) {
  const port = window.location.port || '8080';
  return 'liquidlink_' + key + '_' + port;
}

function getWalletHoldingsKey(addr) {
  const target = addr || getActiveWallet();
  return getStorageKey('holdings_' + (target || '').toLowerCase());
}

function getCachedWalletHoldings(addr) {
  try {
    const raw = localStorage.getItem(getWalletHoldingsKey(addr));
    if (raw) return JSON.parse(raw);
  } catch (e) {}
  return null;
}

function saveWalletHoldings(addr, holdings) {
  const target = addr || getActiveWallet();
  if (!target) return;
  try {
    const prev = getCachedWalletHoldings(target) || {};
    const merged = Object.assign({}, prev, holdings);
    localStorage.setItem(getWalletHoldingsKey(target), JSON.stringify(merged));
  } catch (e) {}
}

function getChannelsStorageKey(addr) {
  const target = addr || getActiveWallet();
  return getStorageKey('pcn_channels_' + (target || '').toLowerCase());
}

function getCachedWalletChannels(addr) {
  try {
    const raw = localStorage.getItem(getChannelsStorageKey(addr));
    if (raw) {
      const parsed = JSON.parse(raw);
      if (Array.isArray(parsed)) return parsed;
    }
  } catch (e) {}
  return [];
}

function saveWalletChannels(addr, channels) {
  const target = addr || getActiveWallet();
  try {
    localStorage.setItem(getChannelsStorageKey(target), JSON.stringify(channels || []));
  } catch (e) {}
}

function getBridgeStorageKey(addr) {
  const target = addr || getActiveWallet();
  return getStorageKey('bridge_mappings_' + (target || '').toLowerCase());
}

function getCachedWalletBridgeMappings(addr) {
  try {
    const raw = localStorage.getItem(getBridgeStorageKey(addr));
    if (raw) {
      const parsed = JSON.parse(raw);
      if (Array.isArray(parsed)) return parsed;
    }
  } catch (e) {}
  return [];
}

function saveWalletBridgeMappings(addr, mappings) {
  const target = addr || getActiveWallet();
  try {
    localStorage.setItem(getBridgeStorageKey(target), JSON.stringify(mappings || []));
  } catch (e) {}
}

function getOfflineQueueStorageKey(addr) {
  const target = addr || getActiveWallet();
  return getStorageKey('offline_queue_' + (target || '').toLowerCase());
}

function getCachedWalletOfflineQueue(addr) {
  try {
    const raw = localStorage.getItem(getOfflineQueueStorageKey(addr));
    if (raw) {
      const parsed = JSON.parse(raw);
      if (Array.isArray(parsed)) return parsed;
    }
  } catch (e) {}
  return [];
}

function saveWalletOfflineQueue(addr, queue) {
  const target = addr || getActiveWallet();
  try {
    localStorage.setItem(getOfflineQueueStorageKey(target), JSON.stringify(queue || []));
  } catch (e) {}
  if (target && target.toLowerCase() === (getActiveWallet() || '').toLowerCase()) {
    renderOfflineQueue();
  }
}

let activeBleDevice = null;
let activeBleServer = null;
let activeBleCharacteristic = null;

function getOfflineQueue() {
  return getCachedWalletOfflineQueue(getActiveWallet());
}

function saveOfflineQueue(queue) {
  saveWalletOfflineQueue(getActiveWallet(), queue);
}

function renderOfflineQueue() {
  const queue = getOfflineQueue();
  const countEl = document.getElementById('bleQueueCount');
  if (countEl) countEl.textContent = queue.length;

  const listEl = document.getElementById('bleQueueList');
  if (!listEl) return;

  if (queue.length === 0) {
    listEl.innerHTML = '<div class="empty-row" style="padding:16px;">No queued offline payments.</div>';
    return;
  }

  let html = '';
  queue.forEach((item, index) => {
    const isPcn = item.type === 'pcn_offline_payment' || !!item.channel_id;
    const targetLabel = isPcn ? 'Channel: ' + (item.channel_id || '').slice(0, 12) + '...' : 'To: ' + (item.receiver || '').slice(0, 12) + '...';
    const timeStr = item.timestamp ? new Date(item.timestamp).toLocaleTimeString() : '';
    html += '<div class="list-row" style="padding:10px 12px; background:var(--surface-2); border-radius:8px; margin-bottom:4px;">' +
      '<div class="row-main">' +
        '<div class="row-title" style="font-size:0.8rem;">' + (isPcn ? 'PCN Off-Chain' : 'Direct Transfer') +
          ' <span class="badge badge-pending" style="font-size:0.6rem;">' + (item.status || 'Pending Sync') + '</span></div>' +
        '<div class="row-sub" style="font-size:0.74rem;">' + targetLabel + (item.sequence ? ' (Seq ' + item.sequence + ')' : '') + ' • ' + timeStr + '</div>' +
      '</div>' +
      '<div class="row-side">' +
        '<div class="top" style="font-size:0.85rem; color:var(--accent-strong);">' + item.amount + ' LQD</div>' +
        '<button class="icon-btn-mini" onclick="removeOfflineQueueItem(' + index + ')" title="Remove" style="margin-left:auto; margin-top:2px;">✕</button>' +
      '</div>' +
    '</div>';
  });
  listEl.innerHTML = html;
}

function removeOfflineQueueItem(index) {
  const queue = getOfflineQueue();
  if (index >= 0 && index < queue.length) {
    queue.splice(index, 1);
    saveOfflineQueue(queue);
    showAlert('Removed item from offline queue');
  }
}

function clearOfflineQueueUI() {
  saveOfflineQueue([]);
  showAlert('Offline transaction queue cleared');
}

function resetHostUrl() {
  const input = document.getElementById('bleHostUrl');
  if (input) input.value = window.location.origin;
}

function toggleBlePaymentInputs() {
  const type = document.getElementById('blePayType').value;
  const grpChan = document.getElementById('grpBleChannel');
  const grpRecv = document.getElementById('grpBleReceiver');
  if (type === 'pcn') {
    if (grpChan) grpChan.style.display = 'block';
    if (grpRecv) grpRecv.style.display = 'none';
  } else {
    if (grpChan) grpChan.style.display = 'none';
    if (grpRecv) grpRecv.style.display = 'block';
  }
}

async function pairWebBluetooth() {
  if (!navigator.bluetooth) {
    showAlert('Web Bluetooth API not available in browser. Using node BLE mesh daemon transport driver.', false);
    updateBleUI(true, 'Local BLE Mesh Driver (hci0)', 'Driver UUID: 00001101-0000-1000-8000-00805F9B34FB');
    return;
  }
  try {
    showAlert('Scanning for nearby BLE devices...');
    const device = await navigator.bluetooth.requestDevice({
      acceptAllDevices: true,
      optionalServices: ['00001101-0000-1000-8000-00805f9b34fb']
    });
    activeBleDevice = device;
    updateBleUI(true, device.name || device.id, 'Device ID: ' + device.id);
    showAlert('Paired with BLE Device: ' + (device.name || device.id));

    device.addEventListener('gattserverdisconnected', onBleDisconnected);

    // Try GATT connect
    if (device.gatt) {
      try {
        activeBleServer = await device.gatt.connect();
        showAlert('GATT connected to ' + (device.name || device.id) + '!');
        const service = await activeBleServer.getPrimaryService('00001101-0000-1000-8000-00805f9b34fb').catch(() => null);
        if (service) {
          const chars = await service.getCharacteristics().catch(() => []);
          if (chars && chars.length > 0) activeBleCharacteristic = chars[0];
        }
      } catch (gattErr) {
        // Paired at Web Bluetooth level; fallback to node BLE mesh transport relay
      }
    }
  } catch (e) {
    showAlert('Web Bluetooth pairing: ' + (e.message || 'Cancelled by user'), true);
  }
}

function onBleDisconnected() {
  updateBleUI(false, 'Disconnected', 'No device paired');
  activeBleDevice = null;
  activeBleServer = null;
  activeBleCharacteristic = null;
  showAlert('Bluetooth device disconnected', true);
}

function disconnectWebBluetooth() {
  if (activeBleDevice && activeBleDevice.gatt && activeBleDevice.gatt.connected) {
    activeBleDevice.gatt.disconnect();
  }
  onBleDisconnected();
}

function updateBleUI(connected, statusText, infoText) {
  const dot = document.getElementById('bleDot');
  const txt = document.getElementById('bleStatusText');
  const info = document.getElementById('bleDeviceInfo');
  if (dot) dot.style.background = connected ? '#10b981' : '#989ba3';
  if (txt) txt.textContent = statusText;
  if (info) info.textContent = infoText;
}

async function connectWebBluetooth() {
  await pairWebBluetooth();
}

async function sendOfflineBluetoothTx() {
  openSheet('sheetBluetooth');
  switchSubpanel('subBlePayment', document.querySelectorAll('#sheetBluetooth .sheet-tabs button')[1], 'sheetBluetooth');
}

async function submitOfflinePaymentForm(e) {
  e.preventDefault();
  const payType = document.getElementById('blePayType').value;
  const amount = parseInt(document.getElementById('bleAmount').value, 10);
  const user = getActiveWallet();

  if (isNaN(amount) || amount <= 0) {
    showAlert('Valid positive amount required', true);
    return;
  }

  let item = null;
  if (payType === 'pcn') {
    let chanId = (document.getElementById('bleChannelID').value || '').trim();
    let channels = getCachedWalletChannels(user);

    let ch = channels.find(c => c.id === chanId);
    if (!ch && channels.length > 0) {
      ch = channels[0];
      chanId = ch.id;
    }
    if (!chanId) {
      showAlert('Channel ID is required for PCN offline payment', true);
      return;
    }

    const currentSeq = ch ? (ch.sequence || 1) : 1;
    const currentBalA = ch ? ((ch.share && ch.share.balance_party_a !== undefined) ? ch.share.balance_party_a : ((ch.share && ch.share.BalancePartyA !== undefined) ? ch.share.BalancePartyA : (ch.balance_a !== undefined ? ch.balance_a : 0))) : 0;
    const currentBalB = ch ? ((ch.share && ch.share.balance_party_b !== undefined) ? ch.share.balance_party_b : ((ch.share && ch.share.BalancePartyB !== undefined) ? ch.share.BalancePartyB : (ch.balance_b !== undefined ? ch.balance_b : 0))) : 0;

    const partyA = ch ? ((ch.share && (ch.share.party_a || ch.share.PartyA)) || ch.party_a || user) : user;
    const partyB = ch ? ((ch.share && (ch.share.party_b || ch.share.PartyB)) || ch.party_b || '') : '';
    const isPartyB = partyB && user && partyB.toLowerCase() === user.toLowerCase();

    let newBalA, newBalB;
    if (isPartyB) {
      if (currentBalB < amount) {
        showAlert('Insufficient channel balance for offline payment (Current Bal: ' + currentBalB + ')', true);
        return;
      }
      newBalB = currentBalB - amount;
      newBalA = currentBalA + amount;
    } else {
      if (currentBalA < amount) {
        showAlert('Insufficient channel balance for offline payment (Current Bal: ' + currentBalA + ')', true);
        return;
      }
      newBalA = currentBalA - amount;
      newBalB = currentBalB + amount;
    }

    const newSeq = currentSeq + 1;

    item = {
      type: 'pcn_offline_payment',
      channel_id: chanId,
      sequence: newSeq,
      party_a: partyA,
      party_b: partyB,
      balance_a: newBalA,
      balance_b: newBalB,
      amount: amount,
      sender: user,
      status: 'Pending Host Sync',
      timestamp: Date.now()
    };
  } else {
    const receiver = (document.getElementById('bleReceiver').value || '').trim();
    if (!receiver) {
      showAlert('Receiver address required for direct transfer', true);
      return;
    }
    item = {
      type: 'direct_tx_offline',
      sender: user,
      receiver: receiver,
      amount: amount,
      nonce: Date.now(),
      status: 'Pending Host Sync',
      timestamp: Date.now()
    };
  }

  // 1. Attempt client-side offline signature via TweetNaCl if available
  if (window.nacl) {
    try {
      const clientKp = await getOrCreateClientKeypair();
      if (clientKp) {
        let hashBytes;
        if (payType === 'pcn') {
          const msg = item.channel_id + '|' + item.sequence + '|' + item.balance_a + '|' + item.balance_b;
          hashBytes = sha256Sync(msg);
        } else {
          item.sender = clientKp.address;
          const msg = item.sender + '|' + item.receiver + '|' + item.amount + '|' + item.nonce;
          hashBytes = sha256Sync(msg);
        }
        const sig = nacl.sign.detached(hashBytes, clientKp.secretKey);
        item.signature = bytesToHex(sig);
        item.public_key = bytesToHex(clientKp.publicKey);
      }
    } catch (localSignErr) {
      console.warn('Local client signing error:', localSignErr);
    }
  }

  // 2. Fallback to node wallet signing if online and not yet signed
  if (!item.signature || !item.public_key) {
    try {
      const signRes = await postData('/api/v1/wallet/sign-offline', {
        type: payType,
        channel_id: item.channel_id,
        sequence: item.sequence,
        balance_a: item.balance_a,
        balance_b: item.balance_b,
        receiver: item.receiver,
        amount: item.amount,
        nonce: item.nonce
      });
      if (signRes && signRes.signature && signRes.public_key) {
        item.signature = signRes.signature;
        item.public_key = signRes.public_key;
        if (signRes.sender && payType === 'direct') {
          item.sender = signRes.sender;
        }
      }
    } catch (signErr) {
      console.warn('Node wallet signing error:', signErr);
    }
  }

  if (!item.signature || !item.public_key) {
    showAlert('Failed to cryptographically sign payment payload.', true);
    return;
  }

  // 3. Render Offline Payment QR code for air-gapped P2P relay
  renderOfflinePaymentQr(item);

  // 4. If BLE characteristic is available, write directly over GATT
  const payloadStr = JSON.stringify(item);
  if (activeBleCharacteristic) {
    try {
      const enc = new TextEncoder();
      if (activeBleCharacteristic.writeValueWithResponse) {
        await activeBleCharacteristic.writeValueWithResponse(enc.encode(payloadStr));
      } else {
        await activeBleCharacteristic.writeValue(enc.encode(payloadStr));
      }
    } catch (err) {}
  }

  // 5. Attempt immediate synchronization if host node is currently reachable
  let alreadySynced = false;
  try {
    const bRes = await postData('/api/v1/transport/broadcast-offline', { payload: payloadStr });
    if (bRes && (bRes.status === 'offline_pcn_synced' || bRes.status === 'offline_tx_validated' || bRes.status === 'relayed_over_bluetooth')) {
      alreadySynced = true;
    }
  } catch (syncErr) {
    // Node is currently unreachable (true offline scenario)
  }

  if (alreadySynced) {
    showAlert('Payment signed and validated by Host node (' + amount + ' LQD)');
    refreshDashboardData();
  } else {
    // Buffer locally in outbox queue when offline
    const queue = getOfflineQueue();
    queue.push(item);
    saveOfflineQueue(queue);

    if (payType === 'pcn') {
      const channels = getCachedWalletChannels(user);
      const chIndex = channels.findIndex(c => c.id === item.channel_id);
      if (chIndex >= 0) {
        channels[chIndex].sequence = item.sequence;
        if (channels[chIndex].share) {
          channels[chIndex].share.balance_party_a = item.balance_a;
          channels[chIndex].share.balance_party_b = item.balance_b;
        }
        saveWalletChannels(user, channels);
        renderChannelsList(channels);
      }
    } else {
      const currentHoldings = getCachedWalletHoldings(user) || { balance: 1000, inr_balance: 1000 };
      if (currentHoldings.balance !== undefined) {
        currentHoldings.balance = Math.max(0, currentHoldings.balance - amount);
        saveWalletHoldings(user, currentHoldings);
        const heroBalEl = document.getElementById('heroTotalBalance');
        if (heroBalEl) heroBalEl.innerHTML = currentHoldings.balance + '<small>LQD</small>';
      }
    }

    showAlert('Offline mode: Payment signed and queued locally (' + amount + ' LQD). Present QR code to receiver.');
  }
}

async function pushOfflineQueueToHost() {
  const queue = getOfflineQueue();
  if (queue.length === 0) {
    showAlert('No offline payments in queue to push', true);
    return;
  }

  const hostUrlInput = document.getElementById('bleHostUrl');
  let hostUrl = (hostUrlInput && hostUrlInput.value.trim()) || window.location.origin;
  if (!/^https?:\/\//i.test(hostUrl)) {
    hostUrl = 'http://' + hostUrl;
  }
  hostUrl = hostUrl.replace(/\/+$/, '');

  const btn = document.getElementById('btnPushHost');
  if (btn) btn.disabled = true;

  showAlert('Pushing ' + queue.length + ' queued offline payments to Host (' + hostUrl + ')...');

  let successCount = 0;
  let remaining = [];
  let lastError = '';

  for (const item of queue) {
    try {
      let endpoint = hostUrl + '/api/v1/transport/broadcast-offline';
      let body = { payload: JSON.stringify(item) };

      if (item.type === 'pcn_offline_payment' || item.channel_id) {
        endpoint = hostUrl + '/api/v1/channels/offline-sync';
        body = {
          type: 'pcn_offline_payment',
          channel_id: item.channel_id,
          sequence: item.sequence,
          party_a: item.party_a,
          party_b: item.party_b,
          balance_a: item.balance_a,
          balance_b: item.balance_b,
          sender: item.sender,
          signature: item.signature,
          public_key: item.public_key
        };
      }

      const res = await postData(endpoint, body);
      if (res && (res.status === 'offline_pcn_synced' || res.status === 'offline_tx_validated' || res.status === 'relayed_over_bluetooth')) {
        successCount++;
      } else {
        remaining.push(item);
      }
    } catch (err) {
      lastError = err.message || (typeof err === 'string' ? err : 'Connection error');
      remaining.push(item);
    }
  }

  if (btn) btn.disabled = false;
  saveOfflineQueue(remaining);
  refreshDashboardData();

  if (successCount > 0) {
    showAlert('Successfully pushed & validated ' + successCount + ' payments on Host node! ');
  } else {
    showAlert('Validation failed with Host (' + hostUrl + '): ' + (lastError || 'Ensure Host is reachable'), true);
  }
}

function calcStakeYield() {
  const amt = parseFloat(document.getElementById('stakeAmount').value) || 0;
  const apyEl = document.getElementById('stakingApyVal');
  const apy = apyEl ? (parseFloat(apyEl.textContent) || 14.5) : 14.5;
  const yieldEst = (amt * (apy / 100)).toFixed(2);
  const prev = document.getElementById('stakeCalcPreview');
  if (prev) prev.textContent = 'Estimated Annual Yield: +' + yieldEst + ' LQD / year (' + apy + '% APY)';
}

function calcBondProfit() {
  const amt = parseFloat(document.getElementById('bondAmount').value) || 0;
  const pEl = document.getElementById('marketPriceVal');
  let price = 1.00;
  if (pEl) {
    const m = pEl.textContent.match(/[\d.]+/);
    if (m) price = parseFloat(m[0]) || 1.00;
  }
  const discount = Math.max(0, 1.00 - price);
  const profit = (amt * discount).toFixed(2);
  const prev = document.getElementById('bondCalcPreview');
  if (prev) prev.textContent = 'Expected Redemption Profit: +' + profit + ' LQD when price returns to ₹1.00 peg';
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
    const user = getActiveWallet();
    const counterparty = document.getElementById('chanCounterparty').value.trim();
    const amount = parseInt(document.getElementById('chanAmount').value, 10);
    if (!counterparty) { showAlert('Counterparty address required', true); return; }
    if (isNaN(amount) || amount <= 0) { showAlert('Valid deposit amount required', true); return; }

    const res = await postData('/api/v1/channels/open', { sender: user, counterparty, amount }, btn);
    showAlert('State channel opened: ' + res.id);
    closeSheet('sheetChannel');
    refreshDashboardData();
  } catch (err) { showAlert(err.message, true); }
}

async function submitUpdateChannel(e) {
  e.preventDefault();
  const btn = e.target.querySelector('button[type="submit"]');
  try {
    let channel_id = (document.getElementById('chanIdUpdate').value || '').trim();
    const customChan = document.getElementById('chanIdUpdateCustom');
    if (customChan && customChan.style.display !== 'none' && customChan.value.trim()) {
      channel_id = customChan.value.trim();
    }
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
    let channel_id = (document.getElementById('chanIdSettle').value || '').trim();
    const customChan = document.getElementById('chanIdSettleCustom');
    if (customChan && customChan.style.display !== 'none' && customChan.value.trim()) {
      channel_id = customChan.value.trim();
    }
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
  const currentWallet = localStorage.getItem(getStorageKey('active_wallet'));
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
    const newAddr = res.address;
    const initialLqd = res.balance || 1000;
    const initialInr = res.inr_balance || 1000;

    localStorage.setItem(getStorageKey('active_wallet'), newAddr);
    localStorage.setItem(getStorageKey('known_wallets'), JSON.stringify([newAddr]));

    // Persist new wallet initial balance immediately in localStorage
    saveWalletHoldings(newAddr, {
      balance: initialLqd,
      inr_balance: initialInr,
      inr_reserved: 0,
      staked_shares: 0,
      bonds_held: 0
    });

    // Reset scoped storage for channels and bridge for this brand-new wallet
    saveWalletChannels(newAddr, []);
    saveWalletBridgeMappings(newAddr, []);
    saveWalletOfflineQueue(newAddr, []);

    // Instantly update DOM with scoped empty lists and new balance
    await loadActiveWalletUI(newAddr, initialLqd, initialInr);
    renderChannelsList([]);
    renderBridgeList([]);
    refreshDashboardData();
    celebrateWalletCreation(newAddr, initialLqd, initialInr);
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
  const active = localStorage.getItem(getStorageKey('active_wallet'));
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
    const newAddr = res.address;
    const initialLqd = res.balance || 1000;
    const initialInr = res.inr_balance || 1000;

    localStorage.setItem(getStorageKey('active_wallet'), newAddr);
    localStorage.setItem(getStorageKey('known_wallets'), JSON.stringify([newAddr]));

    saveWalletHoldings(newAddr, {
      balance: initialLqd,
      inr_balance: initialInr,
      inr_reserved: 0,
      staked_shares: 0,
      bonds_held: 0
    });

    saveWalletChannels(newAddr, []);
    saveWalletBridgeMappings(newAddr, []);
    saveWalletOfflineQueue(newAddr, []);

    await loadActiveWalletUI(newAddr, initialLqd, initialInr);
    renderChannelsList([]);
    renderBridgeList([]);
    refreshDashboardData();
    celebrateWalletCreation(newAddr, initialLqd, initialInr);
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
  return localStorage.getItem(getStorageKey('active_wallet')) || '{{ .Address }}';
}

function getKnownWallets() {
  try {
    const list = JSON.parse(localStorage.getItem(getStorageKey('known_wallets')));
    if (Array.isArray(list) && list.length > 0) return list;
  } catch (e) {}
  return ['{{ .Address }}'];
}

function saveKnownWallet(addr) {
  if (!addr) return;
  const known = getKnownWallets();
  if (!known.includes(addr)) {
    known.push(addr);
    localStorage.setItem(getStorageKey('known_wallets'), JSON.stringify(known));
  }
}

async function loadActiveWalletUI(targetAddr, initialBalance, initialINRBalance) {
  const addr = targetAddr || getActiveWallet();
  localStorage.setItem(getStorageKey('active_wallet'), addr);
  saveKnownWallet(addr);

  const walletAddrEl = document.getElementById('walletAddress');
  if (walletAddrEl) walletAddrEl.textContent = addr;

  const heroBalEl = document.getElementById('heroTotalBalance');
  const inrBalEl = document.getElementById('heroINRBalance');
  const obInrEl = document.getElementById('obAvailableINR');

  if (initialBalance !== undefined) {
    saveWalletHoldings(addr, {
      balance: initialBalance,
      inr_balance: initialINRBalance
    });
  }

  // Load cached holdings immediately so offline mode or offline reload never shows node's balance!
  const cached = getCachedWalletHoldings(addr);
  const isNodeDefault = (addr === '{{ .Address }}');
  const fallbackBal = isNodeDefault ? {{ .Balance }} : 1000;
  const fallbackInr = isNodeDefault ? {{ .INRBalance }} : 1000;

  const currentBal = (cached && cached.balance !== undefined) ? cached.balance : (initialBalance !== undefined ? initialBalance : fallbackBal);
  const currentInr = (cached && cached.inr_balance !== undefined) ? cached.inr_balance : (initialINRBalance !== undefined ? initialINRBalance : fallbackInr);
  const resVal = (cached && cached.inr_reserved !== undefined) ? cached.inr_reserved : 0;

  if (heroBalEl) heroBalEl.innerHTML = currentBal + '<small>LQD</small>';
  const displayINR = '₹' + Number(currentInr).toFixed(2) + ' INR' + (resVal > 0 ? ' (' + Number(resVal).toFixed(2) + ' Reserved)' : '');
  if (inrBalEl) inrBalEl.textContent = displayINR;
  if (obInrEl) obInrEl.textContent = displayINR;

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
        const liveDisplayINR = '₹' + inrStr + ' INR' + (resVal > 0 ? ' (' + resVal.toFixed(2) + ' Reserved)' : '');
        if (inrBalEl) inrBalEl.textContent = liveDisplayINR;
        if (obInrEl) obInrEl.textContent = liveDisplayINR;
      }
      const stEl = document.getElementById('userStakedShares');
      if (stEl && holdings.staked_shares !== undefined) {
        stEl.textContent = holdings.staked_shares + ' Shares';
      }
      const bdEl = document.getElementById('userBondsHeld');
      if (bdEl && holdings.bonds_held !== undefined) {
        bdEl.textContent = holdings.bonds_held + ' Bonds';
      }
      saveWalletHoldings(addr, holdings);
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
  renderChannelsList(getCachedWalletChannels(addr));
  renderBridgeList(getCachedWalletBridgeMappings(addr));
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

function renderChannelsList(channels) {
  const countEl = document.getElementById('channelsCount');
  if (countEl) countEl.textContent = '(' + (channels ? channels.length : 0) + ')';

  populateChannelDropdowns(channels || []);

  const container = document.getElementById('channelsListContainer');
  if (!container) return;

  if (channels && channels.length > 0) {
    let html = '';
    channels.forEach(ch => {
      const isOpened = ch.state === 'OPEN' || ch.state === 0;
      const stateStr = isOpened ? 'OPEN' : ch.state;
      const badgeClass = isOpened ? 'badge-active' : 'badge-pending';
      const partyA = (ch.share && (ch.share.party_a || ch.share.PartyA)) || ch.party_a || '';
      const partyB = (ch.share && (ch.share.party_b || ch.share.PartyB)) || ch.party_b || '';
      const balA = (ch.share && ch.share.balance_party_a !== undefined) ? ch.share.balance_party_a : ((ch.share && ch.share.BalancePartyA !== undefined) ? ch.share.BalancePartyA : (ch.balance_a !== undefined ? ch.balance_a : 0));
      const balB = (ch.share && ch.share.balance_party_b !== undefined) ? ch.share.balance_party_b : ((ch.share && ch.share.BalancePartyB !== undefined) ? ch.share.BalancePartyB : (ch.balance_b !== undefined ? ch.balance_b : 0));
      const shortId = ch.id ? (ch.id.slice(0, 10) + '...' + ch.id.slice(-6)) : 'Channel';

      html += '<div class="list-row" style="flex-wrap:wrap; padding:12px 14px; background:var(--surface); border:1px solid var(--border-soft); border-radius:12px; margin-bottom:10px;">' +
        '<div class="row-main" style="min-width:180px;">' +
          '<div class="row-title" style="display:flex; align-items:center; gap:8px;">' +
            '<span style="font-family:var(--font-mono); font-weight:700;">' + shortId + '</span> ' +
            '<span class="badge ' + badgeClass + '">' + stateStr + '</span>' +
          '</div>' +
          '<div class="row-sub" style="font-size:0.75rem; margin-top:4px;">' + (partyA.slice(0, 8) + '...') + ' → ' + (partyB.slice(0, 8) + '...') + '</div>' +
        '</div>' +
        '<div class="row-side" style="display:flex; flex-direction:column; align-items:flex-end; gap:3px;">' +
          '<div class="top" style="color:var(--text); font-weight:700;">' + balA + ' / ' + balB + ' LQD</div>' +
          '<div class="bottom" style="font-size:0.72rem; color:var(--text-dim);">Sequence ' + (ch.sequence || 1) + '</div>' +
          '<div style="display:flex; gap:4px; margin-top:4px;">' +
            '<button type="button" class="btn btn-sm btn-ghost" onclick="openOfflinePaymentForChannel(\'' + ch.id + '\')" style="padding:3px 8px; font-size:0.72rem;">Pay Offline</button>' +
            '<button type="button" class="btn btn-sm btn-secondary" onclick="openReceiveOfflineModal()" style="padding:3px 8px; font-size:0.72rem;">Receive Offline</button>' +
          '</div>' +
        '</div></div>';
    });
    container.innerHTML = html;
    container.style.paddingBottom = '36px';
  } else {
    container.innerHTML = '<div class="empty-row">No active state channels yet. Tap "+ Open" to create one.</div>';
  }
}

function renderBridgeList(mappings) {
  const countEl = document.getElementById('bridgeCount');
  if (countEl) countEl.textContent = '(' + (mappings ? mappings.length : 0) + ')';

  const container = document.getElementById('bridgeListContainer');
  if (!container) return;

  if (mappings && mappings.length > 0) {
    let html = '';
    mappings.forEach(m => {
      const isSettled = m.status === 'SETTLED' || m.status === 2 || m.status === 'settled';
      const statusStr = isSettled ? 'SETTLED' : (m.status || 'PENDING');
      const badgeClass = isSettled ? 'badge-settled' : 'badge-pending';
      const txId = m.transaction_id || m.TransactionID || '';

      const verifyBtn = !isSettled ? '<button class="btn btn-sm btn-secondary" onclick="event.stopPropagation(); openVerifySheet(\'' + m.utr + '\');">Verify & Settle</button>' : '';
      html += '<div class="list-row" onclick="openVerifySheet(\'' + m.utr + '\')" style="cursor:pointer;"><div class="row-main">' +
        '<div class="row-title">' + m.utr + ' <span class="badge ' + badgeClass + '">' + statusStr + '</span></div>' +
        '<div class="row-sub">Tx ' + txId + '</div></div>' +
        '<div class="row-side" style="display:flex; flex-direction:column; align-items:flex-end; gap:4px;">' + verifyBtn + '<div class="bottom">' + (m.created_at || '') + '</div></div></div>';
    });
    container.innerHTML = html;
  } else {
    container.innerHTML = '<div class="empty-row">No fiat bridge requests yet.</div>';
  }
}

async function refreshChannelsUI() {
  const activeWallet = getActiveWallet();
  let channels = [];
  try {
    const res = await fetchData('/api/v1/channels/list?address=' + encodeURIComponent(activeWallet));
    let rawChannels = (res && res.channels) || [];
    channels = rawChannels.filter(ch => {
      if (!ch) return false;
      const partyA = ((ch.share && (ch.share.party_a || ch.share.PartyA)) || ch.party_a || '').toLowerCase();
      const partyB = ((ch.share && (ch.share.party_b || ch.share.PartyB)) || ch.party_b || '').toLowerCase();
      const u = activeWallet.toLowerCase();
      return partyA === u || partyB === u;
    });

    // Merge with locally cached channels so offline updates with higher sequences are preserved
    const cached = getCachedWalletChannels(activeWallet);
    if (cached && cached.length > 0) {
      cached.forEach(locCh => {
        const srvIdx = channels.findIndex(c => c.id === locCh.id);
        if (srvIdx >= 0) {
          if ((locCh.sequence || 0) > (channels[srvIdx].sequence || 0)) {
            channels[srvIdx] = locCh;
          }
        } else {
          channels.push(locCh);
        }
      });
    }

    // Also overlay any pending offline PCN payments in the outbox queue
    const queue = getOfflineQueue();
    if (queue && queue.length > 0) {
      queue.forEach(qItem => {
        if ((qItem.type === 'pcn_offline_payment' || qItem.channel_id) && qItem.channel_id) {
          const ch = channels.find(c => c.id === qItem.channel_id);
          if (ch && (qItem.sequence || 0) >= (ch.sequence || 0)) {
            ch.sequence = qItem.sequence;
            if (ch.share) {
              ch.share.balance_party_a = Number(qItem.balance_a);
              ch.share.balance_party_b = Number(qItem.balance_b);
            }
            ch.balance_a = Number(qItem.balance_a);
            ch.balance_b = Number(qItem.balance_b);
          }
        }
      });
    }

    saveWalletChannels(activeWallet, channels);
  } catch (err) {
    channels = getCachedWalletChannels(activeWallet);
  }
  renderChannelsList(channels);
}

async function refreshBridgeUI() {
  const activeWallet = getActiveWallet();
  let mappings = [];
  try {
    const res = await fetchData('/api/v1/bridge/list?address=' + encodeURIComponent(activeWallet));
    let rawMappings = (res && res.mappings) || [];
    mappings = rawMappings.filter(m => {
      if (!m) return false;
      const mAddr = (m.address || m.Address || '').toLowerCase();
      const u = activeWallet.toLowerCase();
      return !mAddr || mAddr === u;
    });
    saveWalletBridgeMappings(activeWallet, mappings);
  } catch (err) {
    mappings = getCachedWalletBridgeMappings(activeWallet);
  }
  renderBridgeList(mappings);
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

      // Update Peg Progress Gauge
      const priceVal = (typeof status.price === 'number') ? status.price : 1.00;
      const pct = Math.min(100, Math.max(0, Math.round((priceVal / 1.00) * 100)));
      const barEl = document.getElementById('pegProgressBar');
      const pctEl = document.getElementById('pegProgressPercent');
      if (barEl) barEl.style.width = pct + '%';
      if (pctEl) pctEl.textContent = pct + '%';
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
  const activeAddr = getActiveWallet();
  if (activeAddr && activeAddr !== '{{ .Address }}') {
    const cached = getCachedWalletHoldings(activeAddr);
    const heroBalEl = document.getElementById('heroTotalBalance');
    const inrBalEl = document.getElementById('heroINRBalance');
    const obInrEl = document.getElementById('obAvailableINR');

    const bal = (cached && cached.balance !== undefined) ? cached.balance : 1000;
    const inr = (cached && cached.inr_balance !== undefined) ? cached.inr_balance : 1000;
    const resInr = (cached && cached.inr_reserved !== undefined) ? cached.inr_reserved : 0;

    if (heroBalEl) heroBalEl.innerHTML = bal + '<small>LQD</small>';
    const displayINR = '₹' + Number(inr).toFixed(2) + ' INR' + (resInr > 0 ? ' (' + Number(resInr).toFixed(2) + ' Reserved)' : '');
    if (inrBalEl) inrBalEl.textContent = displayINR;
    if (obInrEl) obInrEl.textContent = displayINR;

    const walletAddrEl = document.getElementById('walletAddress');
    if (walletAddrEl) walletAddrEl.textContent = activeAddr;

    renderChannelsList(getCachedWalletChannels(activeAddr));
    renderBridgeList(getCachedWalletBridgeMappings(activeAddr));
  }
  checkFirstTimeWelcome();
  resetHostUrl();
  renderOfflineQueue();
  refreshDashboardData();
  setInterval(refreshDashboardData, 3000);
}

document.addEventListener('DOMContentLoaded', initAvatars);

/* ---------- Offline P2P Payment QR Generation & Camera Scanning ---------- */
function renderOfflinePaymentQr(item) {
  const card = document.getElementById('offlinePaymentQrCard');
  const canvas = document.getElementById('offlinePaymentQrCanvas');
  const details = document.getElementById('offlineQrDetails');
  if (!card || !canvas) return;

  canvas.innerHTML = '';
  // Build clean minimal wire payload
  const cleanItem = {
    type: item.type,
    channel_id: item.channel_id,
    sequence: item.sequence,
    party_a: item.party_a,
    party_b: item.party_b,
    balance_a: item.balance_a,
    balance_b: item.balance_b,
    amount: item.amount,
    sender: item.sender,
    receiver: item.receiver,
    nonce: item.nonce,
    signature: item.signature,
    public_key: item.public_key
  };
  const payloadStr = JSON.stringify(cleanItem);

  // Dynamic high-visibility responsive QR canvas size
  const qrSize = Math.min(280, Math.max(220, window.innerWidth - 80));

  new QRCode(canvas, {
    text: payloadStr,
    width: qrSize,
    height: qrSize,
    colorDark: "#000000",
    colorLight: "#ffffff",
    correctLevel: QRCode.CorrectLevel.L
  });

  const isPcn = item.type === 'pcn_offline_payment' || !!item.channel_id;
  if (details) {
    details.textContent = isPcn
      ? 'Channel Payment: ' + item.amount + ' LQD (Sequence ' + item.sequence + ')'
      : 'Direct Transfer: ' + item.amount + ' LQD to ' + (item.receiver || '').slice(0, 10) + '...';
  }
  card.style.display = 'block';
  card.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
}

function dismissOfflinePaymentQr() {
  const card = document.getElementById('offlinePaymentQrCard');
  if (card) card.style.display = 'none';
}

let offlineQrScanner = null;
let isProcessingOfflineQr = false;

async function startOfflinePaymentScan() {
  const btnStart = document.getElementById('btnStartOfflineScan');
  const btnStop = document.getElementById('btnStopOfflineScan');
  const hint = document.getElementById('offlineScanHint');
  if (!window.Html5Qrcode) {
    showAlert('Camera scanner loading, please wait...', true);
    return;
  }

  // Explicitly prompt for camera permission immediately on user interaction
  if (navigator.mediaDevices && navigator.mediaDevices.getUserMedia) {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ video: { facingMode: 'environment' } });
      stream.getTracks().forEach(t => t.stop());
    } catch (pErr) {
      console.warn('[Offline Camera] Permission request:', pErr);
    }
  }

  if (offlineQrScanner) {
    try { await offlineQrScanner.stop(); } catch(e) {}
    try { offlineQrScanner.clear(); } catch(e) {}
    offlineQrScanner = null;
  }

  if (hint) hint.innerHTML = 'Starting camera stream... Point at payer QR code.';

  offlineQrScanner = new Html5Qrcode('offlineScannerRegion', {
    formatsToSupport: window.Html5QrcodeSupportedFormats ? [ window.Html5QrcodeSupportedFormats.QR_CODE ] : undefined
  });

  const qrConfig = {
    fps: 10,
    qrbox: function(w, h) {
      const minEdge = Math.min(w, h);
      const edge = Math.max(160, Math.floor(minEdge * 0.85));
      return { width: edge, height: edge };
    }
  };

  try {
    await offlineQrScanner.start(
      { facingMode: 'environment' },
      qrConfig,
      onOfflineQrScanSuccess,
      () => {}
    );
    if (btnStart) btnStart.style.display = 'none';
    if (btnStop) btnStop.style.display = 'block';
    if (hint) hint.innerHTML = 'Point camera at payer QR code to verify.';
  } catch (err) {
    console.warn('[Offline Camera] facingMode failed, falling back to camera enumeration:', err);
    try {
      const cameras = await Html5Qrcode.getCameras();
      if (cameras && cameras.length > 0) {
        const backCamera = cameras.find(c => c.label.toLowerCase().includes('back') || c.label.toLowerCase().includes('rear')) || cameras[0];
        await offlineQrScanner.start(backCamera.id, qrConfig, onOfflineQrScanSuccess, () => {});
        if (btnStart) btnStart.style.display = 'none';
        if (btnStop) btnStop.style.display = 'block';
        if (hint) hint.innerHTML = 'Point camera at payer QR code to verify.';
        return;
      }
    } catch (camErr) {
      console.warn('[Offline Camera] Camera list fallback failed:', camErr);
    }
    if (hint) hint.innerHTML = 'Live stream paused. Tap <b>"Snap Camera"</b> to capture photo directly!';
    showAlert('Live camera paused. Tap "Snap Camera" to capture directly!', true);
  }
}

async function stopOfflinePaymentScan() {
  const btnStart = document.getElementById('btnStartOfflineScan');
  const btnStop = document.getElementById('btnStopOfflineScan');
  const hint = document.getElementById('offlineScanHint');
  if (offlineQrScanner) {
    try { await offlineQrScanner.stop(); } catch(e) {}
    try { offlineQrScanner.clear(); } catch(e) {}
    offlineQrScanner = null;
  }
  if (btnStart) btnStart.style.display = 'block';
  if (btnStop) btnStop.style.display = 'none';
  if (hint) hint.innerHTML = 'Scan the payer\'s offline payment QR code to verify Ed25519 signature and credit channel balance with zero internet.';
}

async function onOfflineQrScanSuccess(decodedText) {
  if (isProcessingOfflineQr) return;
  isProcessingOfflineQr = true;
  if (navigator.vibrate) {
    try { navigator.vibrate(120); } catch (_) {}
  }
  stopOfflinePaymentScan();
  try {
    let item;
    if (typeof decodedText === 'object' && decodedText !== null) {
      item = decodedText;
    } else {
      let raw = (decodedText || '').trim();
      if (raw.startsWith('%7B')) raw = decodeURIComponent(raw);
      item = JSON.parse(raw);
    }
    if (!item || !item.signature || !item.public_key) {
      showAlert('Invalid offline payment QR code', true);
      return;
    }

    // Verify cryptographic Ed25519 signature offline using sha256Sync
    if (window.nacl) {
      const pubBytes = hexToBytes(item.public_key);
      const sigBytes = hexToBytes(item.signature);
      let hashBytes;
      if (item.type === 'pcn_offline_payment' || item.channel_id) {
        const msg = item.channel_id + '|' + item.sequence + '|' + item.balance_a + '|' + item.balance_b;
        hashBytes = sha256Sync(msg);
      } else {
        const msg = item.sender + '|' + item.receiver + '|' + item.amount + '|' + item.nonce;
        hashBytes = sha256Sync(msg);
      }

      const verified = nacl.sign.detached.verify(hashBytes, sigBytes, pubBytes);
      if (!verified) {
        showAlert('Cryptographic signature verification failed on offline payment!', true);
        return;
      }
    }

    // Update local state channel balances if PCN
    const activeUser = getActiveWallet();
    if (item.channel_id) {
      let channels = getCachedWalletChannels(activeUser);
      let ch = channels.find(c => c.id === item.channel_id);
      if (ch) {
        ch.sequence = item.sequence;
        if (ch.share) {
          ch.share.balance_party_a = Number(item.balance_a);
          ch.share.balance_party_b = Number(item.balance_b);
        }
      } else {
        ch = {
          id: item.channel_id,
          party_a: item.party_a || item.sender,
          party_b: item.party_b || activeUser,
          state: 'OPEN',
          sequence: item.sequence || 1,
          share: {
            balance_party_a: Number(item.balance_a),
            balance_party_b: Number(item.balance_b)
          }
        };
        channels.push(ch);
      }
      saveWalletChannels(activeUser, channels);
      renderChannelsList(channels);
    } else {
      const currentHoldings = getCachedWalletHoldings(activeUser) || { balance: 1000, inr_balance: 1000 };
      const amt = Number(item.amount) || 0;
      currentHoldings.balance = (Number(currentHoldings.balance) || 0) + amt;
      currentHoldings.inr_balance = (Number(currentHoldings.inr_balance) || 0) + amt;
      saveWalletHoldings(activeUser, currentHoldings);
      const heroBalEl = document.getElementById('heroTotalBalance');
      if (heroBalEl) heroBalEl.innerHTML = currentHoldings.balance + '<small>LQD</small>';
    }

    // Add signed statement to receiver's outbox queue to push when reconnected
    const queue = getOfflineQueue();
    const exists = queue.some(q => q.signature === item.signature);
    if (!exists) {
      item.received_offline = true;
      item.status = 'Pending Host Sync';
      queue.push(item);
      saveOfflineQueue(queue);
    }

    // Attempt automatic broadcast to host if network reachable
    try {
      const payloadStr = typeof decodedText === 'string' ? decodedText : JSON.stringify(item);
      const bRes = await postData('/api/v1/transport/broadcast-offline', { payload: payloadStr });
      if (bRes && (bRes.status === 'offline_pcn_synced' || bRes.status === 'offline_tx_validated')) {
        item.status = 'Synced to Host';
        saveOfflineQueue(queue);
      }
    } catch (_) {}

    showAlert('Verified offline payment received! +' + item.amount + ' LQD' + (item.sequence ? ' (Seq ' + item.sequence + ')' : ''));
    refreshDashboardData();
    const sheetBt = document.getElementById('sheetBluetooth');
    if (sheetBt && sheetBt.classList.contains('open')) {
      switchSubpanel('subBleQueue', document.querySelectorAll('#sheetBluetooth .sheet-tabs button')[3], 'sheetBluetooth');
    }
  } catch (err) {
    showAlert('Failed to process offline payment: ' + (err.message || 'Invalid format'), true);
  } finally {
    isProcessingOfflineQr = false;
  }
}

async function handleOfflineQrImageUpload(input) {
  if (!input.files || input.files.length === 0) return;
  const file = input.files[0];
  const hint = document.getElementById('offlineScanHint');
  if (hint) hint.textContent = 'Decoding QR image...';

  // Stop live scanner if currently running
  if (offlineQrScanner) {
    try { await offlineQrScanner.stop(); } catch (_) {}
    try { offlineQrScanner.clear(); } catch (_) {}
    offlineQrScanner = null;
  }

  if (!window.Html5Qrcode) {
    showAlert('QR scanner loading, please wait...', true);
    return;
  }
  const fileScanner = new Html5Qrcode('offlineScannerRegion', {
    formatsToSupport: window.Html5QrcodeSupportedFormats ? [ window.Html5QrcodeSupportedFormats.QR_CODE ] : undefined
  });
  fileScanner.scanFile(file, false)
    .then(text => {
      fileScanner.clear();
      input.value = '';
      onOfflineQrScanSuccess(text);
    })
    .catch(() => {
      fileScanner.clear();
      input.value = '';
      if (hint) hint.innerHTML = 'Could not detect a valid QR code in that photo. Please try again.';
      showAlert('No valid offline payment QR code found in image', true);
    });
}


/* ---------- Channel Dropdown & Navigation Helpers ---------- */
function toggleCustomChannelInput() {
  const custom = document.getElementById('bleChannelCustomID');
  const btn = document.getElementById('btnToggleCustomChan');
  if (!custom) return;
  if (custom.style.display === 'none') {
    custom.style.display = 'block';
    custom.focus();
    if (btn) btn.textContent = 'Use Dropdown';
  } else {
    custom.style.display = 'none';
    if (btn) btn.textContent = '+ Custom ID';
  }
}

function toggleCustomUpdateChan() {
  const custom = document.getElementById('chanIdUpdateCustom');
  const btn = document.getElementById('btnToggleCustomUpdate');
  if (!custom) return;
  if (custom.style.display === 'none') {
    custom.style.display = 'block';
    custom.focus();
    if (btn) btn.textContent = 'Use Dropdown';
  } else {
    custom.style.display = 'none';
    if (btn) btn.textContent = '+ Custom ID';
  }
}

function toggleCustomSettleChan() {
  const custom = document.getElementById('chanIdSettleCustom');
  const btn = document.getElementById('btnToggleCustomSettle');
  if (!custom) return;
  if (custom.style.display === 'none') {
    custom.style.display = 'block';
    custom.focus();
    if (btn) btn.textContent = 'Use Dropdown';
  } else {
    custom.style.display = 'none';
    if (btn) btn.textContent = '+ Custom ID';
  }
}

function openReceiveOfflineModal() {
  openSheet('sheetBluetooth');
  const tabs = document.querySelectorAll('#sheetBluetooth .sheet-tabs button');
  if (tabs && tabs[1]) {
    switchSubpanel('subBleReceive', tabs[1], 'sheetBluetooth');
  }
}

function openOfflinePaymentForChannel(channelId) {
  openSheet('sheetBluetooth');
  switchSubpanel('subBlePayment', document.querySelectorAll('#sheetBluetooth .sheet-tabs button')[0], 'sheetBluetooth');
  const sel = document.getElementById('bleChannelID');
  if (sel) {
    sel.value = channelId;
  }
}

function populateChannelDropdowns(channels) {
  const dropdownIds = ['bleChannelID', 'chanIdUpdate', 'chanIdSettle'];
  dropdownIds.forEach(id => {
    const el = document.getElementById(id);
    if (!el) return;
    const currentVal = el.value;
    el.innerHTML = '<option value="">-- Select Active Channel --</option>';
    if (channels && channels.length > 0) {
      channels.forEach(ch => {
        const opt = document.createElement('option');
        opt.value = ch.id;
        const balA = ch.share ? (ch.share.balance_party_a !== undefined ? ch.share.balance_party_a : ch.share.BalancePartyA) : 0;
        const balB = ch.share ? (ch.share.balance_party_b !== undefined ? ch.share.balance_party_b : ch.share.BalancePartyB) : 0;
        const shortId = ch.id.slice(0, 8) + '...' + ch.id.slice(-6);
        opt.textContent = shortId + ' (Bal: ' + balA + ' / ' + balB + ' LQD - Seq ' + (ch.sequence || 1) + ')';
        el.appendChild(opt);
      });
      if (currentVal && channels.some(c => c.id === currentVal)) {
        el.value = currentVal;
      } else if (channels.length > 0 && id === 'bleChannelID') {
        el.value = channels[0].id;
      }
    } else {
      const opt = document.createElement('option');
      opt.value = '';
      opt.textContent = 'No channels open yet - tap "+ Open"';
      el.appendChild(opt);
    }
  });
}

</script>
</body>
</html>`

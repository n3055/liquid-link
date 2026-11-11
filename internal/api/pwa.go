package api

import (
	"net/http"
)

// [PWA] handleManifest serves the Progressive Web App manifest.json for mobile installation.
func (s *Server) handleManifest(w http.ResponseWriter, r *http.Request) {
	manifestJSON := `{
  "name": "LiquidLink - Offline P2P Payment Network",
  "short_name": "LiquidLink",
  "description": "Transport-agnostic off-chain payment network with offline BLE & QR state channel sync",
  "start_url": "/",
  "scope": "/",
  "display": "standalone",
  "orientation": "portrait",
  "background_color": "#090d16",
  "theme_color": "#090d16",
  "icons": [
    {
      "src": "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 192 192' fill='none'%3E%3Crect width='192' height='192' rx='42' fill='%230f172a'/%3E%3Cpath d='M96 36L44 76v70l52 40 52-40V76L96 36z' stroke='%2338bdf8' stroke-width='10' stroke-linejoin='round'/%3E%3Ccircle cx='96' cy='110' r='24' fill='%2338bdf8'/%3E%3C/svg%3E",
      "sizes": "192x192",
      "type": "image/svg+xml",
      "purpose": "any maskable"
    },
    {
      "src": "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 512 512' fill='none'%3E%3Crect width='512' height='512' rx='112' fill='%230f172a'/%3E%3Cpath d='M256 96L118 202v188l138 106 138-106V202L256 96z' stroke='%2338bdf8' stroke-width='24' stroke-linejoin='round'/%3E%3Ccircle cx='256' cy='292' r='64' fill='%2338bdf8'/%3E%3C/svg%3E",
      "sizes": "512x512",
      "type": "image/svg+xml",
      "purpose": "any maskable"
    }
  ]
}`
	w.Header().Set("Content-Type", "application/manifest+json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(manifestJSON))
}

// [PWA] handleServiceWorker serves the offline caching service worker script for standalone execution.
func (s *Server) handleServiceWorker(w http.ResponseWriter, r *http.Request) {
	swScript := `const CACHE_NAME = 'liquidlink-app-v2';
const STATIC_ASSETS = [
  '/',
  '/manifest.json',
  'https://cdnjs.cloudflare.com/ajax/libs/qrcodejs/1.0.0/qrcode.min.js',
  'https://cdnjs.cloudflare.com/ajax/libs/html5-qrcode/2.3.8/html5-qrcode.min.js',
  'https://cdn.jsdelivr.net/npm/tweetnacl@1.0.3/nacl-fast.min.js',
  'https://cdn.jsdelivr.net/npm/tweetnacl-util@0.15.1/nacl-util.min.js'
];

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => {
      return cache.addAll(STATIC_ASSETS).catch((err) => {
        console.warn('[PWA ServiceWorker] Pre-cache asset warning:', err);
      });
    }).then(() => self.skipWaiting())
  );
});

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((keys) => {
      return Promise.all(
        keys.map((key) => {
          if (key !== CACHE_NAME) {
            return caches.delete(key);
          }
        })
      );
    }).then(() => self.clients.claim())
  );
});

self.addEventListener('fetch', (event) => {
  const url = new URL(event.request.url);

  // Dynamic API calls: Network first, with graceful error fallback
  if (url.pathname.startsWith('/api/v1/')) {
    event.respondWith(
      fetch(event.request).catch(() => {
        return new Response(JSON.stringify({
          error: 'Node network unavailable (Offline Mode)',
          offline: true
        }), {
          status: 503,
          headers: { 'Content-Type': 'application/json' }
        });
      })
    );
    return;
  }

  // App Shell & Static Assets: Stale-While-Revalidate with offline fallback
  event.respondWith(
    caches.match(event.request).then((cachedResponse) => {
      if (cachedResponse) {
        // Fetch fresh copy in background if network is active
        fetch(event.request).then((networkResponse) => {
          if (networkResponse && networkResponse.status === 200) {
            caches.open(CACHE_NAME).then((cache) => {
              cache.put(event.request, networkResponse);
            });
          }
        }).catch(() => {});
        return cachedResponse;
      }
      return fetch(event.request).then((networkResponse) => {
        if (!networkResponse || networkResponse.status !== 200 || networkResponse.type !== 'basic') {
          return networkResponse;
        }
        const responseToCache = networkResponse.clone();
        caches.open(CACHE_NAME).then((cache) => {
          cache.put(event.request, responseToCache);
        });
        return networkResponse;
      }).catch(() => {
        // Fallback to cached root dashboard if available
        return caches.match('/');
      });
    })
  );
});
`
	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Service-Worker-Allowed", "/")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(swScript))
}

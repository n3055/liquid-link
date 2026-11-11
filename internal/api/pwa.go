package api

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"net/http"
	"sync"
)

var (
	icon192Bytes []byte
	icon512Bytes []byte
	iconOnce     sync.Once
)

func initIcons() {
	iconOnce.Do(func() {
		icon192Bytes = generatePngIcon(192)
		icon512Bytes = generatePngIcon(512)
	})
}

// generatePngIcon builds a raster PNG app icon with dark background and cyan emblem.
func generatePngIcon(size int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	// Background #090d16
	bg := color.RGBA{R: 9, G: 13, B: 22, A: 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{bg}, image.Point{}, draw.Src)

	// Inner rounded border/accent diamond #38bdf8
	accent := color.RGBA{R: 56, G: 189, B: 248, A: 255}
	center := size / 2
	radius := size * 3 / 8
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := x - center
			if dx < 0 {
				dx = -dx
			}
			dy := y - center
			if dy < 0 {
				dy = -dy
			}
			// Diamond shape
			if dx+dy <= radius && dx+dy >= radius-size/14 {
				img.Set(x, y, accent)
			}
			// Center circle
			if (x-center)*(x-center)+(y-center)*(y-center) <= (size/8)*(size/8) {
				img.Set(x, y, accent)
			}
		}
	}

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

// [PWA] handleIcon192 serves the 192x192 raster PNG app icon for Android WebAPK installation.
func (s *Server) handleIcon192(w http.ResponseWriter, r *http.Request) {
	initIcons()
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=604800")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(icon192Bytes)
}

// [PWA] handleIcon512 serves the 512x512 raster PNG splash icon for Android WebAPK installation.
func (s *Server) handleIcon512(w http.ResponseWriter, r *http.Request) {
	initIcons()
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=604800")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(icon512Bytes)
}

// [PWA] handleManifest serves the Progressive Web App manifest.json for mobile installation.
func (s *Server) handleManifest(w http.ResponseWriter, r *http.Request) {
	manifestJSON := `{
  "id": "/",
  "name": "LiquidLink Wallet",
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
      "src": "/icon-192.png",
      "sizes": "192x192",
      "type": "image/png",
      "purpose": "any maskable"
    },
    {
      "src": "/icon-512.png",
      "sizes": "512x512",
      "type": "image/png",
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
	swScript := `const CACHE_NAME = 'liquidlink-app-v3';
const STATIC_ASSETS = [
  '/',
  '/manifest.json',
  '/icon-192.png',
  '/icon-512.png',
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

package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// [PWA] TestServer_ManifestEndpoint verifies manifest.json metadata response headers and payload structure.
func TestServer_ManifestEndpoint(t *testing.T) {
	s, _, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/manifest.json", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for manifest, got %d", rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/manifest+json") {
		t.Errorf("expected application/manifest+json content type, got %s", contentType)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "LiquidLink") || !strings.Contains(body, "standalone") {
		t.Errorf("manifest payload missing required PWA fields: %s", body)
	}
}

// [PWA] TestServer_ServiceWorkerEndpoint verifies sw.js service worker script serving and headers.
func TestServer_ServiceWorkerEndpoint(t *testing.T) {
	s, _, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/sw.js", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for sw.js, got %d", rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if !strings.Contains(contentType, "javascript") {
		t.Errorf("expected javascript content type, got %s", contentType)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "CACHE_NAME") || !strings.Contains(body, "addEventListener") {
		t.Errorf("service worker missing cache logic: %s", body)
	}
}

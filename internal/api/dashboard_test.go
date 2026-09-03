package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDashboardRendering(t *testing.T) {
	s, _, cleanup := setupTestServer(t)
	defer cleanup()

	routes := []string{"/", "/ui"}

	for _, path := range routes {
		t.Run("GET "+path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rr := httptest.NewRecorder()

			s.Handler().ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("expected status 200 for %s, got %d", path, rr.Code)
			}

			contentType := rr.Header().Get("Content-Type")
			if !strings.Contains(contentType, "text/html") {
				t.Errorf("expected text/html content type for %s, got %s", path, contentType)
			}

			body := rr.Body.String()
			if !strings.Contains(body, "LiquidLink Wallet") {
				t.Errorf("expected dashboard title in body for %s", path)
			}
			if !strings.Contains(body, "test-node-1") {
				t.Errorf("expected node ID in dashboard body for %s", path)
			}
			if !strings.Contains(body, "Total balance") {
				t.Errorf("expected balance section in dashboard body for %s", path)
			}
			if !strings.Contains(body, "async function fetchData") {
				t.Errorf("expected fetchData helper in dashboard body for %s", path)
			}
			if !strings.Contains(body, "userStakedShares") {
				t.Errorf("expected userStakedShares element ID in dashboard body for %s", path)
			}
			if !strings.Contains(body, "sheetReceiveAddress") {
				t.Errorf("expected sheetReceiveAddress element ID in dashboard body for %s", path)
			}
			if !strings.Contains(body, "refreshDashboardData") {
				t.Errorf("expected refreshDashboardData live refetch helper in dashboard body for %s", path)
			}
			if !strings.Contains(body, "obBidsContainer") {
				t.Errorf("expected obBidsContainer element ID in dashboard body for %s", path)
			}
			if !strings.Contains(body, "channelsListContainer") {
				t.Errorf("expected channelsListContainer element ID in dashboard body for %s", path)
			}
			if !strings.Contains(body, "bridgeListContainer") {
				t.Errorf("expected bridgeListContainer element ID in dashboard body for %s", path)
			}
			if !strings.Contains(body, "heroINRBalance") {
				t.Errorf("expected heroINRBalance element ID in dashboard body for %s", path)
			}
			if !strings.Contains(body, "userOrdersContainer") {
				t.Errorf("expected userOrdersContainer element ID in dashboard body for %s", path)
			}
			if !strings.Contains(body, "obAvailableINR") {
				t.Errorf("expected obAvailableINR element ID in dashboard body for %s", path)
			}
			if !strings.Contains(body, "welcomeOverlay") {
				t.Errorf("expected welcomeOverlay onboarding screen in dashboard body for %s", path)
			}
			if !strings.Contains(body, "celebrationModal") {
				t.Errorf("expected celebrationModal animation screen in dashboard body for %s", path)
			}
			if !strings.Contains(body, "priceChartWrap") {
				t.Errorf("expected priceChartWrap container in dashboard body for %s", path)
			}
			if !strings.Contains(body, "metricsChartWrap") {
				t.Errorf("expected metricsChartWrap container in dashboard body for %s", path)
			}
		})
	}
}

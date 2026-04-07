package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"atlas-routex/internal/bootstrap"
)

func TestPlanEndpointCreatesItinerary(t *testing.T) {
	app := bootstrap.NewApp(bootstrap.LoadConfig())
	ts := httptest.NewServer(app.BuildMux())
	defer ts.Close()

	payload := map[string]any{
		"user_id":        "demo-user-1",
		"itinerary_name": "Tokyo Weekend",
		"destination":    map[string]any{"city": "Tokyo"},
		"start_date":     time.Date(2026, 4, 8, 0, 0, 0, 0, time.UTC),
		"end_date":       time.Date(2026, 4, 9, 0, 0, 0, 0, time.UTC),
		"budget":         map[string]any{"total": 800, "currency": "CNY"},
		"preferences":    map[string]any{"tags": []string{"culture", "food"}},
	}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(ts.URL+"/api/v1/plan", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("post plan: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data, ok := out["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected object data, got %#v", out["data"])
	}
	if data["id"] == nil || data["id"] == "" {
		t.Fatalf("expected created itinerary id, got %#v", data)
	}
}

func TestHealthz(t *testing.T) {
	app := bootstrap.NewApp(bootstrap.LoadConfig())
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	app.BuildMux().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

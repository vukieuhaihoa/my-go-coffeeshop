package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vukieuhaihoa/my-go-coffeeshop/internal/product/entity"
	"github.com/vukieuhaihoa/my-go-coffeeshop/pkg/logger"
)

func newTestHandler(t *testing.T) http.Handler {
	t.Helper()
	log, err := logger.NewSlogLogger(io.Discard, "debug")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	return NewHandler(log).Routes()
}

func TestGetItemTypes(t *testing.T) {
	h := newTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/v1/api/item-types", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", rec.Header().Get("Content-Type"))
	}

	got := []entity.ItemType{}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(got) != 10 {
		t.Errorf("expected 10 item types, got %d", len(got))
	}

	if got[0].Name != "CAPPUCCINO" {
		t.Errorf("expected first item type to be CAPPUCCINO, got %s", got[0].Name)
	}
}

func TestGetItemsByTypes(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantCount  int
	}{
		{"valid types", "/v1/api/items-by-types/CAPPUCCINO,ESPRESSO", http.StatusOK, 2},
		{"invalid type", "/v1/api/items-by-types/INVALID_TYPE", http.StatusBadRequest, 0},
		{"empty types", "/v1/api/items-by-types/", http.StatusNotFound, 0},
		{"trailing comma", "/v1/api/items-by-types/LATTE,", http.StatusOK, 1},
		{"leading comma", "/v1/api/items-by-types/,LATTE", http.StatusOK, 1},
		{"multiple commas", "/v1/api/items-by-types/LATTE,,ESPRESSO", http.StatusOK, 2},
		{"mixed valid and invalid types", "/v1/api/items-by-types/CAPPUCCINO,INVALID_TYPE", http.StatusBadRequest, 0},
		{"case insensitive", "/v1/api/items-by-types/cappuccino,espresso", http.StatusBadRequest, 0},
		{"duplicate types", "/v1/api/items-by-types/CAPPUCCINO,CAPPUCCINO", http.StatusOK, 2},
	}

	h := newTestHandler(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}

			if rec.Code == http.StatusOK {
				got := []entity.ItemType{}
				if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
					t.Fatalf("unmarshal response: %v", err)
				}
				if len(got) != tt.wantCount {
					t.Errorf("expected %d items, got %d", tt.wantCount, len(got))
				}
			}
		})
	}
}

func TestMethodNotAllowed(t *testing.T) {
	h := newTestHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/v1/api/item-types", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405 Method Not Allowed, got %d", rec.Code)
	}
}

func TestHealthz(t *testing.T) {
	h := newTestHandler(t)
	req := httptest.NewRequest("GET", "/healthz", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", rec.Header().Get("Content-Type"))
	}

	got := map[string]string{}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got["status"] != "ok" {
		t.Errorf("expected status 'ok', got %s", got["status"])
	}
}

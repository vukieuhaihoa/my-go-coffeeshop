package gateway

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vukieuhaihoa/my-go-coffeeshop/pkg/logger"
)

func TestLogRequests(t *testing.T) {
	tests := []struct {
		name           string
		handler        http.Handler
		wantStatus     int
		wantLogContent string
	}{
		{
			name: "successful request",
			handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
			wantStatus:     http.StatusOK,
			wantLogContent: "status=200",
		},
		{
			name: "internal server error",
			handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}),
			wantStatus:     http.StatusInternalServerError,
			wantLogContent: "status=500",
		},
		{
			name: "not found",
			handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			}),
			wantStatus:     http.StatusNotFound,
			wantLogContent: "status=404",
		},
		{
			name: "bad request",
			handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			}),
			wantStatus:     http.StatusBadRequest,
			wantLogContent: "status=400",
		},
		{
			name: "handler not writing header",
			handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				// Do nothing, should default to 200 OK
				_, _ = w.Write([]byte("ok")) // không gọi WriteHeader
			}),
			wantStatus:     http.StatusOK,
			wantLogContent: "status=200",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			log, err := logger.NewSlogLogger(&buf, "DEBUG")
			if err != nil {
				t.Fatalf("failed to create logger: %v", err)
			}

			h := LogRequests(log, tt.handler)

			req, err := http.NewRequest(http.MethodGet, "/", nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", w.Code, tt.wantStatus)
			}

			logOutput := buf.String()
			if !bytes.Contains([]byte(logOutput), []byte(tt.wantLogContent)) {
				t.Errorf("expected log output to contain '%s', got %s", tt.wantLogContent, logOutput)
			}
		})
	}
}

func TestAllowCORS(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot) // 418
	})

	h := AllowCORS([]string{"http://localhost:8888"}, next)

	tests := []struct {
		name       string
		method     string
		origin     string
		reqMethod  string
		wantStatus int
		wantACAO   string // Access-Control-Allow-Origin
	}{
		{
			name:       "no origin",
			method:     http.MethodGet,
			origin:     "",
			reqMethod:  "",
			wantStatus: http.StatusTeapot,
			wantACAO:   "",
		},
		{
			name:       "allowed preflight",
			method:     http.MethodOptions,
			origin:     "http://localhost:8888",
			reqMethod:  http.MethodGet,
			wantStatus: http.StatusNoContent,
			wantACAO:   "http://localhost:8888",
		},
		{
			name:       "allowed simple",
			method:     http.MethodGet,
			origin:     "http://localhost:8888",
			reqMethod:  "",
			wantStatus: http.StatusTeapot,
			wantACAO:   "http://localhost:8888",
		},
		{
			name:       "other origin simple",
			method:     http.MethodGet,
			origin:     "https://evil.example.com",
			reqMethod:  "",
			wantStatus: http.StatusTeapot,
			wantACAO:   "",
		},
		{
			name:       "other origin preflight",
			method:     http.MethodOptions,
			origin:     "https://evil.example.com",
			reqMethod:  http.MethodGet,
			wantStatus: http.StatusTeapot,
			wantACAO:   "",
		},
		{
			name:       "allowed preflight with no request method",
			method:     http.MethodOptions,
			origin:     "http://localhost:8888",
			reqMethod:  "",
			wantStatus: http.StatusTeapot,
			wantACAO:   "http://localhost:8888",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, "/", nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			if tt.reqMethod != "" {
				req.Header.Set("Access-Control-Request-Method", tt.reqMethod)
			}

			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)

			if w.Header().Get("Vary") != "Origin" {
				t.Errorf("got Vary header %q, want %q", w.Header().Get("Vary"), "Origin")
			}

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", w.Code, tt.wantStatus)
			}
			if gotACAO := w.Header().Get("Access-Control-Allow-Origin"); gotACAO != tt.wantACAO {
				t.Errorf("got Access-Control-Allow-Origin %q, want %q", gotACAO, tt.wantACAO)
			}
		})
	}
}

func TestWithTimeout(t *testing.T) {
	const d = time.Second

	var (
		hasDeadline bool
		left        time.Duration
		gotCtx      context.Context
	)
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		deadline, ok := r.Context().Deadline()
		hasDeadline = ok
		left = time.Until(deadline)
		gotCtx = r.Context()
	})

	WithTimeout(d, next).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if !hasDeadline {
		t.Fatal("next got a context without deadline")
	}
	if left <= 900*time.Millisecond || left > d {
		t.Errorf("deadline in %v, want (900ms, %v]", left, d)
	}

	if gotCtx.Err() == nil {
		t.Error("context was not cancelled after the handler returned")
	}
}

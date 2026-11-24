package gateway

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	productv1 "github.com/vukieuhaihoa/my-go-coffeeshop/gen/go/coffeeshop/product/v1"
	"github.com/vukieuhaihoa/my-go-coffeeshop/internal/product/catalog"
	"github.com/vukieuhaihoa/my-go-coffeeshop/internal/product/grpcapi"
)

func newTestGateway(t *testing.T) http.Handler {
	t.Helper()

	ls := bufconn.Listen(1 << 20)

	srv := grpc.NewServer()
	productv1.RegisterProductServiceServer(srv, grpcapi.NewServer(catalog.New()))

	go func() { _ = srv.Serve(ls) }()
	t.Cleanup(srv.Stop)

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return ls.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("failed to create gRPC client: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	gw, err := New(context.Background(), conn)
	if err != nil {
		t.Fatalf("failed to create gateway: %v", err)
	}
	return gw
}

func TestGateway(t *testing.T) {
	h := newTestGateway(t)
	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantCode   float64
	}{
		{
			name:       "GetItemTypes",
			method:     http.MethodGet,
			path:       "/v1/api/item-types",
			wantStatus: 200,
			wantCode:   -1,
		},
		{
			name:       "unknown",
			method:     http.MethodGet,
			path:       "/v1/api/items-by-types/PHO",
			wantStatus: 400,
			wantCode:   3,
		},
		{
			name:       "GetItemsByType",
			method:     http.MethodGet,
			path:       "/v1/api/items-by-types/ESPRESSO,CAPPUCCINO",
			wantStatus: 200,
			wantCode:   -1,
		},
		{
			name:       "wrong method",
			method:     http.MethodPost,
			path:       "/v1/api/items-by-types/ESPRESSO,CAPPUCCINO",
			wantStatus: 501,
			wantCode:   12,
		},
		{
			name:       "not found",
			method:     http.MethodGet,
			path:       "/v1/api/xxx/UNKNOWN",
			wantStatus: 404,
			wantCode:   5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.path, nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)
			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", w.Code, tt.wantStatus)
			}

			var body map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &body)
			if err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if code, ok := body["code"].(float64); ok && code != tt.wantCode {
				t.Errorf("got code %v, want %v", code, tt.wantCode)
			}
		})
	}
}

func TestGatewayJSONShape(t *testing.T) {
	h := newTestGateway(t)
	req, err := http.NewRequest(http.MethodGet, "/v1/api/item-types", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
	}

	var body struct {
		ItemTypes []map[string]any
	}
	err = json.Unmarshal(w.Body.Bytes(), &body)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(body.ItemTypes) != 10 {
		t.Errorf("expected 10 item types, got %d", len(body.ItemTypes))
	}
}

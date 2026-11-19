package grpcapi

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	productv1 "github.com/vukieuhaihoa/my-go-coffeeshop/gen/go/coffeeshop/product/v1"
	"github.com/vukieuhaihoa/my-go-coffeeshop/internal/product/catalog"
)

func newTestClient(t *testing.T) productv1.ProductServiceClient {
	t.Helper()
	ls := bufconn.Listen(1 << 20)

	srv := grpc.NewServer()
	productv1.RegisterProductServiceServer(srv, NewServer(catalog.New()))

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

	return productv1.NewProductServiceClient(conn)
}

func TestGetItemTypes(t *testing.T) {
	client := newTestClient(t)

	req := &productv1.GetItemTypesRequest{}
	res, err := client.GetItemTypes(context.Background(), req)
	if err != nil {
		t.Fatalf("GetItemTypes failed: %v", err)
	}

	if len(res.ItemTypes) == 0 {
		t.Fatalf("expected non-empty item types, got empty")
	}

	if res.ItemTypes[0].Name != "CAPPUCCINO" {
		t.Errorf("expected item type name 'CAPPUCCINO', got '%s'", res.ItemTypes[0].Name)
	}
}

func TestGetItemsByType(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCode  codes.Code
		wantCount []int32
	}{
		{
			name:      "valid types",
			input:     "CAPPUCCINO,ESPRESSO",
			wantCode:  codes.OK,
			wantCount: []int32{0, 3},
		},
		{
			name:      "unknown type",
			input:     "PHO",
			wantCode:  codes.InvalidArgument,
			wantCount: []int32{},
		},
		{
			name:      "mixed known and unknown types",
			input:     "CAPPUCCINO,PHO",
			wantCode:  codes.InvalidArgument,
			wantCount: []int32{},
		},
		{
			name:      "empty input",
			input:     "",
			wantCode:  codes.InvalidArgument,
			wantCount: []int32{},
		},
		{
			name:      "input with spaces",
			input:     "  CAPPUCCINO , ESPRESSO  ",
			wantCode:  codes.OK,
			wantCount: []int32{0, 3},
		},
		{
			name:      "duplicate types",
			input:     "LATTE,LATTE",
			wantCode:  codes.OK,
			wantCount: []int32{5, 5},
		},
		{
			name:      "just a comma",
			input:     ",",
			wantCode:  codes.InvalidArgument,
			wantCount: []int32{},
		},
	}

	client := newTestClient(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &productv1.GetItemsByTypeRequest{ItemTypes: tt.input}
			res, err := client.GetItemsByType(context.Background(), req)

			if res != nil && len(res.Items) != len(tt.wantCount) {
				t.Errorf("expected %d items, got %d", len(tt.wantCount), len(res.Items))
			}

			for i, want := range tt.wantCount {
				if res.Items[i].Type != want {
					t.Errorf("expected item type %d, got %d", want, res.Items[i].GetType())
				}
			}

			if err != nil {
				st, ok := status.FromError(err)
				if !ok {
					t.Fatalf("failed to parse gRPC error: %v", err)
				}
				if st.Code() != tt.wantCode {
					t.Errorf("expected gRPC code %v, got %v", tt.wantCode, st.Code())
				}
			} else if tt.wantCode != codes.OK {
				t.Errorf("expected gRPC code %v, got OK", tt.wantCode)
			}
		})
	}
}

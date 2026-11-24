package app

import (
	"bytes"
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/vukieuhaihoa/my-go-coffeeshop/cmd/product/config"
	productv1 "github.com/vukieuhaihoa/my-go-coffeeshop/gen/go/coffeeshop/product/v1"
	"github.com/vukieuhaihoa/my-go-coffeeshop/pkg/logger"

	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
)

func testConfig() *config.Config {
	return &config.Config{
		App: config.App{
			Name:    "product-test",
			Version: "1.0.0",
		},
	}
}

func newTestConn(t *testing.T, lis *bufconn.Listener) *grpc.ClientConn {
	t.Helper()
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("failed to create gRPC client: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	return conn
}

func TestRunServesAndShutsDown(t *testing.T) {
	var buf bytes.Buffer
	lg, err := logger.NewSlogLogger(&buf, "DEBUG")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	lis := bufconn.Listen(1 << 20)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runErr := make(chan error, 1)
	go func() {
		app := New(testConfig(), lg)
		runErr <- app.Run(ctx, lis)
	}()

	conn := newTestConn(t, lis)

	healthClient := healthv1.NewHealthClient(conn)
	req := &healthv1.HealthCheckRequest{
		Service: productv1.ProductService_ServiceDesc.ServiceName,
	}

	resp, err := healthClient.Check(context.Background(), req)
	if err != nil {
		t.Fatalf("failed to check health: %v", err)
	}
	if resp.Status != healthv1.HealthCheckResponse_SERVING {
		t.Fatalf("expected SERVING, got %v", resp.Status)
	}

	productClient := productv1.NewProductServiceClient(conn)
	_, err = productClient.GetItemTypes(context.Background(), &productv1.GetItemTypesRequest{})
	if err != nil {
		t.Fatalf("GetItemTypes failed: %v", err)
	}

	cancel()
	select {
	case err := <-runErr:
		if err != nil {
			t.Fatalf("Run = %v, want nil", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return in time after context cancellation")
	}

	logOutput := buf.String()

	if !bytes.Contains([]byte(logOutput), []byte("msg=rpc")) {
		t.Errorf("expected log output to contain 'msg=rpc', got %s", logOutput)
	}

	if !bytes.Contains([]byte(logOutput), []byte("GetItemTypes")) {
		t.Errorf("expected log output to contain 'GetItemTypes', got %s", logOutput)
	}

	if !bytes.Contains([]byte(logOutput), []byte("shutting down")) {
		t.Errorf("expected log output to contain 'shutting down', got %s", logOutput)
	}
	if !bytes.Contains([]byte(logOutput), []byte("msg=bye")) {
		t.Errorf("expected log output to contain 'msg=bye', got %s", logOutput)
	}
}

func TestRunWithCancelledContext(t *testing.T) {
	var buf bytes.Buffer
	lg, err := logger.NewSlogLogger(&buf, "DEBUG")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	lis := bufconn.Listen(1 << 20)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	runErr := make(chan error, 1)
	go func() {
		app := New(testConfig(), lg)
		runErr <- app.Run(ctx, lis)
	}()

	select {
	case err := <-runErr:
		if err != nil {
			t.Fatalf("Run = %v, want nil", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return in time after context cancellation")
	}

	logOutput := buf.String()

	if !bytes.Contains([]byte(logOutput), []byte("shutting down")) {
		t.Errorf("expected log output to contain 'shutting down', got %s", logOutput)
	}
	if !bytes.Contains([]byte(logOutput), []byte("msg=bye")) {
		t.Errorf("expected log output to contain 'msg=bye', got %s", logOutput)
	}
}

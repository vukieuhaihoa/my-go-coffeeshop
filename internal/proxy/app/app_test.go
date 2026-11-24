package app

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/vukieuhaihoa/my-go-coffeeshop/cmd/proxy/config"
	productv1 "github.com/vukieuhaihoa/my-go-coffeeshop/gen/go/coffeeshop/product/v1"
	"github.com/vukieuhaihoa/my-go-coffeeshop/internal/product/catalog"
	"github.com/vukieuhaihoa/my-go-coffeeshop/internal/product/grpcapi"
	"github.com/vukieuhaihoa/my-go-coffeeshop/pkg/logger"
)

func startProduct(t *testing.T) string {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	srv := grpc.NewServer()
	productv1.RegisterProductServiceServer(srv, grpcapi.NewServer(catalog.New()))
	go func() {
		_ = srv.Serve(lis)
	}()
	t.Cleanup(srv.Stop)

	return lis.Addr().String()
}

func TestRunProxiesToProduct(t *testing.T) {
	cfg := &config.Config{}
	cfg.ProductAddr = startProduct(t)

	var buf bytes.Buffer
	lg, err := logger.NewSlogLogger(&buf, "DEBUG")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runErr := make(chan error, 1)
	go func() {
		runErr <- New(cfg, lg).Run(ctx, lis)
	}()

	resp, err := http.Get("http://" + lis.Addr().String() + "/v1/api/items-by-types/LATTE")
	if err != nil {
		t.Fatalf("failed to make GET request: %v", err)
	}
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", resp.StatusCode)
	}

	cancel()
	select {
	case err := <-runErr:
		if err != nil {
			t.Fatalf("Run returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return in time after context cancellation")
	}

	logOutput := buf.String()
	if !bytes.Contains([]byte(logOutput), []byte("status=200")) {
		t.Errorf("expected log output to contain 'status=200', got %s", logOutput)
	}
	if !bytes.Contains([]byte(logOutput), []byte("msg=bye")) {
		t.Errorf("expected log output to contain 'msg=bye', got %s", logOutput)
	}
}

func TestRunProductDown(t *testing.T) {
	cfg := &config.Config{}

	var buf bytes.Buffer
	lg, err := logger.NewSlogLogger(&buf, "DEBUG")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	dead, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	cfg.ProductAddr = dead.Addr().String()
	_ = dead.Close() // Close immediately to simulate product service down

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runErr := make(chan error, 1)
	go func() {
		runErr <- New(cfg, lg).Run(ctx, lis)
	}()

	resp, err := http.Get("http://" + lis.Addr().String() + "/v1/api/items-by-types/LATTE")
	if err != nil {
		t.Fatalf("failed to make GET request: %v", err)
	}
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503 Service Unavailable, got %d", resp.StatusCode)
	}

	cancel()
	select {
	case err := <-runErr:
		if err != nil {
			t.Fatalf("Run returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return in time after context cancellation")
	}

	logOutput := buf.String()
	if !bytes.Contains([]byte(logOutput), []byte("status=503")) {
		t.Errorf("expected log output to contain 'status=503', got %s", logOutput)
	}

	if !bytes.Contains([]byte(logOutput), []byte("msg=bye")) {
		t.Errorf("expected log output to contain 'msg=bye', got %s", logOutput)
	}
}

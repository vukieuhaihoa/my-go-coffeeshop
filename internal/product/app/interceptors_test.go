package app

import (
	"bytes"
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/vukieuhaihoa/my-go-coffeeshop/pkg/logger"
)

func TestRecoverUnary(t *testing.T) {
	info := &grpc.UnaryServerInfo{
		FullMethod: "/coffeeshop.product.v1.ProductService/GetItemTypes",
	}

	var buf bytes.Buffer
	log, err := logger.NewSlogLogger(&buf, "DEBUG")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	panicky := func(context.Context, any) (any, error) {
		panic("oh no")
	}

	interceptor := recoverUnary(log)
	_, err = interceptor(context.Background(), nil, info, panicky)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %T", err)
	}

	if st.Code() != codes.Internal {
		t.Errorf("expected code %v, got %v", codes.Internal, st.Code())
	}
	if st.Message() != "internal error" {
		t.Errorf("expected message 'internal error', got '%s'", st.Message())
	}

	logOutput := buf.String()
	if !bytes.Contains([]byte(logOutput), []byte("panic recovered")) {
		t.Errorf("expected log output to contain 'panic recovered', got %s", logOutput)
	}
}

func TestLogUnary(t *testing.T) {
	info := &grpc.UnaryServerInfo{
		FullMethod: "/coffeeshop.product.v1.ProductService/GetItemTypes",
	}

	var buf bytes.Buffer
	log, err := logger.NewSlogLogger(&buf, "DEBUG")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	successful := func(context.Context, any) (any, error) {
		return "success", nil
	}

	interceptor := logUnary(log)
	_, err = interceptor(context.Background(), nil, info, successful)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	logOutput := buf.String()
	if !bytes.Contains([]byte(logOutput), []byte("rpc")) {
		t.Errorf("expected log output to contain 'rpc', got %s", logOutput)
	}
	if !bytes.Contains([]byte(logOutput), []byte(info.FullMethod)) {
		t.Errorf("expected log output to contain method name, got %s", logOutput)
	}
	if !bytes.Contains([]byte(logOutput), []byte(codes.OK.String())) {
		t.Errorf("expected log output to contain code OK, got %s", logOutput)
	}
}

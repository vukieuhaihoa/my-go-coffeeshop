package app

import (
	"bytes"
	"context"
	"errors"
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
	tests := []struct {
		name     string
		err      error
		wantCode string
	}{
		{"OK", nil, "code=OK"},
		{"InvalidArgument", status.Error(codes.InvalidArgument, "invalid argument"), "code=InvalidArgument"},
		{"Unknown", errors.New("unknown error"), "code=Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			log, err := logger.NewSlogLogger(&buf, "DEBUG")
			if err != nil {
				t.Fatalf("failed to create logger: %v", err)
			}
			info := &grpc.UnaryServerInfo{
				FullMethod: "/coffeeshop.product.v1.ProductService/GetItemTypes",
			}

			handler := func(context.Context, any) (any, error) {
				return "resp", tt.err
			}

			resp, err := logUnary(log)(context.Background(), nil, info, handler)
			if !errors.Is(err, tt.err) {
				t.Errorf("expected error %v, got %v", tt.err, err)
			}
			if resp != "resp" && tt.err == nil {
				t.Errorf("expected response to be 'resp', got %v", resp)
			}

			logOutput := buf.String()
			if !bytes.Contains([]byte(logOutput), []byte("msg=rpc")) {
				t.Errorf("expected log output to contain 'msg=rpc', got %s", logOutput)
			}
			if !bytes.Contains([]byte(logOutput), []byte("method="+info.FullMethod)) {
				t.Errorf("expected log output to contain method name, got %s", logOutput)
			}
			if !bytes.Contains([]byte(logOutput), []byte(tt.wantCode)) {
				t.Errorf("expected log output to contain '%s', got %s", tt.wantCode, logOutput)
			}
		})
	}
}

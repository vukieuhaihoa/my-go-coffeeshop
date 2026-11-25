// Package gateway provides the implementation of the proxy gateway for the coffee shop application.
package gateway

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	productv1 "github.com/vukieuhaihoa/my-go-coffeeshop/gen/go/coffeeshop/product/v1"
)

// New returns an http.Handler that serves the REST routes declared in the .proto files
// by calling the product service over conn.
func New(ctx context.Context, conn *grpc.ClientConn) (http.Handler, error) {
	mux := runtime.NewServeMux()
	if err := productv1.RegisterProductServiceHandler(ctx, mux, conn); err != nil {
		return nil, err
	}
	return mux, nil
}

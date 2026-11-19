package app

import (
	"context"
	"runtime/debug"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/vukieuhaihoa/my-go-coffeeshop/pkg/logger"
)

func recoverUnary(log logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		next grpc.UnaryHandler,
	) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered",
					"method", info.FullMethod,
					"panic", r,
					"stack", string(debug.Stack()),
				)
				err = status.Error(codes.Internal, "internal error")
			}
		}()
		return next(ctx, req)
	}
}

func logUnary(log logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		next grpc.UnaryHandler,
	) (resp any, err error) {
		start := time.Now()
		resp, err = next(ctx, req)
		log.Info("rpc",
			"method", info.FullMethod,
			"code", status.Code(err).String(),
			"took", time.Since(start))
		return resp, err
	}
}

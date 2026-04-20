package server

import (
	"context"
	"fmt"
	"net/http"

	"aeibi/api"
	"aeibi/internal/config"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// NewGatewayHandler builds the gRPC-Gateway HTTP handler.
func NewGatewayHandler(ctx context.Context, cfg *config.Config) (http.Handler, error) {
	mux := runtime.NewServeMux()

	// Gateway proxies to the local gRPC server.
	gatewayEndpoint := cfg.Server.GRPCAddr
	gatewayDialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	// Fail fast if any gateway handler cannot be registered.
	if err := api.RegisterUserServiceHandlerFromEndpoint(ctx, mux, gatewayEndpoint, gatewayDialOpts); err != nil {
		return nil, fmt.Errorf("register gateway handlers: %w", err)
	}
	if err := api.RegisterFollowServiceHandlerFromEndpoint(ctx, mux, gatewayEndpoint, gatewayDialOpts); err != nil {
		return nil, fmt.Errorf("register gateway handlers: %w", err)
	}
	if err := api.RegisterPostServiceHandlerFromEndpoint(ctx, mux, gatewayEndpoint, gatewayDialOpts); err != nil {
		return nil, fmt.Errorf("register gateway handlers: %w", err)
	}
	if err := api.RegisterFileServiceHandlerFromEndpoint(ctx, mux, gatewayEndpoint, gatewayDialOpts); err != nil {
		return nil, fmt.Errorf("register gateway handlers: %w", err)
	}
	if err := api.RegisterCommentServiceHandlerFromEndpoint(ctx, mux, gatewayEndpoint, gatewayDialOpts); err != nil {
		return nil, fmt.Errorf("register gateway handlers: %w", err)
	}
	if err := api.RegisterMessageServiceHandlerFromEndpoint(ctx, mux, gatewayEndpoint, gatewayDialOpts); err != nil {
		return nil, fmt.Errorf("register gateway handlers: %w", err)
	}
	if err := api.RegisterReportServiceHandlerFromEndpoint(ctx, mux, gatewayEndpoint, gatewayDialOpts); err != nil {
		return nil, fmt.Errorf("register gateway handlers: %w", err)
	}
	if err := api.RegisterConfigServiceHandlerFromEndpoint(ctx, mux, gatewayEndpoint, gatewayDialOpts); err != nil {
		return nil, fmt.Errorf("register gateway handlers: %w", err)
	}

	return mux, nil
}

// StartGateway starts the gRPC-Gateway HTTP server and returns it plus an error channel.
func StartGateway(ctx context.Context, cfg *config.Config) (*http.Server, <-chan error, error) {
	handler, err := NewGatewayHandler(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}

	httpServer, errCh := StartHTTPServer(cfg.Server.HTTPAddr, handler)
	return httpServer, errCh, nil
}

package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"aeibi/api"
	"aeibi/internal/config"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// StartGateway starts the gRPC-Gateway HTTP server and returns it plus an error channel.
func StartGateway(ctx context.Context, cfg *config.Config) (*http.Server, <-chan error, error) {
	mux := runtime.NewServeMux()

	// Gateway proxies to the local gRPC server.
	gatewayEndpoint := cfg.Server.GRPCAddr
	gatewayDialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	// Fail fast if any gateway handler cannot be registered.
	if err := api.RegisterUserServiceHandlerFromEndpoint(ctx, mux, gatewayEndpoint, gatewayDialOpts); err != nil {
		return nil, nil, fmt.Errorf("register gateway handlers: %w", err)
	}
	if err := api.RegisterFollowServiceHandlerFromEndpoint(ctx, mux, gatewayEndpoint, gatewayDialOpts); err != nil {
		return nil, nil, fmt.Errorf("register gateway handlers: %w", err)
	}
	if err := api.RegisterPostServiceHandlerFromEndpoint(ctx, mux, gatewayEndpoint, gatewayDialOpts); err != nil {
		return nil, nil, fmt.Errorf("register gateway handlers: %w", err)
	}
	if err := api.RegisterFileServiceHandlerFromEndpoint(ctx, mux, gatewayEndpoint, gatewayDialOpts); err != nil {
		return nil, nil, fmt.Errorf("register gateway handlers: %w", err)
	}
	if err := api.RegisterCommentServiceHandlerFromEndpoint(ctx, mux, gatewayEndpoint, gatewayDialOpts); err != nil {
		return nil, nil, fmt.Errorf("register gateway handlers: %w", err)
	}
	if err := api.RegisterMessageServiceHandlerFromEndpoint(ctx, mux, gatewayEndpoint, gatewayDialOpts); err != nil {
		return nil, nil, fmt.Errorf("register gateway handlers: %w", err)
	}
	if err := api.RegisterReportServiceHandlerFromEndpoint(ctx, mux, gatewayEndpoint, gatewayDialOpts); err != nil {
		return nil, nil, fmt.Errorf("register gateway handlers: %w", err)
	}

	httpServer := &http.Server{
		Addr:    cfg.Server.HTTPAddr,
		Handler: mux,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	return httpServer, errCh, nil
}

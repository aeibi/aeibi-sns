package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"aeibi/internal/auth"
	"aeibi/internal/config"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// StartGateway starts the gRPC-Gateway HTTP server and returns it plus an error channel.
func StartGateway(ctx context.Context, cfg *config.Config, registrars []ServiceRegistrar) (*http.Server, <-chan error, error) {
	mux := runtime.NewServeMux(
		runtime.WithMetadata(auth.GatewayMetadataExtractor),
	)
	for _, registrar := range registrars {
		if registrar.RegisterGateway == nil {
			continue
		}
		if err := registrar.RegisterGateway(ctx, mux); err != nil {
			return nil, nil, fmt.Errorf("register HTTP gateway %s: %w", registrar.Name, err)
		}
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

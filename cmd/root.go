package main

import (
	"aeibi/internal/config"
	"aeibi/internal/env"
	"aeibi/server"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "aeibi",
	Short: "Start the AeiBi backend server",
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath, err := cmd.Flags().GetString("config")
		if err != nil {
			return err
		}

		cfg, err := config.Load(configPath)
		if err != nil {
			return err
		}

		return RunRoot(cmd.Context(), cfg)
	},
}

// Run boots the application with the provided configuration.
func RunRoot(ctx context.Context, cfg *config.Config) error {
	// Initialize shared runtime dependencies.
	dbPool, err := env.InitDB(ctx, cfg.Database)
	if err != nil {
		return err
	}
	defer dbPool.Close()

	ossClient, err := env.InitOSS(ctx, cfg.OSS)
	if err != nil {
		return err
	}

	riverClient, err := env.InitRiverClient(cfg, dbPool)
	if err != nil {
		return err
	}

	_, riverErrCh, err := server.StartRiverWorker(ctx, riverClient)
	if err != nil {
		return err
	}

	// Start gRPC server
	grpcServer, grpcErrCh, err := server.StartGRPCServer(ctx, cfg, dbPool, ossClient, riverClient)
	if err != nil {
		stopCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if stopErr := riverClient.Stop(stopCtx); stopErr != nil {
			slog.Warn("stop river worker client", "error", stopErr)
		}
		return err
	}

	// Build gateway + frontend handlers on one HTTP server.
	gatewayHandler, err := server.NewGatewayHandler(ctx, cfg)
	if err != nil {
		grpcServer.GracefulStop()
		stopCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if stopErr := riverClient.Stop(stopCtx); stopErr != nil {
			slog.Warn("stop river worker client", "error", stopErr)
		}
		return err
	}

	frontendHandler, err := server.NewFrontendHandler()
	if err != nil {
		grpcServer.GracefulStop()
		stopCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if stopErr := riverClient.Stop(stopCtx); stopErr != nil {
			slog.Warn("stop river worker client", "error", stopErr)
		}
		return err
	}

	httpMux := http.NewServeMux()
	httpMux.Handle("/api/", gatewayHandler)
	httpMux.Handle("/file/", gatewayHandler)
	httpMux.Handle("/", frontendHandler)

	httpServer, httpErrCh := server.StartHTTPServer(cfg.Server.HTTPAddr, httpMux)

	slog.Info("gRPC server listening", "addr", cfg.Server.GRPCAddr)
	slog.Info("HTTP server listening", "addr", cfg.Server.HTTPAddr)
	slog.Info("web frontend available", "url", "http://localhost"+cfg.Server.HTTPAddr)

	// Wait for termination.
	var runErr error
	select {
	case err := <-riverErrCh:
		runErr = fmt.Errorf("river worker: %w", err)
	case err := <-grpcErrCh:
		runErr = fmt.Errorf("gRPC server: %w", err)
	case err := <-httpErrCh:
		runErr = fmt.Errorf("HTTP server: %w", err)
	case <-ctx.Done():
	}

	grpcServer.GracefulStop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Warn("HTTP shutdown", "error", err)
	}

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer stopCancel()
	if stopErr := riverClient.Stop(stopCtx); stopErr != nil {
		slog.Warn("stop river worker client", "error", stopErr)
	}

	return runErr
}

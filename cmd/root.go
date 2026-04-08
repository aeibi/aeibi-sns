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

		return RootRun(cmd.Context(), cfg)
	},
}

// Run boots the application with the provided configuration.
func RootRun(ctx context.Context, cfg *config.Config) error {
	// Initialize shared runtime dependencies.
	dbConn, err := env.InitDB(ctx, cfg.Database)
	if err != nil {
		return err
	}
	defer dbConn.Close()

	ossClient, err := env.InitOSS(ctx, cfg.OSS)
	if err != nil {
		return err
	}

	// Start gRPC server
	grpcServer, grpcErrCh, err := server.StartGRPCServer(ctx, cfg, dbConn, ossClient)
	if err != nil {
		return err
	}

	// Start gRPC-Gateway HTTP server
	httpServer, httpErrCh, err := server.StartGateway(ctx, cfg)
	if err != nil {
		grpcServer.GracefulStop()
		return err
	}

	slog.Info("gRPC server listening", "addr", cfg.Server.GRPCAddr)
	slog.Info("HTTP gateway listening", "addr", cfg.Server.HTTPAddr)

	// Wait for termination
	select {
	case err := <-grpcErrCh:
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Warn("HTTP shutdown after gRPC failure", "error", err)
		}
		return fmt.Errorf("gRPC server: %w", err)
	case err := <-httpErrCh:
		grpcServer.GracefulStop()
		return fmt.Errorf("HTTP server: %w", err)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		grpcServer.GracefulStop()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			slog.Warn("HTTP shutdown", "error", err)
		}
	}

	return nil
}

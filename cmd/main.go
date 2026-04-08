package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	rootCmd.PersistentFlags().String("config", "", "Path to config file")
	_ = rootCmd.MarkPersistentFlagRequired("config")

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		slog.Error("run", "error", err)
		os.Exit(1)
	}
}

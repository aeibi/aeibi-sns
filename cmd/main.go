package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"aeibi/cmd/server"
	"aeibi/internal/config"

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

		return server.Run(cmd.Context(), cfg)
	},
}

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

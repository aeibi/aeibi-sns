package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
)

// StartRiverWorker starts the River worker and returns the client and a stop channel.
func StartRiverWorker(ctx context.Context, riverClient *river.Client[pgx.Tx]) (*river.Client[pgx.Tx], <-chan error, error) {
	if err := riverClient.Start(context.WithoutCancel(ctx)); err != nil {
		return nil, nil, fmt.Errorf("start river worker client: %w", err)
	}

	errCh := make(chan error, 1)
	go func() {
		<-riverClient.Stopped()
		select {
		case <-ctx.Done():
		default:
			errCh <- errors.New("river worker stopped unexpectedly")
		}
	}()

	return riverClient, errCh, nil
}

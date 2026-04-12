package env

import (
	"fmt"

	"aeibi/internal/async"
	"aeibi/internal/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

func InitRiverClient(cfg *config.Config, pool *pgxpool.Pool) (*river.Client[pgx.Tx], error) {
	workers := river.NewWorkers()

	if err := river.AddWorkerSafely(workers, async.NewFollowInboxWorker(pool)); err != nil {
		return nil, fmt.Errorf("register follow inbox worker: %w", err)
	}
	if err := river.AddWorkerSafely(workers, async.NewCommentInboxWorker(pool)); err != nil {
		return nil, fmt.Errorf("register comment inbox worker: %w", err)
	}

	client, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Workers: workers,
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 100},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create river insert-only client: %w", err)
	}

	return client, nil
}

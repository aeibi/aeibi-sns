package env

import (
	"fmt"

	"aeibi/internal/async"
	searchrepo "aeibi/internal/repository/search"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

func InitRiverClient(pool *pgxpool.Pool, search *searchrepo.Search) (*river.Client[pgx.Tx], error) {
	workers := river.NewWorkers()

	if err := river.AddWorkerSafely(workers, async.NewFollowInboxWorker(pool)); err != nil {
		return nil, fmt.Errorf("register follow inbox worker: %w", err)
	}
	if err := river.AddWorkerSafely(workers, async.NewCommentInboxWorker(pool)); err != nil {
		return nil, fmt.Errorf("register comment inbox worker: %w", err)
	}
	if err := river.AddWorkerSafely(workers, async.NewUpdatePostSearchWorker(pool, search)); err != nil {
		return nil, fmt.Errorf("register post search worker: %w", err)
	}
	if err := river.AddWorkerSafely(workers, async.NewUpdateUserSearchWorker(pool, search)); err != nil {
		return nil, fmt.Errorf("register user search worker: %w", err)
	}
	if err := river.AddWorkerSafely(workers, async.NewUpdateTagSearchWorker(search)); err != nil {
		return nil, fmt.Errorf("register tag search worker: %w", err)
	}

	client, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Workers: workers,
		Queues: map[string]river.QueueConfig{
			async.QueueFollowInbox:  {MaxWorkers: 100},
			async.QueueCommentInbox: {MaxWorkers: 100},
			async.QueuePostSearch:   {MaxWorkers: 100},
			async.QueueUserSearch:   {MaxWorkers: 100},
			async.QueueTagSearch:    {MaxWorkers: 100},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create river insert-only client: %w", err)
	}

	return client, nil
}

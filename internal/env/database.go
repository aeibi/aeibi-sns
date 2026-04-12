package env

import (
	"context"
	"fmt"

	"aeibi/internal/config"
	"aeibi/internal/repository/db"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
)

// InitDB initializes the pgx pool, runs migrations, and verifies readiness.
func InitDB(ctx context.Context, cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	if err := db.Migration(cfg.MigrationsSource, pool); err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrate database: %w", err)
	}

	rivermigrator, err := rivermigrate.New(riverpgxv5.New(pool), nil)
	if err != nil {
		return nil, fmt.Errorf("create river migrator: %w", err)
	}

	if _, err := rivermigrator.Migrate(ctx, rivermigrate.DirectionUp, nil); err != nil {
		return nil, fmt.Errorf("run river migrations: %w", err)
	}

	return pool, nil
}

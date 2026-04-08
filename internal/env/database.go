package env

import (
	"context"
	"database/sql"
	"fmt"

	"aeibi/internal/config"
	"aeibi/internal/repository/db"

	_ "github.com/lib/pq"
)

// InitDB opens the database connection and pings it to ensure readiness.
func InitDB(ctx context.Context, cfg config.DatabaseConfig) (*sql.DB, error) {
	dbConn, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Migration(cfg.MigrationsSource, dbConn); err != nil {
		return nil, fmt.Errorf("migrate database: %w", err)
	}

	if err := dbConn.PingContext(ctx); err != nil {
		dbConn.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	return dbConn, nil
}

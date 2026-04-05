package db

import (
	"context"
	"database/sql"
)

func WithTx(
	ctx context.Context,
	dbx *sql.DB,
	db *Queries,
	fn func(*Queries) error,
) error {
	tx, err := dbx.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := db.WithTx(tx)
	if err := fn(qtx); err != nil {
		return err
	}
	return tx.Commit()
}

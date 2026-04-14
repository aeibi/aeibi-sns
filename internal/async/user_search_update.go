package async

import (
	"aeibi/internal/repository/db"
	searchrepo "aeibi/internal/repository/search"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
)

type UserSearchAction string

const (
	UserSearchActionUpsert UserSearchAction = "upsert"
	QueueUserSearch        string           = "search_user_update"
)

type UpdateUserSearchArgs struct {
	UserUID uuid.UUID        `json:"user_uid"`
	Action  UserSearchAction `json:"action"`
}

func (UpdateUserSearchArgs) Kind() string {
	return "search.user.update"
}

type UpdateUserSearchWorker struct {
	river.WorkerDefaults[UpdateUserSearchArgs]
	db     *db.Queries
	search *searchrepo.Search
}

func NewUpdateUserSearchWorker(pool *pgxpool.Pool, search *searchrepo.Search) *UpdateUserSearchWorker {
	return &UpdateUserSearchWorker{
		db:     db.New(pool),
		search: search,
	}
}

func (w *UpdateUserSearchWorker) Work(ctx context.Context, job *river.Job[UpdateUserSearchArgs]) error {
	if job.Args.UserUID == uuid.Nil {
		return fmt.Errorf("user uid is required")
	}

	switch job.Args.Action {
	case "", UserSearchActionUpsert:
		row, err := w.db.GetUserByUid(ctx, job.Args.UserUID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			return fmt.Errorf("get user by uid: %w", err)
		}

		if err := w.search.UpsertUsers([]searchrepo.UserDocument{
			{
				UID:         row.Uid.String(),
				Nickname:    row.Nickname,
				AvatarUrl:   row.AvatarUrl,
				Description: row.Description,
				Status:      string(row.Status),
			},
		}); err != nil {
			return fmt.Errorf("upsert user to search: %w", err)
		}
		return nil

	default:
		return fmt.Errorf("unsupported user search action: %q", job.Args.Action)
	}
}

func (p *Producer) EnqueueUpdateUserSearchTx(ctx context.Context, tx pgx.Tx, args UpdateUserSearchArgs) error {
	_, err := p.Client.InsertTx(ctx, tx, args, &river.InsertOpts{
		Queue: QueueUserSearch,
	})
	if err != nil {
		return fmt.Errorf("insert update user search job: %w", err)
	}

	return nil
}

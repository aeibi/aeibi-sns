package async

import (
	"aeibi/internal/repository/db"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
)

type FollowInboxArgs struct {
	MessageUID  uuid.UUID `json:"message_uid"`
	ReceiverUID uuid.UUID `json:"receiver_uid"`
	ActorUID    uuid.UUID `json:"actor_uid"`
}

const QueueFollowInbox = "inbox_follow"

func (FollowInboxArgs) Kind() string {
	return "inbox.follow"
}

type FollowInboxWorker struct {
	river.WorkerDefaults[FollowInboxArgs]
	db *db.Queries
}

func NewFollowInboxWorker(pool *pgxpool.Pool) *FollowInboxWorker {
	return &FollowInboxWorker{
		db: db.New(pool),
	}
}

func (w *FollowInboxWorker) Work(ctx context.Context, job *river.Job[FollowInboxArgs]) error {
	_, err := w.db.CreateFollowInboxMessage(ctx, db.CreateFollowInboxMessageParams{
		Uid:         job.Args.MessageUID,
		ReceiverUid: job.Args.ReceiverUID,
		ActorUid:    job.Args.ActorUID,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil
		}
		return fmt.Errorf("create follow inbox message: %w", err)
	}

	return nil
}

func (p *Producer) EnqueueFollowInboxTx(ctx context.Context, tx pgx.Tx, args FollowInboxArgs) error {
	_, err := p.Client.InsertTx(ctx, tx, args, &river.InsertOpts{
		Queue: QueueFollowInbox,
	})
	if err != nil {
		return fmt.Errorf("insert follow inbox job: %w", err)
	}

	return nil
}

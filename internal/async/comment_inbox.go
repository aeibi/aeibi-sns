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

type CommentInboxArgs struct {
	MessageUID  uuid.UUID `json:"message_uid"`
	ReceiverUID uuid.UUID `json:"receiver_uid"`
	ActorUID    uuid.UUID `json:"actor_uid"`
	CommentUID  uuid.UUID `json:"comment_uid"`
	PostUID     uuid.UUID `json:"post_uid"`
	ParentUID   uuid.UUID `json:"parent_uid"`
}

const QueueCommentInbox = "inbox_comment"

func (CommentInboxArgs) Kind() string {
	return "inbox.comment"
}

type CommentInboxWorker struct {
	river.WorkerDefaults[CommentInboxArgs]
	db *db.Queries
}

func NewCommentInboxWorker(pool *pgxpool.Pool) *CommentInboxWorker {
	return &CommentInboxWorker{
		db: db.New(pool),
	}
}

func (w *CommentInboxWorker) Work(ctx context.Context, job *river.Job[CommentInboxArgs]) error {
	_, err := w.db.CreateCommentInboxMessage(ctx, db.CreateCommentInboxMessageParams{
		Uid:              job.Args.MessageUID,
		ReceiverUid:      job.Args.ReceiverUID,
		ActorUid:         job.Args.ActorUID,
		CommentUid:       job.Args.CommentUID,
		PostUid:          job.Args.PostUID,
		ParentCommentUid: uuid.NullUUID{UUID: job.Args.ParentUID, Valid: job.Args.ParentUID != uuid.Nil},
	})
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return nil
	}
	if err != nil {
		return fmt.Errorf("create comment inbox message: %w", err)
	}

	return nil
}

func (p *Producer) EnqueueCommentInboxTx(ctx context.Context, tx pgx.Tx, args CommentInboxArgs) error {
	_, err := p.Client.InsertTx(ctx, tx, args, &river.InsertOpts{
		Queue: QueueCommentInbox,
	})
	if err != nil {
		return fmt.Errorf("insert comment inbox job: %w", err)
	}

	return nil
}

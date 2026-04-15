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

type PostSearchAction string

const (
	PostSearchActionUpsert PostSearchAction = "upsert"
	PostSearchActionDelete PostSearchAction = "delete"
	QueuePostSearch                         = "search_post_update"
)

type UpdatePostSearchArgs struct {
	PostUID uuid.UUID        `json:"post_uid"`
	Action  PostSearchAction `json:"action"`
}

func (UpdatePostSearchArgs) Kind() string {
	return "search.post.update"
}

type UpdatePostSearchWorker struct {
	river.WorkerDefaults[UpdatePostSearchArgs]
	db     *db.Queries
	search *searchrepo.Search
}

func NewUpdatePostSearchWorker(pool *pgxpool.Pool, search *searchrepo.Search) *UpdatePostSearchWorker {
	return &UpdatePostSearchWorker{
		db:     db.New(pool),
		search: search,
	}
}

func (w *UpdatePostSearchWorker) Work(ctx context.Context, job *river.Job[UpdatePostSearchArgs]) error {
	if job.Args.PostUID == uuid.Nil {
		return fmt.Errorf("post uid is required")
	}

	switch job.Args.Action {
	case "", PostSearchActionUpsert:
		row, err := w.db.GetPostByUid(ctx, db.GetPostByUidParams{
			Viewer: uuid.NullUUID{},
			Uid:    job.Args.PostUID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				if err := w.search.DeletePostsByUIDs([]string{job.Args.PostUID.String()}); err != nil {
					return fmt.Errorf("delete missing post from search: %w", err)
				}
				return nil
			}
			return fmt.Errorf("get post by uid: %w", err)
		}

		doc := searchrepo.PostDocument{
			UID:             row.Uid.String(),
			AuthorUID:       row.AuthorUid.String(),
			AuthorNickname:  row.AuthorNickname,
			Text:            row.Text,
			TagNames:        row.TagNames,
			Images:          row.Images,
			Attachments:     row.Attachments,
			ImageCount:      len(row.Images),
			AttachmentCount: len(row.Attachments),
			CommentCount:    int(row.CommentCount),
			CollectionCount: int(row.CollectionCount),
			LikeCount:       int(row.LikeCount),
			Pinned:          row.Pinned,
			Visibility:      string(row.Visibility),
			Status:          string(row.Status),
			LatestRepliedOn: row.LatestRepliedOn.Time.Unix(),
			CreatedAt:       row.CreatedAt.Time.Unix(),
			UpdatedAt:       row.UpdatedAt.Time.Unix(),
		}
		if err := w.search.UpsertPosts([]searchrepo.PostDocument{doc}); err != nil {
			return fmt.Errorf("upsert post to search: %w", err)
		}
		return nil

	case PostSearchActionDelete:
		if err := w.search.DeletePostsByUIDs([]string{job.Args.PostUID.String()}); err != nil {
			return fmt.Errorf("delete post from search: %w", err)
		}
		return nil

	default:
		return fmt.Errorf("unsupported post search action: %q", job.Args.Action)
	}
}

func (p *Producer) EnqueueUpdatePostSearchTx(ctx context.Context, tx pgx.Tx, args UpdatePostSearchArgs) error {
	_, err := p.Client.InsertTx(ctx, tx, args, &river.InsertOpts{
		Queue: QueuePostSearch,
	})
	if err != nil {
		return fmt.Errorf("insert update post search job: %w", err)
	}

	return nil
}

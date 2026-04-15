package async

import (
	searchrepo "aeibi/internal/repository/search"
	"aeibi/util"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
)

const QueueTagSearch = "search_tag_update"

type UpdateTagSearchArgs struct {
	TagNames []string `json:"tag_names"`
}

func (UpdateTagSearchArgs) Kind() string {
	return "search.tag.update"
}

type UpdateTagSearchWorker struct {
	river.WorkerDefaults[UpdateTagSearchArgs]
	search *searchrepo.Search
}

func NewUpdateTagSearchWorker(search *searchrepo.Search) *UpdateTagSearchWorker {
	return &UpdateTagSearchWorker{
		search: search,
	}
}

func (w *UpdateTagSearchWorker) Work(ctx context.Context, job *river.Job[UpdateTagSearchArgs]) error {
	names := util.NormalizeStrings(job.Args.TagNames)
	if len(names) == 0 {
		return nil
	}

	docs := make([]searchrepo.TagDocument, 0, len(names))
	for _, name := range names {
		docs = append(docs, searchrepo.TagDocument{
			ID:   name,
			Name: name,
		})
	}

	if err := w.search.UpsertTags(docs); err != nil {
		return fmt.Errorf("upsert tags to search: %w", err)
	}
	return nil
}

func (p *Producer) EnqueueUpdateTagSearchTx(ctx context.Context, tx pgx.Tx, args UpdateTagSearchArgs) error {
	_, err := p.Client.InsertTx(ctx, tx, args, &river.InsertOpts{
		Queue: QueueTagSearch,
	})
	if err != nil {
		return fmt.Errorf("insert update tag search job: %w", err)
	}

	return nil
}

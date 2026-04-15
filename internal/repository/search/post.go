package search

import (
	"fmt"
	"strconv"

	"github.com/meilisearch/meilisearch-go"
)

const IndexPosts = "posts"

type SearchPostsParams struct {
	Query     string
	ViewerUID string
	AuthorUID string
	TagName   string
	Limit     int64
	Offset    int64
	SortBy    string // "", "latest", "active", "hot"
}

type SearchPostsResult struct {
	Hits               []PostDocument `json:"hits"`
	EstimatedTotalHits int64          `json:"estimated_total_hits"`
	ProcessingTimeMs   int64          `json:"processing_time_ms"`
}

func (s *Search) setupPosts() error {
	if err := s.ensureIndex(IndexPosts, "uid"); err != nil {
		return err
	}

	task, err := s.client.Index(IndexPosts).UpdateSettings(&meilisearch.Settings{
		SearchableAttributes: []string{
			"tag_names",
			"author_nickname",
			"text",
		},
		DisplayedAttributes: []string{
			"uid",
			"author_uid",
			"author_nickname",
			"text",
			"tag_names",
			"images",
			"attachments",
			"image_count",
			"attachment_count",
			"comment_count",
			"collection_count",
			"like_count",
			"pinned",
			"visibility",
			"status",
			"latest_replied_on",
			"created_at",
			"updated_at",
		},
		FilterableAttributes: []string{
			"author_uid",
			"tag_names",
			"visibility",
			"status",
			"pinned",
			"created_at",
			"latest_replied_on",
		},
		SortableAttributes: []string{
			"created_at",
			"latest_replied_on",
			"comment_count",
			"collection_count",
			"like_count",
		},
		RankingRules: []string{
			"words",
			"typo",
			"proximity",
			"attribute",
			"sort",
			"exactness",
			"pinned:desc",
			"created_at:desc",
		},
	})
	if err != nil {
		return err
	}
	return s.waitTaskSucceeded(task)
}

func (s *Search) UpsertPosts(docs []PostDocument) error {
	if len(docs) == 0 {
		return nil
	}

	task, err := s.client.Index(IndexPosts).AddDocuments(docs, nil)
	if err != nil {
		return err
	}
	return s.waitTaskSucceeded(task)
}

func (s *Search) DeletePostsByUIDs(uids []string) error {
	if len(uids) == 0 {
		return nil
	}

	task, err := s.client.Index(IndexPosts).DeleteDocuments(uids, nil)
	if err != nil {
		return err
	}
	return s.waitTaskSucceeded(task)
}

func (s *Search) SearchPosts(p SearchPostsParams) (*SearchPostsResult, error) {
	if p.Limit <= 0 || p.Limit > 20 {
		p.Limit = 20
	}

	filters := []string{
		fmt.Sprintf("status = %s", strconv.Quote("NORMAL")),
	}

	if p.ViewerUID == "" {
		filters = append(filters, fmt.Sprintf("visibility = %s", strconv.Quote("PUBLIC")))
	} else {
		filters = append(filters,
			fmt.Sprintf("(%s OR %s)",
				fmt.Sprintf("visibility = %s", strconv.Quote("PUBLIC")),
				fmt.Sprintf("author_uid = %s", strconv.Quote(p.ViewerUID)),
			),
		)
	}

	if p.AuthorUID != "" {
		filters = append(filters, fmt.Sprintf("author_uid = %s", strconv.Quote(p.AuthorUID)))
	}
	if p.TagName != "" {
		filters = append(filters, fmt.Sprintf("tag_names = %s", strconv.Quote(p.TagName)))
	}

	req := &meilisearch.SearchRequest{
		Offset: p.Offset,
		Limit:  p.Limit,
		Filter: filters,
		AttributesToRetrieve: []string{
			"uid",
			"author_uid",
			"author_nickname",
			"text",
			"tag_names",
			"images",
			"attachments",
			"image_count",
			"attachment_count",
			"comment_count",
			"collection_count",
			"like_count",
			"pinned",
			"visibility",
			"status",
			"latest_replied_on",
			"created_at",
			"updated_at",
		},
		AttributesToHighlight: []string{"text"},
		AttributesToCrop:      []string{"text:40"},
	}

	switch p.SortBy {
	case "latest":
		req.Sort = []string{"created_at:desc"}
	case "active":
		req.Sort = []string{"latest_replied_on:desc"}
	case "hot":
		req.Sort = []string{"like_count:desc", "comment_count:desc", "created_at:desc"}
	}

	resp, err := s.client.Index(IndexPosts).Search(p.Query, req)
	if err != nil {
		return nil, err
	}

	var hits []PostDocument
	if err := resp.Hits.DecodeInto(&hits); err != nil {
		return nil, err
	}

	return &SearchPostsResult{
		Hits:               hits,
		EstimatedTotalHits: resp.EstimatedTotalHits,
		ProcessingTimeMs:   resp.ProcessingTimeMs,
	}, nil
}

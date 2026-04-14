package search

import (
	"github.com/meilisearch/meilisearch-go"
)

const IndexTags = "tags"

type SearchTagsParams struct {
	Query  string
	Limit  int64
	Offset int64
}

type SearchTagsResult struct {
	Hits               []TagDocument `json:"hits"`
	EstimatedTotalHits int64         `json:"estimated_total_hits"`
	ProcessingTimeMs   int64         `json:"processing_time_ms"`
}

func (s *Search) setupTags() error {
	if err := s.ensureIndex(IndexTags, "id"); err != nil {
		return err
	}

	task, err := s.client.Index(IndexTags).UpdateSettings(&meilisearch.Settings{
		SearchableAttributes: []string{"name"},
		DisplayedAttributes:  []string{"id", "name"},
	})
	if err != nil {
		return err
	}
	return s.waitTaskSucceeded(task)
}

func (s *Search) UpsertTags(docs []TagDocument) error {
	if len(docs) == 0 {
		return nil
	}

	task, err := s.client.Index(IndexTags).AddDocuments(docs, nil)
	if err != nil {
		return err
	}
	return s.waitTaskSucceeded(task)
}

func (s *Search) SearchTags(p SearchTagsParams) (*SearchTagsResult, error) {
	if p.Limit <= 0 || p.Limit > 20 {
		p.Limit = 20
	}

	resp, err := s.client.Index(IndexTags).Search(p.Query, &meilisearch.SearchRequest{
		Offset: p.Offset,
		Limit:  p.Limit,
		AttributesToRetrieve: []string{
			"id", "name",
		},
	})
	if err != nil {
		return nil, err
	}

	var hits []TagDocument
	if err := resp.Hits.DecodeInto(&hits); err != nil {
		return nil, err
	}

	return &SearchTagsResult{
		Hits:               hits,
		EstimatedTotalHits: resp.EstimatedTotalHits,
		ProcessingTimeMs:   resp.ProcessingTimeMs,
	}, nil
}

func (s *Search) SuggestTagsByName(prefix string, limit int64) (*SearchTagsResult, error) {
	if limit <= 0 || limit > 10 {
		limit = 10
	}

	resp, err := s.client.Index(IndexTags).Search(prefix, &meilisearch.SearchRequest{
		Limit: limit,
		AttributesToRetrieve: []string{
			"id",
			"name",
		},
	})
	if err != nil {
		return nil, err
	}

	var hits []TagDocument
	if err := resp.Hits.DecodeInto(&hits); err != nil {
		return nil, err
	}

	return &SearchTagsResult{
		Hits:               hits,
		EstimatedTotalHits: resp.EstimatedTotalHits,
		ProcessingTimeMs:   resp.ProcessingTimeMs,
	}, nil
}

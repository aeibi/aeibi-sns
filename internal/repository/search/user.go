package search

import (
	"github.com/meilisearch/meilisearch-go"
)

const IndexUsers = "users"

type SearchUsersParams struct {
	Query  string
	Limit  int64
	Offset int64
}

type SearchUsersResult struct {
	Hits               []UserDocument `json:"hits"`
	EstimatedTotalHits int64          `json:"estimated_total_hits"`
	ProcessingTimeMs   int64          `json:"processing_time_ms"`
}

func (s *Search) setupUsers() error {
	if err := s.ensureIndex(IndexUsers, "uid"); err != nil {
		return err
	}

	task, err := s.client.Index(IndexUsers).UpdateSettings(&meilisearch.Settings{
		SearchableAttributes: []string{
			"nickname",
			"description",
		},
		DisplayedAttributes: []string{
			"uid",
			"nickname",
			"avatar_url",
			"description",
			"status",
		},
		FilterableAttributes: []string{
			"status",
		},
	})
	if err != nil {
		return err
	}
	return s.waitTaskSucceeded(task)
}

func (s *Search) UpsertUsers(docs []UserDocument) error {
	if len(docs) == 0 {
		return nil
	}

	task, err := s.client.Index(IndexUsers).AddDocuments(docs, nil)
	if err != nil {
		return err
	}
	return s.waitTaskSucceeded(task)
}

func (s *Search) SearchUsers(p SearchUsersParams) (*SearchUsersResult, error) {
	if p.Limit <= 0 || p.Limit > 20 {
		p.Limit = 20
	}

	resp, err := s.client.Index(IndexUsers).Search(p.Query, &meilisearch.SearchRequest{
		Offset: p.Offset,
		Limit:  p.Limit,
		Filter: `status = "NORMAL"`,
		AttributesToRetrieve: []string{
			"uid", "nickname", "avatar_url", "description", "status",
		},
	})
	if err != nil {
		return nil, err
	}

	var hits []UserDocument
	if err := resp.Hits.DecodeInto(&hits); err != nil {
		return nil, err
	}

	return &SearchUsersResult{
		Hits:               hits,
		EstimatedTotalHits: resp.EstimatedTotalHits,
		ProcessingTimeMs:   resp.ProcessingTimeMs,
	}, nil
}

func (s *Search) SuggestUsersByNickname(prefix string, limit int64) (*SearchUsersResult, error) {
	if limit <= 0 || limit > 10 {
		limit = 10
	}

	resp, err := s.client.Index(IndexUsers).Search(prefix, &meilisearch.SearchRequest{
		Limit:  limit,
		Filter: `status = "NORMAL"`,
		AttributesToRetrieve: []string{
			"uid", "nickname", "avatar_url", "description", "status",
		},
	})
	if err != nil {
		return nil, err
	}

	var hits []UserDocument
	if err := resp.Hits.DecodeInto(&hits); err != nil {
		return nil, err
	}

	return &SearchUsersResult{
		Hits:               hits,
		EstimatedTotalHits: resp.EstimatedTotalHits,
		ProcessingTimeMs:   resp.ProcessingTimeMs,
	}, nil
}

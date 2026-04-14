package env

import (
	"context"
	"fmt"

	"aeibi/internal/config"
	searchrepo "aeibi/internal/repository/search"

	"github.com/meilisearch/meilisearch-go"
)

func InitSearch(ctx context.Context, cfg config.SearchConfig) (*searchrepo.Search, error) {
	client, err := meilisearch.Connect(
		cfg.Host,
		meilisearch.WithAPIKey(cfg.APIKey),
	)
	if err != nil {
		return nil, fmt.Errorf("create meilisearch client: %w", err)
	}

	if _, err := client.HealthWithContext(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("check meilisearch health: %w", err)
	}

	search := searchrepo.New(client)
	if err := search.Setup(); err != nil {
		client.Close()
		return nil, fmt.Errorf("setup search indexes: %w", err)
	}

	return search, nil
}

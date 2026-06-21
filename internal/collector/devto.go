package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type DevToCollector struct {
	client *http.Client
}

func NewDevToCollector() *DevToCollector {
	return &DevToCollector{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (d *DevToCollector) Name() string {
	return "devto"
}

type devToArticle struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Score       int    `json:"positive_reactions_count"`
	PublishedAt string `json:"published_timestamp"`
}

func (d *DevToCollector) Collect(ctx context.Context) ([]*CollectedItem, error) {
	// Fetch articles from dev.to API
	// Tag can be customized. Let's pull programming, go, python, etc.
	req, err := http.NewRequestWithContext(ctx, "GET", "https://dev.to/api/articles?state=rising&per_page=30", nil)
	if err != nil {
		return nil, err
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch dev.to articles: %w", err)
	}
	defer resp.Body.Close()

	var articles []devToArticle
	if err := json.NewDecoder(resp.Body).Decode(&articles); err != nil {
		return nil, fmt.Errorf("failed to decode dev.to response: %w", err)
	}

	var items []*CollectedItem
	for _, a := range articles {
		items = append(items, &CollectedItem{
			Title:       a.Title,
			URL:         a.URL,
			Source:      d.Name(),
			Score:       float64(a.Score),
			PublishedAt: a.PublishedAt,
		})
	}

	return items, nil
}

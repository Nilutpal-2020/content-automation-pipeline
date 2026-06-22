package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type HackerNewsCollector struct {
	client *http.Client
}

func NewHackerNewsCollector() *HackerNewsCollector {
	return &HackerNewsCollector{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (h *HackerNewsCollector) Name() string {
	return "hackernews"
}

func (h *HackerNewsCollector) Collect(ctx context.Context) ([]*CollectedItem, error) {
	// 1. Get top stories
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://hacker-news.firebaseio.com/v0/topstories.json", nil)
	if err != nil {
		return nil, fmt.Errorf("create top stories request: %w", err)
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch top stories: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("fetch top stories: unexpected status %s", resp.Status)
	}

	var storyIDs []int
	if err := json.NewDecoder(resp.Body).Decode(&storyIDs); err != nil {
		return nil, fmt.Errorf("failed to decode top stories: %w", err)
	}

	// Limit to top 30 for now to avoid long collection times
	limit := 30
	if len(storyIDs) < limit {
		limit = len(storyIDs)
	}

	var items []*CollectedItem
	for i := 0; i < limit; i++ {
		id := storyIDs[i]
		item, err := h.fetchStory(ctx, id)
		if err != nil {
			// Skip errors for individual stories to not halt collection
			continue
		}
		if item.URL != "" { // Only collect stories with URLs
			items = append(items, item)
		}
	}

	return items, nil
}

type hnStory struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Score int    `json:"score"`
	Time  int64  `json:"time"`
	Text  string `json:"text"`
}

func (h *HackerNewsCollector) fetchStory(ctx context.Context, id int) (*CollectedItem, error) {
	url := fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json", id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("fetch story %d: unexpected status %s", id, resp.Status)
	}

	var story hnStory
	if err := json.NewDecoder(resp.Body).Decode(&story); err != nil {
		return nil, err
	}

	return &CollectedItem{
		Title:       story.Title,
		URL:         story.URL,
		Source:      h.Name(),
		Score:       float64(story.Score),
		PublishedAt: time.Unix(story.Time, 0).Format(time.RFC3339),
		Summary:     story.Text,
	}, nil
}

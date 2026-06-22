package scheduler

import (
	"context"
	"testing"

	"content-automation-pipeline/internal/collector"
	"content-automation-pipeline/internal/filter"
	"content-automation-pipeline/internal/generator"
	"content-automation-pipeline/internal/publisher"
	"content-automation-pipeline/pkg/logger"
	"go.uber.org/zap"
)

type fakeCollector struct{ items []*collector.CollectedItem }

func (f fakeCollector) Name() string { return "fake" }
func (f fakeCollector) Collect(context.Context) ([]*collector.CollectedItem, error) {
	return f.items, nil
}

type fakeGenerator struct{}

func (fakeGenerator) RewriteArticle(_ context.Context, _, title, _, _ string) (*generator.GeneratedContent, error) {
	return &generator.GeneratedContent{PostText: title, Hashtags: "#test", ImagePrompt: "test prompt"}, nil
}

type fakePublisher struct {
	existing map[string]bool
	published []publisher.PublishRequest
}

func (f *fakePublisher) HasPublished(_ context.Context, key string) (bool, error) { return f.existing[key], nil }
func (f *fakePublisher) Publish(_ context.Context, req publisher.PublishRequest) error {
	f.published = append(f.published, req)
	return nil
}

func TestRunCycleQueuesRankedItemsPerCategoryAndSkipsExisting(t *testing.T) {
	logger.Log = zap.NewNop()
	items := []*collector.CollectedItem{
		{Title: "OpenAI launches a new model", URL: "https://example.com/ai", Score: 800},
		{Title: "Postgres backend patterns", URL: "https://example.com/backend", Score: 700},
		{Title: "Kubernetes deployment guide", URL: "https://example.com/devops", Score: 600},
		{Title: "Browser engines get faster", URL: "https://example.com/news", Score: 500},
	}
	existingKey := publisher.ContentKey(items[0].Title, items[0].URL)
	pub := &fakePublisher{existing: map[string]bool{existingKey: true}}
	pipeline := NewScheduler([]collector.Collector{fakeCollector{items: items}}, filter.NewScorer(), fakeGenerator{}, pub, 1)

	pipeline.RunCycle(context.Background())

	if got, want := len(pub.published), 3; got != want {
		t.Fatalf("published %d items, want %d", got, want)
	}
	seenCategories := make(map[string]bool)
	for _, req := range pub.published {
		seenCategories[req.Category] = true
		if req.ContentKey == "" {
			t.Fatal("published request had no content key")
		}
	}
	for _, category := range []string{filter.CategoryBackend, filter.CategoryDevOps, filter.CategoryTechNews} {
		if !seenCategories[category] {
			t.Errorf("missing queued category %q", category)
		}
	}
}

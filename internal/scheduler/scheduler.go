package scheduler

import (
	"context"
	"sort"

	"content-automation-pipeline/internal/collector"
	"content-automation-pipeline/internal/filter"
	"content-automation-pipeline/internal/generator"
	"content-automation-pipeline/internal/publisher"
	"content-automation-pipeline/pkg/logger"
	"go.uber.org/zap"
)

type Scheduler struct {
	collectors []collector.Collector
	scorer     *filter.Scorer
	gen        generator.Generator
	pub        publisher.Publisher
}

func NewScheduler(collectors []collector.Collector, scorer *filter.Scorer, gen generator.Generator, pub publisher.Publisher) *Scheduler {
	return &Scheduler{
		collectors: collectors,
		scorer:     scorer,
		gen:        gen,
		pub:        pub,
	}
}

// RunCycle executes one complete content pipeline cycle
func (s *Scheduler) RunCycle(ctx context.Context) {
	logger.Log.Info("Starting pipeline cycle")

	var allItems []*collector.CollectedItem

	// 1. Collect
	for _, c := range s.collectors {
		items, err := c.Collect(ctx)
		if err != nil {
			logger.Log.Error("Failed to collect items", zap.String("source", c.Name()), zap.Error(err))
			continue
		}
		allItems = append(allItems, items...)
		logger.Log.Info("Collected items", zap.String("source", c.Name()), zap.Int("count", len(items)))
	}

	// 2. Deduplicate by URL
	seenURLs := make(map[string]bool)
	var uniqueItems []*collector.CollectedItem
	for _, item := range allItems {
		if !seenURLs[item.URL] {
			seenURLs[item.URL] = true
			uniqueItems = append(uniqueItems, item)
		}
	}

	// 3. Score & Rank Top 10
	for _, item := range uniqueItems {
		item.Score = s.scorer.Score(item)
	}

	// Sort by Score descending
	sort.Slice(uniqueItems, func(i, j int) bool {
		return uniqueItems[i].Score > uniqueItems[j].Score
	})

	limit := 10
	if len(uniqueItems) < limit {
		limit = len(uniqueItems)
	}
	topItems := uniqueItems[:limit]

	if len(topItems) == 0 {
		logger.Log.Info("No items found to process")
		return
	}

	// 4. Generate & 5. Publish to Notion
	for _, item := range topItems {
		logger.Log.Info("Processing top item", zap.String("title", item.Title), zap.Float64("score", item.Score))

		genContent, err := s.gen.RewriteArticle(ctx, item.Title, item.URL, "")
		if err != nil {
			logger.Log.Error("Failed to generate post", zap.Error(err))
			continue
		}

		req := publisher.PublishRequest{
			Title:       item.Title,
			Category:    "Tech News",
			SourceURL:   item.URL,
			PostText:    genContent.PostText,
			Hashtags:    genContent.Hashtags,
			ImagePrompt: genContent.ImagePrompt,
		}

		if err := s.pub.Publish(ctx, req); err != nil {
			logger.Log.Error("Failed to publish post to Notion", zap.Error(err))
			continue
		}
	}

	logger.Log.Info("Pipeline cycle completed successfully")
}

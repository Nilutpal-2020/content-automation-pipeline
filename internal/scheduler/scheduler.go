package scheduler

import (
	"context"

	"content-automation-pipeline/internal/collector"
	"content-automation-pipeline/internal/filter"
	"content-automation-pipeline/internal/generator"
	"content-automation-pipeline/internal/publisher"
	"content-automation-pipeline/internal/store"
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

	// 2. Score & Filter
	var topItem *collector.CollectedItem
	maxScore := -1.0

	for _, item := range allItems {
		score := s.scorer.Score(item)
		item.Score = score

		// Save to MongoDB
		article := &store.Article{
			Title:       item.Title,
			URL:         item.URL,
			Source:      item.Source,
			Score:       item.Score,
			Posted:      false,
		}
		if err := store.SaveArticle(ctx, article); err != nil {
			logger.Log.Error("Failed to save article", zap.Error(err))
		}

		if score > maxScore {
			maxScore = score
			topItem = item
		}
	}

	if topItem == nil {
		logger.Log.Info("No items found to process")
		return
	}

	logger.Log.Info("Selected top item", zap.String("title", topItem.Title), zap.Float64("score", maxScore))

	// 3. Generate Post
	postText, err := s.gen.RewriteArticle(ctx, topItem.Title, topItem.URL, "")
	if err != nil {
		logger.Log.Error("Failed to generate post", zap.Error(err))
		return
	}

	// 4. Publish
	if err := s.pub.Publish(ctx, postText); err != nil {
		logger.Log.Error("Failed to publish post", zap.Error(err))
		return
	}

	// TODO: Update MongoDB to mark item as Posted = true

	logger.Log.Info("Pipeline cycle completed successfully")
}

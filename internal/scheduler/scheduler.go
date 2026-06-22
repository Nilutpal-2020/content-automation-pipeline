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
	collectors       []collector.Collector
	scorer           *filter.Scorer
	gen              generator.Generator
	pub              publisher.Publisher
	postsPerCategory int
}

func NewScheduler(collectors []collector.Collector, scorer *filter.Scorer, gen generator.Generator, pub publisher.Publisher, postsPerCategory int) *Scheduler {
	if postsPerCategory < 1 {
		postsPerCategory = 3
	}
	return &Scheduler{
		collectors:       collectors,
		scorer:           scorer,
		gen:              gen,
		pub:              pub,
		postsPerCategory: postsPerCategory,
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

	// 3. Score, categorize, and rank. A separate ranking per category keeps a
	// popular topic from crowding out the rest of the daily content calendar.
	itemsByCategory := make(map[string][]*collector.CollectedItem)
	for _, item := range uniqueItems {
		item.Score = s.scorer.Score(item)
		category := filter.Categorize(item)
		itemsByCategory[category] = append(itemsByCategory[category], item)
	}

	if len(itemsByCategory) == 0 {
		logger.Log.Info("No items found to process")
		return
	}

	// 4. Generate & 5. Publish to Notion. We keep each category's quota
	// independent, which is more useful than a global top-N for a social queue.
	for _, category := range []string{filter.CategoryAI, filter.CategoryBackend, filter.CategoryDevOps, filter.CategoryMinimalist, filter.CategoryProductivity, filter.CategoryTechNews} {
		items := itemsByCategory[category]
		sort.Slice(items, func(i, j int) bool { return items[i].Score > items[j].Score })
		limit := s.postsPerCategory
		if len(items) < limit {
			limit = len(items)
		}

		for _, item := range items[:limit] {
			contentKey := publisher.ContentKey(item.Title, item.URL)
			alreadyPublished, err := s.pub.HasPublished(ctx, contentKey)
			if err != nil {
				logger.Log.Error("Failed to check idempotency key", zap.String("category", category), zap.String("title", item.Title), zap.Error(err))
				continue
			}
			if alreadyPublished {
				logger.Log.Info("Skipping already queued content", zap.String("category", category), zap.String("title", item.Title))
				continue
			}

			logger.Log.Info("Processing ranked item", zap.String("category", category), zap.String("title", item.Title), zap.Float64("score", item.Score))

			genContent, err := s.gen.RewriteArticle(ctx, category, item.Title, item.URL, item.Summary)
			if err != nil {
				logger.Log.Error("Failed to generate post", zap.String("category", category), zap.Error(err))
				continue
			}

			req := publisher.PublishRequest{
				Title:       item.Title,
				Category:    category,
				SourceURL:   item.URL,
				PostText:    genContent.PostText,
				Hashtags:    genContent.Hashtags,
				ImagePrompt: genContent.ImagePrompt,
				ContentKey:  contentKey,
			}

			if err := s.pub.Publish(ctx, req); err != nil {
				logger.Log.Error("Failed to publish post to Notion", zap.String("category", category), zap.Error(err))
				continue
			}
		}
	}

	logger.Log.Info("Pipeline cycle completed successfully")
}

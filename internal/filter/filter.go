package filter

import (
	"strings"
	"time"

	"content-automation-pipeline/internal/collector"
)

type Scorer struct {
	weights  map[string]float64
	keywords []string
}

func NewScorer() *Scorer {
	return &Scorer{
		weights: map[string]float64{
			"popularity": 0.5,
			"recency":    0.3,
			"relevance":  0.2,
		},
		keywords: []string{"ai", "go", "golang", "python", "backend", "system design", "docker", "kubernetes", "aws"},
	}
}

func (s *Scorer) Score(item *collector.CollectedItem) float64 {
	// 1. Popularity (Normalize score roughly between 0-1)
	popScore := item.Score / 500.0 // Assuming 500 is very popular
	if popScore > 1.0 {
		popScore = 1.0
	}

	// 2. Recency (Decay based on hours since published)
	recencyScore := 0.0
	if item.PublishedAt != "" {
		if parsedTime, err := time.Parse(time.RFC3339, item.PublishedAt); err == nil {
			hoursOld := time.Since(parsedTime).Hours()
			if hoursOld >= 0 {
				recencyScore = 1.0 / ((hoursOld / 24.0) + 1.0) // 1.0 if new, approaches 0 as days increase
			}
		}
	} else {
		recencyScore = 0.5 // Default if unknown
	}

	// 3. Relevance (Keyword matching in title)
	relevanceScore := 0.1
	lowerTitle := strings.ToLower(item.Title)
	for _, kw := range s.keywords {
		if strings.Contains(lowerTitle, kw) {
			relevanceScore = 1.0
			break
		}
	}

	finalScore := (popScore * s.weights["popularity"]) +
		(recencyScore * s.weights["recency"]) +
		(relevanceScore * s.weights["relevance"])

	return finalScore
}

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

const (
	CategoryAI           = "AI"
	CategoryBackend      = "Backend"
	CategoryDevOps       = "DevOps"
	CategoryMinimalist   = "Minimalist"
	CategoryProductivity = "Productivity"
	CategoryTechNews     = "Tech News"
)

// Categorize assigns every article one queue category. Ordering is intentional:
// operational topics should not be swallowed by the broader Backend category.
func Categorize(item *collector.CollectedItem) string {
	text := strings.ToLower(item.Title + " " + item.Summary)
	for _, category := range []struct {
		name     string
		keywords []string
	}{
		{CategoryProductivity, []string{"productivity", "focus", "deep work", "habit", "workflow", "attention", "decision fatigue", "time management", "routine"}},
		{CategoryMinimalist, []string{"minimalism", "minimalist", "simplicity", "simple", "essential", "less is more", "declutter", "intentional"}},
		{CategoryAI, []string{"artificial intelligence", "machine learning", "llm", "openai", "anthropic", "gemini", "model", "agent"}},
		{CategoryDevOps, []string{"kubernetes", "docker", "terraform", "ci/cd", "cloud", "aws", "observability", "devops", "deploy"}},
		{CategoryBackend, []string{"golang", "go ", "python", "database", "api", "backend", "server", "postgres", "distributed system"}},
	} {
		for _, keyword := range category.keywords {
			if strings.Contains(text, keyword) {
				return category.name
			}
		}
	}
	return CategoryTechNews
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

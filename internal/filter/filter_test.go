package filter

import (
	"testing"
	"time"

	"content-automation-pipeline/internal/collector"
)

func TestScorer_Score(t *testing.T) {
	scorer := NewScorer()
	now := time.Now()

	tests := []struct {
		name     string
		item     *collector.CollectedItem
		minScore float64
	}{
		{
			name: "High popularity, recent, relevant",
			item: &collector.CollectedItem{
				Title:       "How Go handles memory in 2026",
				Score:       1000,
				PublishedAt: now.Format(time.RFC3339),
			},
			minScore: 0.9,
		},
		{
			name: "Low popularity, old, irrelevant",
			item: &collector.CollectedItem{
				Title:       "My dog's birthday",
				Score:       5,
				PublishedAt: now.Add(-720 * time.Hour).Format(time.RFC3339), // 30 days old
			},
			minScore: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scorer.Score(tt.item)
			if got < tt.minScore {
				t.Errorf("Scorer.Score() = %v, want >= %v", got, tt.minScore)
			}
		})
	}
}

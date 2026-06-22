package collector

import "context"

type CollectedItem struct {
	Title       string
	URL         string
	Source      string
	Score       float64
	PublishedAt string
	Summary     string
}

type Collector interface {
	Collect(ctx context.Context) ([]*CollectedItem, error)
	Name() string
}

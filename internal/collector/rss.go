package collector

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// RSSCollector supports both RSS 2.0 and Atom feeds. It provides a low-cost,
// source-agnostic way to add trusted publisher feeds without another scraper.
type RSSCollector struct {
	name    string
	feedURL string
	limit   int
	client  *http.Client
}

func NewRSSCollector(name, feedURL string, limit int) *RSSCollector {
	if limit < 1 {
		limit = 20
	}
	return &RSSCollector{
		name:    name,
		feedURL: feedURL,
		limit:   limit,
		client:  &http.Client{Timeout: 15 * time.Second},
	}
}

func (r *RSSCollector) Name() string { return r.name }

func (r *RSSCollector) Collect(ctx context.Context) ([]*CollectedItem, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create %s feed request: %w", r.name, err)
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s feed: %w", r.name, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("fetch %s feed: unexpected status %s", r.name, resp.Status)
	}

	var feed rssFeed
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, fmt.Errorf("decode %s feed: %w", r.name, err)
	}

	entries := append(feed.Channel.Items, feed.Entries...)
	if len(entries) > r.limit {
		entries = entries[:r.limit]
	}
	items := make([]*CollectedItem, 0, len(entries))
	for _, entry := range entries {
		url := entry.Link.Href
		if url == "" {
			url = entry.Link.Value
		}
		if entry.Title == "" || url == "" {
			continue
		}
		items = append(items, &CollectedItem{
			Title:       cleanFeedText(entry.Title),
			URL:         url,
			Source:      r.name,
			PublishedAt: normalizeFeedTime(entry.PublishedAt, entry.AtomPublishedAt, entry.AtomUpdatedAt),
			Summary:     cleanFeedText(firstNonEmpty(entry.Description, entry.Summary, entry.Content)),
		})
	}
	return items, nil
}

type rssFeed struct {
	Channel struct {
		Items []rssEntry `xml:"item"`
	} `xml:"channel"`
	Entries []rssEntry `xml:"entry"`
}

type rssEntry struct {
	Title           string   `xml:"title"`
	Link            feedLink `xml:"link"`
	Description     string   `xml:"description"`
	Summary         string   `xml:"summary"`
	Content         string   `xml:"content"`
	PublishedAt     string   `xml:"pubDate"`
	AtomPublishedAt string   `xml:"published"`
	AtomUpdatedAt   string   `xml:"updated"`
}

type feedLink struct {
	Href  string `xml:"href,attr"`
	Value string `xml:",chardata"`
}

func normalizeFeedTime(values ...string) string {
	for _, value := range values {
		for _, layout := range []string{time.RFC3339, time.RFC1123Z, time.RFC1123, time.RFC822Z, time.RFC822} {
			if parsed, err := time.Parse(layout, strings.TrimSpace(value)); err == nil {
				return parsed.Format(time.RFC3339)
			}
		}
	}
	return ""
}

func cleanFeedText(value string) string {
	return strings.Join(strings.Fields(strings.NewReplacer("<![CDATA[", "", "]]>", "", "<p>", " ", "</p>", " ", "<br>", " ").Replace(value)), " ")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

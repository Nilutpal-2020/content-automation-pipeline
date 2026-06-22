package collector

import (
	"encoding/xml"
	"testing"
)

func TestRSSFeedParsesRSSAndAtomLinks(t *testing.T) {
	tests := []struct {
		name string
		body string
		url  string
	}{
		{"rss", `<rss><channel><item><title>RSS item</title><link>https://example.com/rss</link><pubDate>Mon, 02 Jan 2006 15:04:05 -0700</pubDate></item></channel></rss>`, "https://example.com/rss"},
		{"atom", `<feed><entry><title>Atom item</title><link href="https://example.com/atom"/><published>2006-01-02T15:04:05Z</published></entry></feed>`, "https://example.com/atom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var feed rssFeed
			if err := xml.Unmarshal([]byte(tt.body), &feed); err != nil {
				t.Fatal(err)
			}
			entries := append(feed.Channel.Items, feed.Entries...)
			if got := entries[0].Link.Href + entries[0].Link.Value; got != tt.url {
				t.Errorf("parsed URL %q, want %q", got, tt.url)
			}
		})
	}
}

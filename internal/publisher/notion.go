package publisher

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"content-automation-pipeline/pkg/config"
	"content-automation-pipeline/pkg/logger"
	"github.com/jomei/notionapi"
	"go.uber.org/zap"
)

type PublishRequest struct {
	Title       string
	Category    string
	SourceURL   string
	PostText    string
	Hashtags    string
	ImagePrompt string
	ContentKey  string
}

type Publisher interface {
	Publish(ctx context.Context, req PublishRequest) error
	HasPublished(ctx context.Context, contentKey string) (bool, error)
}

type NotionPublisher struct {
	client     *notionapi.Client
	databaseID notionapi.DatabaseID
}

func NewNotionPublisher(cfg *config.Config) Publisher {
	if cfg.NotionToken == "" || cfg.NotionDatabaseID == "" {
		logger.Log.Warn("Notion API credentials missing, using MockPublisher")
		return &MockPublisher{}
	}
	return &NotionPublisher{
		client:     notionapi.NewClient(notionapi.Token(cfg.NotionToken)),
		databaseID: notionapi.DatabaseID(cfg.NotionDatabaseID),
	}
}

func (n *NotionPublisher) Publish(ctx context.Context, req PublishRequest) error {
	now := notionapi.Date(time.Now())
	properties := notionapi.Properties{
		"Title": notionapi.TitleProperty{
			Title: []notionapi.RichText{
				{Text: &notionapi.Text{Content: req.Title}},
			},
		},
		"Category": notionapi.SelectProperty{
			Select: notionapi.Option{Name: req.Category},
		},
		"Source": notionapi.URLProperty{
			URL: req.SourceURL,
		},
		"Generated Post": notionapi.RichTextProperty{
			RichText: []notionapi.RichText{
				{Text: &notionapi.Text{Content: req.PostText}},
			},
		},
		"Hashtags": notionapi.RichTextProperty{
			RichText: []notionapi.RichText{
				{Text: &notionapi.Text{Content: req.Hashtags}},
			},
		},
		"Image Prompt": notionapi.RichTextProperty{
			RichText: []notionapi.RichText{
				{Text: &notionapi.Text{Content: req.ImagePrompt}},
			},
		},
		"Content Key": notionapi.RichTextProperty{
			RichText: []notionapi.RichText{
				{Text: &notionapi.Text{Content: req.ContentKey}},
			},
		},
		"Status": notionapi.SelectProperty{
			Select: notionapi.Option{Name: "Ready"},
		},
		"Date": notionapi.DateProperty{
			Date: &notionapi.DateObject{
				Start: &now,
			},
		},
	}

	pageReq := &notionapi.PageCreateRequest{
		Parent: notionapi.Parent{
			DatabaseID: n.databaseID,
		},
		Properties: properties,
	}

	res, err := n.client.Page.Create(ctx, pageReq)
	if err != nil {
		return fmt.Errorf("failed to create Notion page: %w", err)
	}

	logger.Log.Info("Successfully published to Notion", zap.String("pageID", res.ID.String()))
	return nil
}

// HasPublished uses the Content Key rich-text property as Notion-backed durable
// idempotency state. The property must be added to the target Notion database.
func (n *NotionPublisher) HasPublished(ctx context.Context, contentKey string) (bool, error) {
	result, err := n.client.Database.Query(ctx, n.databaseID, &notionapi.DatabaseQueryRequest{
		Filter: notionapi.PropertyFilter{
			Property: "Content Key",
			RichText: &notionapi.TextFilterCondition{Equals: contentKey},
		},
		PageSize: 1,
	})
	if err != nil {
		return false, fmt.Errorf("query existing Notion content key: %w", err)
	}
	return len(result.Results) > 0, nil
}

// ContentKey is stable across scheduler runs and changes if the URL or source
// title changes. It deliberately avoids hashing generated copy, which would be
// non-deterministic and therefore unsuitable for idempotency.
func ContentKey(title, sourceURL string) string {
	sum := sha256.Sum256([]byte(title + "\n" + sourceURL))
	return hex.EncodeToString(sum[:])
}

type MockPublisher struct{}

func (m *MockPublisher) Publish(ctx context.Context, req PublishRequest) error {
	logger.Log.Info("MOCK PUBLISH to Notion", zap.String("title", req.Title))
	return nil
}

func (m *MockPublisher) HasPublished(ctx context.Context, contentKey string) (bool, error) {
	return false, nil
}

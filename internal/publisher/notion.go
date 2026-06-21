package publisher

import (
	"context"
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
}

type Publisher interface {
	Publish(ctx context.Context, req PublishRequest) error
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

type MockPublisher struct{}

func (m *MockPublisher) Publish(ctx context.Context, req PublishRequest) error {
	logger.Log.Info("MOCK PUBLISH to Notion", zap.String("title", req.Title))
	return nil
}

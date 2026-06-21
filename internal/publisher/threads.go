package publisher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"content-automation-pipeline/pkg/config"
	"content-automation-pipeline/pkg/logger"
	"go.uber.org/zap"
)

type Publisher interface {
	Publish(ctx context.Context, content string) error
}

type ThreadsPublisher struct {
	userID      string
	accessToken string
	client      *http.Client
}

func NewThreadsPublisher(cfg *config.Config) Publisher {
	if cfg.ThreadsUserID == "" || cfg.ThreadsAccessToken == "" {
		logger.Log.Warn("Threads API credentials missing, using MockPublisher")
		return &MockPublisher{}
	}
	return &ThreadsPublisher{
		userID:      cfg.ThreadsUserID,
		accessToken: cfg.ThreadsAccessToken,
		client:      &http.Client{Timeout: 10 * time.Second},
	}
}

// Publish creates a Threads container and then publishes it
func (t *ThreadsPublisher) Publish(ctx context.Context, content string) error {
	// Step 1: Create a container
	containerID, err := t.createContainer(ctx, content)
	if err != nil {
		return fmt.Errorf("failed to create threads container: %w", err)
	}

	// Step 2: Publish the container
	err = t.publishContainer(ctx, containerID)
	if err != nil {
		return fmt.Errorf("failed to publish threads container: %w", err)
	}

	logger.Log.Info("Successfully published to Threads", zap.String("containerID", containerID))
	return nil
}

func (t *ThreadsPublisher) createContainer(ctx context.Context, text string) (string, error) {
	url := fmt.Sprintf("https://graph.threads.net/v1.0/%s/threads", t.userID)

	payload := map[string]string{
		"media_type":   "TEXT",
		"text":         text,
		"access_token": t.accessToken,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("API error creating container: %s", resp.Status)
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.ID, nil
}

func (t *ThreadsPublisher) publishContainer(ctx context.Context, containerID string) error {
	url := fmt.Sprintf("https://graph.threads.net/v1.0/%s/threads_publish", t.userID)

	payload := map[string]string{
		"creation_id":  containerID,
		"access_token": t.accessToken,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error publishing container: %s", resp.Status)
	}

	return nil
}

// --- Mock Publisher ---

type MockPublisher struct{}

func (m *MockPublisher) Publish(ctx context.Context, content string) error {
	logger.Log.Info("MOCK PUBLISH:", zap.String("content", content))
	return nil
}

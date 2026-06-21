package generator

import (
	"context"
	"fmt"
)

// --- OpenAI Generator ---

type OpenAIGenerator struct {
	apiKey string
}

func (o *OpenAIGenerator) RewriteArticle(ctx context.Context, title, url, summary string) (*GeneratedContent, error) {
	// Stub implementation
	return &GeneratedContent{
		PostText:    fmt.Sprintf("🚀 OpenAI Rewrite: %s\n\nRead more: %s", title, url),
		Hashtags:    "#backend #programming",
		ImagePrompt: "A futuristic server room with glowing blue lights",
	}, nil
}

// --- Claude Generator ---

type ClaudeGenerator struct {
	apiKey string
}

func (c *ClaudeGenerator) RewriteArticle(ctx context.Context, title, url, summary string) (*GeneratedContent, error) {
	// Stub implementation
	return &GeneratedContent{
		PostText:    fmt.Sprintf("🤖 Claude Rewrite: %s\n\nRead more: %s", title, url),
		Hashtags:    "#tech #news",
		ImagePrompt: "A robot typing on a computer",
	}, nil
}

// --- Gemini Generator ---

type GeminiGenerator struct {
	apiKey string
}

func (g *GeminiGenerator) RewriteArticle(ctx context.Context, title, url, summary string) (*GeneratedContent, error) {
	// Stub implementation
	return &GeneratedContent{
		PostText:    fmt.Sprintf("✨ Gemini Rewrite: %s\n\nRead more: %s", title, url),
		Hashtags:    "#ai #development",
		ImagePrompt: "A glowing AI brain connecting to the cloud",
	}, nil
}

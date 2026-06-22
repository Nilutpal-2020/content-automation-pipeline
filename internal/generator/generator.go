package generator

import (
	"context"
	"errors"

	"content-automation-pipeline/pkg/config"
	"content-automation-pipeline/pkg/logger"
	"go.uber.org/zap"
)

type GeneratedContent struct {
	PostText    string `json:"postText"`
	Hashtags    string `json:"hashtags"`
	ImagePrompt string `json:"imagePrompt"`
}

type Generator interface {
	RewriteArticle(ctx context.Context, category, title, url, summary string) (*GeneratedContent, error)
}

func NewGenerator(cfg *config.Config) (Generator, error) {
	if cfg.OpenAIKey != "" {
		logger.Log.Info("Using OpenAI as primary generator", zap.String("provider", "openai"))
		return &OpenAIGenerator{apiKey: cfg.OpenAIKey}, nil
	}
	if cfg.ClaudeKey != "" {
		logger.Log.Info("Using Claude as primary generator", zap.String("provider", "claude"))
		return &ClaudeGenerator{apiKey: cfg.ClaudeKey}, nil
	}
	if cfg.GeminiKey != "" {
		logger.Log.Info("Using Gemini as primary generator", zap.String("provider", "gemini"))
		return &GeminiGenerator{apiKey: cfg.GeminiKey}, nil
	}
	return nil, errors.New("no LLM API keys provided in config")
}

const BasePrompt = `You are a senior software engineer.
Read this article context. Summarize it into a Threads post.

Requirements:
- Hook in first line
- Maximum 400 characters
- Easy language
- Actionable
- Include emoji
- Include CTA
- Include hashtags

Category: %s
Article Title: %s
URL: %s
Summary Context: %s

Return JSON only: {"postText":"...", "hashtags":"...", "imagePrompt":"..."}.`

package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	OpenAIKey        string
	OpenAIModel      string
	ClaudeKey        string
	GeminiKey        string
	NotionToken      string
	NotionDatabaseID string
	LogLevel         string
	Port             string
	CronSchedule     string
	PostsPerCategory int
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load() // Ignore error if .env doesn't exist

	cfg := &Config{
		OpenAIKey:        getEnv("OPENAI_API_KEY", ""),
		OpenAIModel:      getEnv("OPENAI_MODEL", "gpt-5.4-mini"),
		ClaudeKey:        getEnv("ANTHROPIC_API_KEY", ""),
		GeminiKey:        getEnv("GEMINI_API_KEY", ""),
		NotionToken:      getEnv("NOTION_TOKEN", ""),
		NotionDatabaseID: getEnv("NOTION_DATABASE_ID", ""),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
		Port:             getEnv("PORT", "8080"),
		CronSchedule:     getEnv("CRON_SCHEDULE", "0 8 * * *"),
		PostsPerCategory: getEnvAsInt("POSTS_PER_CATEGORY", 3),
	}

	if cfg.PostsPerCategory < 1 {
		return nil, fmt.Errorf("POSTS_PER_CATEGORY must be at least 1")
	}
	if strings.EqualFold(getEnv("ENV", "development"), "production") {
		if cfg.NotionToken == "" || cfg.NotionDatabaseID == "" {
			return nil, fmt.Errorf("NOTION_TOKEN and NOTION_DATABASE_ID are required in production")
		}
		if cfg.OpenAIKey == "" && cfg.ClaudeKey == "" && cfg.GeminiKey == "" {
			return nil, fmt.Errorf("at least one LLM API key is required in production")
		}
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvAsInt(name string, fallback int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return fallback
}

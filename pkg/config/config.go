package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI           string
	MongoDBName        string
	RedisAddr          string
	RedisPassword      string
	OpenAIKey          string
	ClaudeKey          string
	GeminiKey          string
	ThreadsUserID      string
	ThreadsAccessToken string
	LogLevel           string
	Port               string
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load() // Ignore error if .env doesn't exist

	return &Config{
		MongoURI:           getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName:        getEnv("MONGO_DB_NAME", "content_pipeline"),
		RedisAddr:          getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:      getEnv("REDIS_PASSWORD", ""),
		OpenAIKey:          getEnv("OPENAI_API_KEY", ""),
		ClaudeKey:          getEnv("ANTHROPIC_API_KEY", ""),
		GeminiKey:          getEnv("GEMINI_API_KEY", ""),
		ThreadsUserID:      getEnv("THREADS_USER_ID", ""),
		ThreadsAccessToken: getEnv("THREADS_ACCESS_TOKEN", ""),
		LogLevel:           getEnv("LOG_LEVEL", "info"),
		Port:               getEnv("PORT", "8080"),
	}, nil
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

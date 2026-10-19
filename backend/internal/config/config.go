package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port         string
	CORSOrigins  []string
	RedisURL     string
	ReviewTTL    time.Duration
	LLMProvider  string
	LLMAPIKey    string
	LLMModel     string
	LLMURL       string
	LLMMaxTokens int
}

func Load() (Config, error) {
	origins := parseList(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if len(origins) == 0 {
		origins = []string{"http://localhost:5173"}
	}

	reviewTTL, err := time.ParseDuration(getEnvOrDefault("REVIEW_TTL", "1h"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid REVIEW_TTL: %w", err)
	}
	if reviewTTL <= 0 {
		return Config{}, fmt.Errorf("REVIEW_TTL must be greater than 0")
	}

	llmMaxTokens, err := strconv.Atoi(getEnvOrDefault("LLM_MAX_TOKENS", "2048"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid LLM_MAX_TOKENS: %w", err)
	}
	if llmMaxTokens <= 0 {
		return Config{}, fmt.Errorf("LLM_MAX_TOKENS must be greater than 0")
	}

	return Config{
		Port:         getEnvOrDefault("PORT", "8080"),
		CORSOrigins:  origins,
		RedisURL:     os.Getenv("REDIS_URL"),
		ReviewTTL:    reviewTTL,
		LLMProvider:  os.Getenv("LLM_PROVIDER"),
		LLMAPIKey:    os.Getenv("LLM_API_KEY"),
		LLMModel:     os.Getenv("LLM_MODEL"),
		LLMURL:       os.Getenv("LLM_URL"),
		LLMMaxTokens: llmMaxTokens,
	}, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func parseList(s string) []string {
	var result []string
	for _, val := range strings.Split(s, ",") {
		if trimmed := strings.TrimSpace(val); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

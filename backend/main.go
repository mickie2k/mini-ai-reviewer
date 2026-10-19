package main

import (
	"log"
	"mini-ai-reviewer/internal/cache"
	"mini-ai-reviewer/internal/config"
	"mini-ai-reviewer/internal/health"
	"mini-ai-reviewer/internal/llm"
	"mini-ai-reviewer/internal/review"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	var cacheClient cache.Client

	if cfg.RedisURL != "" {
		redisClient, err := cache.NewRedisClient(cfg.RedisURL)
		if err != nil {
			log.Fatalf("Failed to create Redis client: %v", err)
		}
		cacheClient = redisClient
	} else {
		log.Println("REDIS_URL not set; running without cache")
		cacheClient = cache.NewNoopClient()
	}

	var llmClient llm.Client

	switch cfg.LLMProvider {
	case "openai":
		llmClient = llm.NewOpenAIClient(cfg.LLMURL, cfg.LLMAPIKey)
	case "anthropic":
		llmClient = llm.NewAnthropicClient(cfg.LLMURL, cfg.LLMAPIKey)
	default:
		log.Fatal("Invalid LLM_PROVIDER. Must be 'openai' or 'anthropic'.")
	}

	reviewHandler := review.NewReview(llmClient, cfg.LLMModel, cacheClient, cfg.ReviewTTL, cfg.LLMMaxTokens)

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORSOrigins,
		AllowMethods: []string{"GET", "POST"},
		AllowHeaders: []string{"Content-Type"},
		MaxAge:       12 * time.Hour,
	}))

	router.GET("/health", health.Handler)
	router.POST("/lint", reviewHandler.Review)

	router.Run(":" + cfg.Port)
}

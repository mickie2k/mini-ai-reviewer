package review

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mini-ai-reviewer/internal/cache"
	"mini-ai-reviewer/internal/llm"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ReviewHandler struct {
	llm              llm.Client
	model            string
	maxTokens        int
	cache            cache.Client
	reviewExpiration time.Duration
}

func NewReview(llmClient llm.Client, model string, cacheClient cache.Client, reviewExpiration time.Duration, maxTokens int) *ReviewHandler {
	return &ReviewHandler{
		llm:              llmClient,
		model:            model,
		maxTokens:        maxTokens,
		cache:            cacheClient,
		reviewExpiration: reviewExpiration,
	}
}

func (r *ReviewHandler) Review(c *gin.Context) {
	var request ReviewRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if strings.TrimSpace(request.Code) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code must not be blank"})
		return
	}

	request.Code = normalizeCode(request.Code)

	cachedKey := r.cacheKey(request)
	if cached, ok := r.getCachedResponse(c.Request.Context(), cachedKey); ok {
		log.Printf("Cache hit for key: %s", cachedKey)
		c.JSON(http.StatusOK, cached)
		return
	}

	prompt, lineCount := buildPrompt(request.Language, request.Code)

	llmRequest := llm.Request{
		System:    systemPrompt,
		Message:   prompt,
		MaxTokens: r.maxTokens,
	}

	resp, err := r.llm.Send(c.Request.Context(), llmRequest, r.model)
	if err != nil {
		log.Printf("Error sending request to LLM: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to get response from LLM"})
		return
	}

	result, err := parseReviewResponse(resp.Message, lineCount)
	if err != nil {
		log.Printf("Error parsing LLM response: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to parse LLM response"})
		return
	}

	err = r.addCacheResponse(cachedKey, result)
	if err != nil {
		log.Printf("Error adding cache response: %v", err)
	}

	c.JSON(http.StatusOK, result)
}

func (r *ReviewHandler) cacheKey(request ReviewRequest) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%s|%s", systemPrompt, r.model, request.Language, request.Code)))
	return fmt.Sprintf("review:%s", hex.EncodeToString(sum[:]))
}

func (r *ReviewHandler) getCachedResponse(ctx context.Context, key string) (ReviewResponse, bool) {
	data, err := r.cache.Get(ctx, key)
	if err != nil {
		if !errors.Is(err, cache.ErrMiss) {
			log.Printf("cache error for key %s: %v", key, err)
		}
		return ReviewResponse{}, false
	}

	var response ReviewResponse
	if err := json.Unmarshal([]byte(data), &response); err != nil {
		log.Printf("cache json unmarshal error for key %s: %v", key, err)
		return ReviewResponse{}, false
	}

	return response, true
}

func (r *ReviewHandler) addCacheResponse(key string, response ReviewResponse) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := json.Marshal(response)
	if err != nil {
		return err
	}

	return r.cache.Set(ctx, key, string(data), r.reviewExpiration)
}

func normalizeCode(code string) string {
	return strings.ReplaceAll(code, "\r\n", "\n")
}

func extractReviewJSON(response string) (ReviewResponse, error) {

	// find from last to first, avoid reasoning content from LLM that may in result and contain JSON but not the actual review response
	for i := strings.LastIndex(response, "{"); i >= 0; i = strings.LastIndex(response[:i], "{") {
		var result ReviewResponse
		if err := json.NewDecoder(strings.NewReader(response[i:])).Decode(&result); err == nil && result.Comments != nil {
			return result, nil
		}
	}
	return ReviewResponse{}, errors.New("no JSON object with \"comments\" found")
}

func parseReviewResponse(response string, lineCount int) (ReviewResponse, error) {
	result, err := extractReviewJSON(response)
	if err != nil {
		return ReviewResponse{}, fmt.Errorf("invalid JSON from LLM: %w", err)
	}

	validComments := make([]Comment, 0, len(result.Comments))
	for _, comment := range result.Comments {
		if comment.StartLine < 1 || comment.EndLine < comment.StartLine || comment.EndLine > lineCount || strings.TrimSpace(comment.Explanation) == "" {
			continue
		}
		validComments = append(validComments, comment)
	}
	return ReviewResponse{Comments: validComments}, nil
}

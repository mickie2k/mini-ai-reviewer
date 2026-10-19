package review

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"mini-ai-reviewer/internal/cache"
	"mini-ai-reviewer/internal/llm"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testModel  = "test-model"
	testTTL    = time.Minute
	testTokens = 1024
	testCode   = "line one\nline two\nline three"
)

func newTestHandler(t *testing.T) (*ReviewHandler, *llm.MockClient, *cache.MockClient) {
	t.Helper()
	llmMock := llm.NewMockClient(t)
	cacheMock := cache.NewMockClient(t)
	return NewReview(llmMock, testModel, cacheMock, testTTL, testTokens), llmMock, cacheMock
}

func sendReview(t *testing.T, h *ReviewHandler, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/review", h.Review)

	req := httptest.NewRequest(http.MethodPost, "/review", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func reviewBody(t *testing.T, language Language, code string) string {
	t.Helper()
	b, err := json.Marshal(ReviewRequest{Language: language, Code: code})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	return string(b)
}

func decodeReview(t *testing.T, w *httptest.ResponseRecorder) ReviewResponse {
	t.Helper()
	var resp ReviewResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v (body: %s)", err, w.Body.String())
	}
	return resp
}

func TestReview_BadRequest(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"malformed json", `{not json`},
		{"empty body", ``},
		{"missing code", `{"language":"golang"}`},
		{"missing language", `{"code":"x := 1"}`},
		{"unsupported language", `{"code":"x := 1","language":"rust"}`},
		{"whitespace-only code", `{"code":"  \n\t ","language":"golang"}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h, _, _ := newTestHandler(t)

			w := sendReview(t, h, tc.body)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Contains(t, w.Body.String(), "error")
		})
	}
}

func TestReview_CacheHit(t *testing.T) {
	h, _, cacheMock := newTestHandler(t)

	cached := ReviewResponse{Comments: []Comment{{StartLine: 1, EndLine: 2, Explanation: "cached", Fix: "fix"}}}
	data, err := json.Marshal(cached)
	if err != nil {
		t.Fatalf("marshal cached response: %v", err)
	}

	cacheMock.EXPECT().Get(mock.Anything, mock.AnythingOfType("string")).Return(string(data), nil).Once()

	w := sendReview(t, h, reviewBody(t, Go, testCode))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, cached, decodeReview(t, w))
}

func TestReview_CacheMiss_Success(t *testing.T) {
	h, llmMock, cacheMock := newTestHandler(t)

	wantKey := h.cacheKey(ReviewRequest{Language: Python, Code: testCode})
	cacheMock.EXPECT().Get(mock.Anything, wantKey).Return("", cache.ErrMiss).Once()

	llmReply := `{"comments":[
  {"startLine":1,"endLine":2,"explanation":"valid","fix":"ok"},
  {"startLine":0,"endLine":1,"explanation":"start below 1","fix":""},
  {"startLine":3,"endLine":2,"explanation":"end before start","fix":""},
  {"startLine":2,"endLine":99,"explanation":"beyond last line","fix":""},
  {"startLine":1,"endLine":1,"explanation":"","fix":""}
]}`

	llmMock.EXPECT().
		Send(mock.Anything, mock.Anything, testModel).
		RunAndReturn(func(_ context.Context, r llm.Request, _ string) (llm.Response, error) {
			assert.Equal(t, systemPrompt, r.System)
			assert.Equal(t, testTokens, r.MaxTokens)
			assert.Contains(t, r.Message, "1: line one")
			assert.Contains(t, r.Message, "3: line three")
			return llm.Response{Message: llmReply}, nil
		}).
		Once()

	want := ReviewResponse{Comments: []Comment{{StartLine: 1, EndLine: 2, Explanation: "valid", Fix: "ok"}}}
	wantJSON, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal want: %v", err)
	}
	cacheMock.EXPECT().Set(mock.Anything, wantKey, string(wantJSON), testTTL).Return(nil).Once()

	w := sendReview(t, h, reviewBody(t, Python, testCode))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, want, decodeReview(t, w))
}

func TestReview_Normalizes(t *testing.T) {
	h, llmMock, cacheMock := newTestHandler(t)

	wantKey := h.cacheKey(ReviewRequest{Language: Go, Code: "a\nb"})
	cacheMock.EXPECT().Get(mock.Anything, wantKey).Return("", cache.ErrMiss).Once()
	llmMock.EXPECT().Send(mock.Anything, mock.Anything, testModel).
		Return(llm.Response{Message: `{"comments":[]}`}, nil).Once()
	cacheMock.EXPECT().Set(mock.Anything, wantKey, mock.Anything, testTTL).Return(nil).Once()

	w := sendReview(t, h, reviewBody(t, Go, "a\r\nb"))

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReview_LLMError(t *testing.T) {
	h, llmMock, cacheMock := newTestHandler(t)

	cacheMock.EXPECT().Get(mock.Anything, mock.Anything).Return("", cache.ErrMiss).Once()
	llmMock.EXPECT().Send(mock.Anything, mock.Anything, testModel).
		Return(llm.Response{}, errors.New("upstream down")).Once()

	w := sendReview(t, h, reviewBody(t, Go, testCode))

	assert.Equal(t, http.StatusBadGateway, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to get response from LLM")
}

func TestReview_LLMInvalidJSON(t *testing.T) {
	h, llmMock, cacheMock := newTestHandler(t)

	cacheMock.EXPECT().Get(mock.Anything, mock.Anything).Return("", cache.ErrMiss).Once()
	llmMock.EXPECT().Send(mock.Anything, mock.Anything, testModel).
		Return(llm.Response{Message: "sorry, I cannot help"}, nil).Once()

	w := sendReview(t, h, reviewBody(t, Go, testCode))

	assert.Equal(t, http.StatusBadGateway, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to parse LLM response")
}

func TestReview_CacheGetError_Success(t *testing.T) {
	h, llmMock, cacheMock := newTestHandler(t)

	cacheMock.EXPECT().Get(mock.Anything, mock.Anything).Return("", errors.New("redis down")).Once()
	llmMock.EXPECT().Send(mock.Anything, mock.Anything, testModel).
		Return(llm.Response{Message: `{"comments":[]}`}, nil).Once()
	cacheMock.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything, testTTL).Return(nil).Once()

	w := sendReview(t, h, reviewBody(t, Go, testCode))

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReview_CorruptCacheEntry_Success(t *testing.T) {
	h, llmMock, cacheMock := newTestHandler(t)

	cacheMock.EXPECT().Get(mock.Anything, mock.Anything).Return("not json", nil).Once()
	llmMock.EXPECT().Send(mock.Anything, mock.Anything, testModel).
		Return(llm.Response{Message: `{"comments":[]}`}, nil).Once()
	cacheMock.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything, testTTL).Return(nil).Once()

	w := sendReview(t, h, reviewBody(t, Go, testCode))

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReview_CacheSetError_Success(t *testing.T) {
	h, llmMock, cacheMock := newTestHandler(t)

	cacheMock.EXPECT().Get(mock.Anything, mock.Anything).Return("", cache.ErrMiss).Once()
	llmMock.EXPECT().Send(mock.Anything, mock.Anything, testModel).
		Return(llm.Response{Message: `{"comments":[{"startLine":1,"endLine":1,"explanation":"bug","fix":"f"}]}`}, nil).Once()
	cacheMock.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything, testTTL).Return(errors.New("write failed")).Once()

	w := sendReview(t, h, reviewBody(t, Go, testCode))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Len(t, decodeReview(t, w).Comments, 1)
}

func TestCacheKey(t *testing.T) {
	h, _, _ := newTestHandler(t)
	base := ReviewRequest{Language: Go, Code: "x"}

	assert.Equal(t, h.cacheKey(base), h.cacheKey(base), "same input → same key")
	assert.True(t, strings.HasPrefix(h.cacheKey(base), "review:"))
	assert.NotEqual(t, h.cacheKey(base), h.cacheKey(ReviewRequest{Language: Python, Code: "x"}), "language changes key")
	assert.NotEqual(t, h.cacheKey(base), h.cacheKey(ReviewRequest{Language: Go, Code: "y"}), "code changes key")

	other := NewReview(nil, "other-model", nil, testTTL, testTokens)
	assert.NotEqual(t, h.cacheKey(base), other.cacheKey(base), "model changes key")
}

func TestParseReviewResponse(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		lineCount int
		want      []Comment
		wantErr   bool
	}{
		{"empty comments", `{"comments":[]}`, 3, []Comment{}, false},
		{"strips markdown fences", "```json\n{\"comments\":[{\"startLine\":1,\"endLine\":1,\"explanation\":\"e\",\"fix\":\"\"}]}\n```", 3,
			[]Comment{{StartLine: 1, EndLine: 1, Explanation: "e"}}, false},
		{"drops out-of-range", `{"comments":[{"startLine":1,"endLine":4,"explanation":"e","fix":""}]}`, 3, []Comment{}, false},
		{"reasoning with braces before json", "Line 5: `func SumInts(slice []int) int {` looks correct.\nNo errors.{\"comments\": []}", 3, []Comment{}, false},
		{"reasoning quotes example before answer", "Format is {\"comments\": []} so:\n{\"comments\":[{\"startLine\":2,\"endLine\":2,\"explanation\":\"e\",\"fix\":\"f\"}]}", 3,
			[]Comment{{StartLine: 2, EndLine: 2, Explanation: "e", Fix: "f"}}, false},
		{"no json at all", `nope`, 3, nil, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseReviewResponse(tc.input, tc.lineCount)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			assert.Equal(t, tc.want, got.Comments)
		})
	}
}

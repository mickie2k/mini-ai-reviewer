package llm

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type openAIClient struct {
	client openai.Client
}

func NewOpenAIClient(baseURL string, apiKey string) Client {
	opts := []option.RequestOption{
		option.WithAPIKey(apiKey),
	}

	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}

	client := openai.NewClient(
		opts...,
	)

	return &openAIClient{
		client: client,
	}
}

func (c *openAIClient) Send(ctx context.Context, request Request, model string) (Response, error) {

	params := openai.ChatCompletionNewParams{
		Model: model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(request.System),
			openai.UserMessage(request.Message),
		},
		MaxCompletionTokens: openai.Int(int64(request.MaxTokens)),
	}

	completion, err := c.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return Response{}, fmt.Errorf("openai request failed: %w", err)
	}

	if len(completion.Choices) == 0 {
		return Response{}, fmt.Errorf("openai request returned no content")
	}

	return Response{Message: completion.Choices[0].Message.Content}, nil
}

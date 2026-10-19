package llm

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type anthropicClient struct {
	client anthropic.Client
}

func NewAnthropicClient(baseURL string, apiKey string) Client {
	opts := []option.RequestOption{
		option.WithAPIKey(apiKey),
	}

	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}

	client := anthropic.NewClient(
		opts...,
	)

	return &anthropicClient{
		client: client,
	}
}

func (c *anthropicClient) Send(ctx context.Context, request Request, model string) (Response, error) {

	params := anthropic.MessageNewParams{
		MaxTokens: int64(request.MaxTokens),
		System: []anthropic.TextBlockParam{
			{Text: request.System},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(request.Message)),
		},
		Model: model,
	}

	message, err := c.client.Messages.New(ctx, params)
	if err != nil {
		return Response{}, fmt.Errorf("anthropic request failed: %w", err)
	}

	// Only extract type:text from the message
	var text string
	for _, block := range message.Content {
		if textBlock, ok := block.AsAny().(anthropic.TextBlock); ok {
			text += textBlock.Text
		}
	}

	if len(text) == 0 {
		return Response{}, fmt.Errorf("anthropic request returned no content")
	}

	return Response{Message: text}, nil
}

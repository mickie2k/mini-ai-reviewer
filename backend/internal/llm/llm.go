package llm

import (
	"context"
)

type Request struct {
	System    string
	Message   string
	MaxTokens int
}

type Response struct {
	Message string
}

type Client interface {
	Send(ctx context.Context, request Request, model string) (Response, error)
}

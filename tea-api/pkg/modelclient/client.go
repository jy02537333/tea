package modelclient

import "context"

// Client is a minimal interface for text generation model adapters.
type Client interface {
	// Generate sends a prompt and returns the generated text or an error.
	Generate(ctx context.Context, prompt string) (string, error)
}

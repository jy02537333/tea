package modelclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"tea-test/pkg/env"
	"time"
)

// AnthropicClient is a minimal adapter for calling an Anthropic-style HTTP API.
// It is intentionally conservative: it reads configuration from environment
// variables and tries to extract a text completion from common response shapes.
type AnthropicClient struct {
	apiKey     string
	apiURL     string
	model      string
	timeout    time.Duration
	httpClient *http.Client
}

// NewAnthropicClient constructs an AnthropicClient with explicit parameters.
func NewAnthropicClient(apiKey, apiURL, model string, timeout time.Duration) (*AnthropicClient, error) {
	if apiKey == "" {
		return nil, errors.New("apiKey is empty")
	}
	if apiURL == "" {
		apiURL = "https://api.anthropic.com/v1/complete"
	}
	if model == "" {
		model = "claude-sonnet-4.5"
	}
	if timeout <= 0 {
		timeout = 20 * time.Second
	}

	return &AnthropicClient{
		apiKey:     apiKey,
		apiURL:     apiURL,
		model:      model,
		timeout:    timeout,
		httpClient: &http.Client{Timeout: timeout},
	}, nil
}

// NewAnthropicClientFromEnv constructs a client using environment variables:
// MODEL_API_KEY (required), MODEL_API_URL (optional), MODEL_NAME (optional), MODEL_TIMEOUT
func NewAnthropicClientFromEnv() (*AnthropicClient, error) {
	apiKey := env.Get("MODEL_API_KEY", "")
	if apiKey == "" {
		return nil, errors.New("MODEL_API_KEY is not set in environment")
	}
	apiURL := env.Get("MODEL_API_URL", "")
	model := env.Get("MODEL_NAME", "")
	to := 20 * time.Second
	if t := env.Get("MODEL_TIMEOUT", ""); t != "" {
		if parsed, err := time.ParseDuration(t); err == nil {
			to = parsed
		}
	}

	return NewAnthropicClient(apiKey, apiURL, model, to)
}

// Generate sends a prompt and attempts to extract a completion string from the response.
func (c *AnthropicClient) Generate(ctx context.Context, prompt string) (string, error) {
	// Build a generic request body that many Anthropic-like endpoints accept.
	body := map[string]interface{}{
		"model":       c.model,
		"prompt":      prompt,
		"max_tokens":  512,
		"temperature": 0.2,
	}
	b, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.apiURL, bytes.NewReader(b))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("model endpoint returned status %d", resp.StatusCode)
	}

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	// Flexible extraction: try common fields
	if v, ok := raw["completion"]; ok {
		if s, ok := v.(string); ok {
			return s, nil
		}
	}
	// Choices style (openai-like)
	if choices, ok := raw["choices"]; ok {
		if arr, ok := choices.([]interface{}); ok && len(arr) > 0 {
			if first, ok := arr[0].(map[string]interface{}); ok {
				if text, ok := first["text"].(string); ok {
					return text, nil
				}
				// message.content style
				if msg, ok := first["message"].(map[string]interface{}); ok {
					if content, ok := msg["content"].(string); ok {
						return content, nil
					}
				}
			}
		}
	}

	// outputs / data first string
	if outputs, ok := raw["outputs"]; ok {
		if arr, ok := outputs.([]interface{}); ok && len(arr) > 0 {
			if first, ok := arr[0].(map[string]interface{}); ok {
				// try text / content / completion
				for _, k := range []string{"text", "content", "completion"} {
					if v, ok := first[k]; ok {
						if s, ok := v.(string); ok {
							return s, nil
						}
					}
				}
			}
		}
	}

	// Last resort: find any string value in the top-level map
	for _, v := range raw {
		if s, ok := v.(string); ok && len(s) > 0 {
			return s, nil
		}
	}

	// Nothing found
	enc, _ := json.MarshalIndent(raw, "", "  ")
	return "", fmt.Errorf("unable to extract completion from response: %s", string(enc))
}

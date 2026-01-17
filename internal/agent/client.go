package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultBaseURL    = "https://api.anthropic.com"
	anthropicVersion  = "2023-06-01"
	defaultMaxTokens  = 4096
	maxRetries        = 3
	baseRetryDelay    = 1 * time.Second
)

// Client handles communication with the Anthropic API
type Client struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
}

// Message represents a conversation message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// messageRequest represents an API request to the messages endpoint
type messageRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system,omitempty"`
	Messages  []Message `json:"messages"`
}

// messageResponse represents an API response from the messages endpoint
type messageResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// errorResponse represents an API error response
type errorResponse struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// NewClient creates a new Anthropic API client
func NewClient(apiKey, model string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// SendMessage sends a message to Claude and returns the response text
// Implements retry with exponential backoff on rate limits (HTTP 429)
func (c *Client) SendMessage(ctx context.Context, system string, messages []Message) (string, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		response, err := c.doRequest(ctx, system, messages)
		if err == nil {
			return c.extractText(response), nil
		}

		lastErr = err

		// Check if it's a rate limit error (429)
		if isRateLimitError(err) && attempt < maxRetries {
			delay := baseRetryDelay * time.Duration(1<<attempt) // Exponential backoff
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(delay):
				continue
			}
		}

		// For non-rate-limit errors, don't retry
		if !isRateLimitError(err) {
			return "", err
		}
	}

	return "", fmt.Errorf("max retries exceeded: %w", lastErr)
}

// doRequest performs the actual HTTP request to the Anthropic API
func (c *Client) doRequest(ctx context.Context, system string, messages []Message) (*messageResponse, error) {
	reqBody := messageRequest{
		Model:     c.model,
		MaxTokens: defaultMaxTokens,
		System:    system,
		Messages:  messages,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/messages", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp errorResponse
		if err := json.Unmarshal(body, &errResp); err == nil {
			return nil, &APIError{
				StatusCode: resp.StatusCode,
				Type:       errResp.Error.Type,
				Message:    errResp.Error.Message,
			}
		}
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	var msgResp messageResponse
	if err := json.Unmarshal(body, &msgResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &msgResp, nil
}

// extractText extracts the text content from a message response
func (c *Client) extractText(resp *messageResponse) string {
	for _, content := range resp.Content {
		if content.Type == "text" {
			return content.Text
		}
	}
	return ""
}

// APIError represents an error from the Anthropic API
type APIError struct {
	StatusCode int
	Type       string
	Message    string
}

func (e *APIError) Error() string {
	if e.Type != "" {
		return fmt.Sprintf("API error %d (%s): %s", e.StatusCode, e.Type, e.Message)
	}
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}

// isRateLimitError checks if an error is a rate limit error (HTTP 429)
func isRateLimitError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == http.StatusTooManyRequests
	}
	return false
}

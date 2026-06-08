package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// LLMConfig holds configuration for an OpenAI-compatible LLM endpoint.
type LLMConfig struct {
	Endpoint  string // e.g. "https://api.openai.com/v1"
	APIKey    string
	Model     string // e.g. "gpt-4o-mini"
	MaxTokens int    // default 2048
}

// ChatMessage represents a single message in a chat conversation.
type ChatMessage struct {
	Role    string `json:"role"`    // system/user/assistant
	Content string `json:"content"`
}

// ChatRequest is the request body sent to the chat completions endpoint.
type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

// LLMClient is a general-purpose client for OpenAI-compatible LLM APIs.
type LLMClient struct {
	config LLMConfig
	client *http.Client
}

// NewLLMClient creates a new LLM client with the given configuration.
func NewLLMClient(config LLMConfig) *LLMClient {
	maxTokens := config.MaxTokens
	if maxTokens == 0 {
		maxTokens = 2048
	}
	return &LLMClient{
		config: LLMConfig{
			Endpoint:  strings.TrimRight(config.Endpoint, "/"),
			APIKey:    config.APIKey,
			Model:     config.Model,
			MaxTokens: maxTokens,
		},
		client: &http.Client{},
	}
}

// chatResponse is the non-streaming response structure from the API.
type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// streamDelta is the delta structure within a streaming chunk.
type streamDelta struct {
	Content string `json:"content"`
}

// streamChoice is a single choice within a streaming chunk.
type streamChoice struct {
	Delta streamDelta `json:"delta"`
}

// streamChunk is a single SSE chunk from a streaming response.
type streamChunk struct {
	Choices []streamChoice `json:"choices"`
}

// Chat sends a non-streaming chat completion request and returns the assistant message content.
func (c *LLMClient) Chat(messages []ChatMessage) (string, error) {
	if c.config.Endpoint == "" {
		return "", fmt.Errorf("LLM endpoint not configured")
	}

	reqBody, err := json.Marshal(ChatRequest{
		Model:     c.config.Model,
		Messages:  messages,
		MaxTokens: c.config.MaxTokens,
	})
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.config.Endpoint+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("llm request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("llm returned %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("llm returned no choices")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// ChatStream sends a streaming chat completion request. It returns a channel that
// emits content delta strings as they arrive. The channel is closed when the
// stream ends or when an error occurs.
func (c *LLMClient) ChatStream(messages []ChatMessage) (<-chan string, error) {
	if c.config.Endpoint == "" {
		return nil, fmt.Errorf("LLM endpoint not configured")
	}

	reqBody, err := json.Marshal(ChatRequest{
		Model:       c.config.Model,
		Messages:    messages,
		MaxTokens:   c.config.MaxTokens,
		Temperature: 0,
		Stream:      true,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.config.Endpoint+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("llm request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("llm returned %d: %s", resp.StatusCode, string(body))
	}

	ch := make(chan string, 64)

	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			// SSE lines start with "data: "
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")

			// Stream complete
			if data == "[DONE]" {
				return
			}

			var chunk streamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				// Skip malformed chunks
				continue
			}

			for _, choice := range chunk.Choices {
				if choice.Delta.Content != "" {
					ch <- choice.Delta.Content
				}
			}
		}
	}()

	return ch, nil
}

// CountTokens returns a rough token estimate (len(text) / 4).
func (c *LLMClient) CountTokens(text string) int {
	return len(text) / 4
}

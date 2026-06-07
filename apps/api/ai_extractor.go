package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type AIExtractor struct {
	baseURL    string
	model      string
	apiKey     string
	httpClient *http.Client
}

func NewAIExtractor(baseURL, model, apiKey string) *AIExtractor {
	return &AIExtractor{
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func NewAIExtractorFromEnv() *AIExtractor {
	return NewAIExtractor(
		os.Getenv("LIFEOPS_LLM_BASE_URL"),
		os.Getenv("LIFEOPS_LLM_MODEL"),
		os.Getenv("LIFEOPS_LLM_API_KEY"),
	)
}

func (e *AIExtractor) Extract(content string) (string, error) {
	if e.baseURL == "" {
		return "", fmt.Errorf("LIFEOPS_LLM_BASE_URL not configured")
	}

	messages := buildMessages(content)

	reqBody := map[string]interface{}{
		"model":       e.model,
		"messages":    messages,
		"temperature": 0,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, e.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if e.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+e.apiKey)
	}

	resp, err := e.httpClient.Do(req)
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

	var chatResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("llm returned no choices")
	}

	return chatResp.Choices[0].Message.Content, nil
}

func buildMessages(content string) []map[string]string {
	return []map[string]string{
		{
			"role": "system",
			"content": `You are a family information organizer. Your job is to convert messy family messages into structured drafts.

Rules:
- Output exactly one JSON object with a top-level "drafts" array.
- Each draft must have: draft_type, title, confidence.
- Allowed draft_type values: event, task, shopping_item, note.
- Use null for unknown values. Do not invent facts.
- Do not output Markdown. Do not explain the JSON.
- Preserve uncertainty. If a date is ambiguous, set due_at to null.`,
		},
		{
			"role": "user",
			"content": `周五孩子要带彩笔，别忘了交水费。`,
		},
		{
			"role": "assistant",
			"content": `{"drafts":[{"draft_type":"task","title":"准备彩笔","description":"周五孩子需要带彩笔","due_at":null,"assignee_hint":null,"confidence":0.9},{"draft_type":"task","title":"交水费","description":"提醒家庭处理水费缴纳","due_at":null,"assignee_hint":null,"confidence":0.86}]}`,
		},
		{
			"role": "user",
			"content": `下周二 18:30 家长会，我可能来不及，你看能不能去。`,
		},
		{
			"role": "assistant",
			"content": `{"drafts":[{"draft_type":"event","title":"家长会","description":"下周二 18:30 有家长会，需要确认谁参加","due_at":null,"assignee_hint":null,"confidence":0.82},{"draft_type":"task","title":"确认家长会参加人","description":"需要确认伴侣是否能参加家长会","due_at":null,"assignee_hint":"partner","confidence":0.78}]}`,
		},
		{
			"role": "user",
			"content": `周日买牛奶、鸡蛋、纸巾，晚上少油少辣。`,
		},
		{
			"role": "assistant",
			"content": `{"drafts":[{"draft_type":"shopping_item","title":"牛奶","description":null,"quantity":null,"confidence":0.95},{"draft_type":"shopping_item","title":"鸡蛋","description":null,"quantity":null,"confidence":0.95},{"draft_type":"shopping_item","title":"纸巾","description":null,"quantity":null,"confidence":0.95},{"draft_type":"note","title":"晚餐偏好","description":"晚上少油少辣","topic":"food","confidence":0.88}]}`,
		},
		{
			"role": "user",
			"content": content,
		},
	}
}

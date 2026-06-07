package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWebhookInboxAcceptsMessage(t *testing.T) {
	server := NewServer("dev-token")

	body := bytes.NewBufferString(`{
		"source": "wechat",
		"sender": "partner",
		"content": "周五孩子要带彩笔，别忘了交水费。"
	}`)
	request := httptest.NewRequest(http.MethodPost, "/api/inbox/webhook", body)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-LifeOps-Token", "dev-token")
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d with body %s", response.Code, response.Body.String())
	}

	var item InboxItem
	if err := json.Unmarshal(response.Body.Bytes(), &item); err != nil {
		t.Fatalf("response is not a valid inbox item: %v", err)
	}
	if item.ID == "" {
		t.Fatal("expected generated id")
	}
	if item.Source != "wechat" {
		t.Fatalf("expected source wechat, got %q", item.Source)
	}
	if item.Sender != "partner" {
		t.Fatalf("expected sender partner, got %q", item.Sender)
	}
	if item.Content != "周五孩子要带彩笔，别忘了交水费。" {
		t.Fatalf("unexpected content %q", item.Content)
	}
	if item.Status != "new" {
		t.Fatalf("expected status new, got %q", item.Status)
	}
}

func TestWebhookRejectsInvalidToken(t *testing.T) {
	server := NewServer("dev-token")

	body := bytes.NewBufferString(`{"source":"wechat","sender":"partner","content":"test"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/inbox/webhook", body)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-LifeOps-Token", "wrong-token")
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", response.Code)
	}
}

func TestWebhookRejectsMissingToken(t *testing.T) {
	server := NewServer("dev-token")

	body := bytes.NewBufferString(`{"source":"wechat","sender":"partner","content":"test"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/inbox/webhook", body)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", response.Code)
	}
}

func TestWebhookRejectsNonPost(t *testing.T) {
	server := NewServer("dev-token")

	request := httptest.NewRequest(http.MethodGet, "/api/inbox/webhook", nil)
	request.Header.Set("X-LifeOps-Token", "dev-token")
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", response.Code)
	}
}

func TestWebhookRejectsInvalidJSON(t *testing.T) {
	server := NewServer("dev-token")

	body := bytes.NewBufferString(`not json`)
	request := httptest.NewRequest(http.MethodPost, "/api/inbox/webhook", body)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-LifeOps-Token", "dev-token")
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}

	var errResp errorResponse
	if err := json.Unmarshal(response.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("response is not valid error json: %v", err)
	}
	if errResp.Error != "invalid_request" {
		t.Fatalf("expected error code invalid_request, got %q", errResp.Error)
	}
}

func TestWebhookRejectsMissingSource(t *testing.T) {
	server := NewServer("dev-token")

	body := bytes.NewBufferString(`{"sender":"partner","content":"test"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/inbox/webhook", body)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-LifeOps-Token", "dev-token")
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}

	var errResp errorResponse
	if err := json.Unmarshal(response.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("response is not valid error json: %v", err)
	}
	if errResp.Error != "invalid_request" {
		t.Fatalf("expected error code invalid_request, got %q", errResp.Error)
	}
}

func TestWebhookRejectsEmptyContent(t *testing.T) {
	server := NewServer("dev-token")

	body := bytes.NewBufferString(`{"source":"wechat","sender":"partner","content":"   "}`)
	request := httptest.NewRequest(http.MethodPost, "/api/inbox/webhook", body)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-LifeOps-Token", "dev-token")
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}

	var errResp errorResponse
	if err := json.Unmarshal(response.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("response is not valid error json: %v", err)
	}
	if errResp.Error != "empty_content" {
		t.Fatalf("expected error code empty_content, got %q", errResp.Error)
	}
}

func TestWebhookAcceptsMessageWithoutSender(t *testing.T) {
	server := NewServer("dev-token")

	body := bytes.NewBufferString(`{"source":"telegram","content":"buy milk"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/inbox/webhook", body)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-LifeOps-Token", "dev-token")
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d with body %s", response.Code, response.Body.String())
	}
}

func TestWebhookErrorResponseIsJSON(t *testing.T) {
	server := NewServer("dev-token")

	tests := []struct {
		name        string
		body        string
		wantCode    string
		wantMessage string
	}{
		{
			name:        "invalid json",
			body:        `{broken`,
			wantCode:    "invalid_request",
			wantMessage: "invalid json",
		},
		{
			name:        "missing source",
			body:        `{"content":"test"}`,
			wantCode:    "invalid_request",
			wantMessage: "source is required",
		},
		{
			name:        "empty content",
			body:        `{"source":"wechat","content":""}`,
			wantCode:    "empty_content",
			wantMessage: "content is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := bytes.NewBufferString(tt.body)
			request := httptest.NewRequest(http.MethodPost, "/api/inbox/webhook", body)
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("X-LifeOps-Token", "dev-token")
			response := httptest.NewRecorder()

			server.ServeHTTP(response, request)

			ct := response.Header().Get("Content-Type")
			if !strings.Contains(ct, "application/json") {
				t.Fatalf("expected content-type application/json, got %q", ct)
			}

			var errResp errorResponse
			if err := json.Unmarshal(response.Body.Bytes(), &errResp); err != nil {
				t.Fatalf("response is not valid error json: %v", err)
			}
			if errResp.Error != tt.wantCode {
				t.Fatalf("expected error code %q, got %q", tt.wantCode, errResp.Error)
			}
			if errResp.Message != tt.wantMessage {
				t.Fatalf("expected message %q, got %q", tt.wantMessage, errResp.Message)
			}
		})
	}
}

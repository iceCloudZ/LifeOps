package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

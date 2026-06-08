package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func testServer(t *testing.T) *Server {
	t.Helper()
	dir := t.TempDir()
	store, err := NewStore(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return NewServer("dev-token", store, nil)
}

func authRequest(method, url, body string) *http.Request {
	req := httptest.NewRequest(method, url, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer dev-token")
	return req
}

func TestWebhookInboxAcceptsMessage(t *testing.T) {
	server := testServer(t)

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
	if item.Source != "wechat" || item.Sender != "partner" || item.Status != "new" {
		t.Fatalf("unexpected item: %+v", item)
	}
}

func TestWebhookRejectsInvalidToken(t *testing.T) {
	server := testServer(t)

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
	server := testServer(t)

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
	server := testServer(t)
	request := httptest.NewRequest(http.MethodGet, "/api/inbox/webhook", nil)
	request.Header.Set("X-LifeOps-Token", "dev-token")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", response.Code)
	}
}

func TestWebhookRejectsInvalidJSON(t *testing.T) {
	server := testServer(t)
	body := bytes.NewBufferString(`not json`)
	request := httptest.NewRequest(http.MethodPost, "/api/inbox/webhook", body)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-LifeOps-Token", "dev-token")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}

func TestWebhookRejectsMissingSource(t *testing.T) {
	server := testServer(t)
	body := bytes.NewBufferString(`{"sender":"partner","content":"test"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/inbox/webhook", body)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-LifeOps-Token", "dev-token")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}

func TestWebhookRejectsEmptyContent(t *testing.T) {
	server := testServer(t)
	body := bytes.NewBufferString(`{"source":"wechat","sender":"partner","content":"   "}`)
	request := httptest.NewRequest(http.MethodPost, "/api/inbox/webhook", body)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-LifeOps-Token", "dev-token")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}

func TestWebhookAcceptsMessageWithoutSender(t *testing.T) {
	server := testServer(t)
	body := bytes.NewBufferString(`{"source":"telegram","content":"buy milk"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/inbox/webhook", body)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-LifeOps-Token", "dev-token")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", response.Code)
	}
}

func TestWebhookErrorResponseIsJSON(t *testing.T) {
	server := testServer(t)
	tests := []struct {
		name        string
		body        string
		wantCode    string
		wantMessage string
	}{
		{"invalid json", `{broken`, "invalid_request", "invalid json"},
		{"missing source", `{"content":"test"}`, "invalid_request", "source is required"},
		{"empty content", `{"source":"wechat","content":""}`, "empty_content", "content is required"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := bytes.NewBufferString(tt.body)
			request := httptest.NewRequest(http.MethodPost, "/api/inbox/webhook", body)
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("X-LifeOps-Token", "dev-token")
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)
			if !strings.Contains(response.Header().Get("Content-Type"), "application/json") {
				t.Fatalf("expected json content-type")
			}
			var errResp errorResponse
			json.Unmarshal(response.Body.Bytes(), &errResp)
			if errResp.Error != tt.wantCode || errResp.Message != tt.wantMessage {
				t.Fatalf("expected (%q, %q), got (%q, %q)", tt.wantCode, tt.wantMessage, errResp.Error, errResp.Message)
			}
		})
	}
}

func TestBearerTokenAuth(t *testing.T) {
	server := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/drafts/", nil)
	req.Header.Set("Authorization", "Bearer dev-token")
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	if w.Code == http.StatusUnauthorized {
		t.Fatal("valid bearer token should be accepted")
	}
}

func TestBearerTokenRejectsWrongToken(t *testing.T) {
	server := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/drafts/", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for wrong bearer token, got %d", w.Code)
	}
}

func TestXLifeOpsTokenAcceptedForAPIEndpoints(t *testing.T) {
	server := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/drafts/", nil)
	req.Header.Set("X-LifeOps-Token", "dev-token")
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	if w.Code == http.StatusUnauthorized {
		t.Fatal("X-LifeOps-Token should also be accepted for API endpoints")
	}
}

func TestAPINoTokenReturns401(t *testing.T) {
	server := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/drafts/", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", w.Code)
	}
}

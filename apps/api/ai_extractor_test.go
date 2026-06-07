package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExtractReturnsDraftJSON(t *testing.T) {
	wantContent := `{"drafts":[{"draft_type":"task","title":"buy milk","confidence":0.9}]}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("expected path /chat/completions, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("expected Authorization Bearer test-key, got %q", r.Header.Get("Authorization"))
		}

		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if reqBody["temperature"] != float64(0) {
			t.Errorf("expected temperature 0, got %v", reqBody["temperature"])
		}

		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"content": wantContent,
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	extractor := NewAIExtractor(server.URL, "test-model", "test-key")
	got, err := extractor.Extract("buy milk tomorrow")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != wantContent {
		t.Fatalf("expected %q, got %q", wantContent, got)
	}
}

func TestExtractWorksWithoutAPIKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Errorf("expected no Authorization header, got %q", r.Header.Get("Authorization"))
		}
		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{"message": map[string]interface{}{"content": `{"drafts":[]}`}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	extractor := NewAIExtractor(server.URL, "test-model", "")
	_, err := extractor.Extract("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExtractReturnsErrorOnNetworkFailure(t *testing.T) {
	extractor := NewAIExtractor("http://localhost:1", "test-model", "test-key")
	_, err := extractor.Extract("test")
	if err == nil {
		t.Fatal("expected error for network failure")
	}
}

func TestExtractReturnsErrorOnNon200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal"}`))
	}))
	defer server.Close()

	extractor := NewAIExtractor(server.URL, "test-model", "test-key")
	_, err := extractor.Extract("test")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestExtractReturnsErrorOnEmptyChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{"choices": []map[string]interface{}{}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	extractor := NewAIExtractor(server.URL, "test-model", "test-key")
	_, err := extractor.Extract("test")
	if err == nil {
		t.Fatal("expected error for empty choices")
	}
}

func TestExtractReturnsErrorWithoutBaseURL(t *testing.T) {
	extractor := NewAIExtractor("", "test-model", "")
	_, err := extractor.Extract("test")
	if err == nil {
		t.Fatal("expected error when base URL is empty")
	}
}

func TestBuildMessagesContainsUserContent(t *testing.T) {
	msgs := buildMessages("hello world")
	last := msgs[len(msgs)-1]
	if last["role"] != "user" {
		t.Fatalf("expected last message role user, got %q", last["role"])
	}
	if last["content"] != "hello world" {
		t.Fatalf("expected last message content 'hello world', got %q", last["content"])
	}
}

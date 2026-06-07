package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type InboxItem struct {
	ID        string    `json:"id"`
	Source    string    `json:"source"`
	Sender    string    `json:"sender"`
	Content   string    `json:"content"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type WebhookInboxRequest struct {
	Source  string `json:"source"`
	Sender  string `json:"sender"`
	Content string `json:"content"`
}

type errorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type Server struct {
	token string
	mux   *http.ServeMux
	store *Store
}

func NewServer(token string, store *Store) *Server {
	server := &Server{
		token: token,
		mux:   http.NewServeMux(),
		store: store,
	}
	server.mux.HandleFunc("/api/inbox/webhook", server.handleWebhookInbox)
	server.mux.HandleFunc("/api/drafts", server.handleListDrafts)
	server.mux.HandleFunc("/api/drafts/", server.handleDraftRoutes)
	return server
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) handleDraftRoutes(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/drafts/")
	if id == "" {
		s.handleListDrafts(w, r)
		return
	}
	s.handleGetDraft(w, r, id)
}

func writeJSONError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: code, Message: message})
}

func (s *Server) handleWebhookInbox(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if r.Header.Get("X-LifeOps-Token") != s.token {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var request WebhookInboxRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid_request", "invalid json")
		return
	}

	if strings.TrimSpace(request.Source) == "" {
		writeJSONError(w, http.StatusBadRequest, "invalid_request", "source is required")
		return
	}
	if strings.TrimSpace(request.Content) == "" {
		writeJSONError(w, http.StatusBadRequest, "empty_content", "content is required")
		return
	}

	item := InboxItem{
		ID:        newID(),
		Source:    request.Source,
		Sender:    request.Sender,
		Content:   request.Content,
		Status:    "new",
		CreatedAt: time.Now().UTC(),
	}

	if err := s.store.CreateInboxItem(&item); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal_error", "failed to store item")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(item)
}

func newID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return hex.EncodeToString([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(bytes[:])
}

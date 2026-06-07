package main

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleListDrafts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	status := r.URL.Query().Get("status")
	if status == "" {
		status = "pending"
	}

	drafts, err := s.store.ListDrafts(status)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal_error", "failed to list drafts")
		return
	}

	if drafts == nil {
		drafts = []Draft{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(drafts)
}

func (s *Server) handleGetDraft(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	draft, err := s.store.GetDraft(id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal_error", "failed to get draft")
		return
	}
	if draft == nil {
		writeJSONError(w, http.StatusNotFound, "not_found", "draft not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(draft)
}

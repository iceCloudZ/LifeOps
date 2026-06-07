package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type ReviewRequest struct {
	Action  string `json:"action"`
	Updates struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		DueAt       *string `json:"due_at"`
	} `json:"updates"`
}

func (s *Server) handleReviewDraft(w http.ResponseWriter, r *http.Request, draftID string) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	draft, err := s.store.GetDraft(draftID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal_error", "query failed")
		return
	}
	if draft == nil {
		writeJSONError(w, http.StatusNotFound, "not_found", "draft not found")
		return
	}
	if draft.Status != "pending" {
		writeJSONError(w, http.StatusConflict, "conflict", "draft is not pending")
		return
	}

	var req ReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid_request", "invalid json")
		return
	}

	switch req.Action {
	case "confirm":
		s.confirmDraft(w, r, draft, &req)
	case "discard":
		s.discardDraft(w, draft)
	default:
		writeJSONError(w, http.StatusBadRequest, "invalid_request", "action must be confirm or discard")
	}
}

func (s *Server) confirmDraft(w http.ResponseWriter, r *http.Request, draft *Draft, req *ReviewRequest) {
	if req.Updates.Title != "" {
		draft.Title = strings.TrimSpace(req.Updates.Title)
	}
	if req.Updates.Description != "" {
		draft.Description = strings.TrimSpace(req.Updates.Description)
	}
	if req.Updates.DueAt != nil {
		draft.DueAt = req.Updates.DueAt
	}

	now := time.Now().UTC().Format(time.RFC3339)
	entityID := newID()

	switch draft.DraftType {
	case "task":
		task := &FamilyTask{
			ID: entityID, Title: draft.Title, Description: draft.Description,
			DueAt: draft.DueAt, Status: "open", SourceInboxItemID: nilSafe(draft.InboxItemID),
			CreatedAt: now, UpdatedAt: now,
		}
		if err := s.store.CreateTask(task); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "create entity failed")
			return
		}
		s.store.UpdateDraftStatus(draft.ID, "confirmed", entityID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)

	case "event":
		event := &FamilyEvent{
			ID: entityID, Title: draft.Title, Description: draft.Description,
			StartsAt: draft.DueAt, SourceInboxItemID: nilSafe(draft.InboxItemID),
			CreatedAt: now, UpdatedAt: now,
		}
		if err := s.store.CreateEvent(event); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "create entity failed")
			return
		}
		s.store.UpdateDraftStatus(draft.ID, "confirmed", entityID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(event)

	case "shopping_item":
		qty := ""
		if draft.Quantity != nil {
			qty = *draft.Quantity
		}
		item := &ShoppingItem{
			ID: entityID, Name: draft.Title, Quantity: qty,
			Status: "open", SourceInboxItemID: nilSafe(draft.InboxItemID),
			CreatedAt: now, UpdatedAt: now,
		}
		if err := s.store.CreateShoppingItem(item); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "create entity failed")
			return
		}
		s.store.UpdateDraftStatus(draft.ID, "confirmed", entityID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(item)

	case "note":
		topic := ""
		if draft.Topic != nil {
			topic = *draft.Topic
		}
		note := &FamilyNote{
			ID: entityID, Title: draft.Title, Content: draft.Description,
			Topic: topic, SourceInboxItemID: nilSafe(draft.InboxItemID),
			CreatedAt: now, UpdatedAt: now,
		}
		if err := s.store.CreateNote(note); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "create entity failed")
			return
		}
		s.store.UpdateDraftStatus(draft.ID, "confirmed", entityID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(note)

	default:
		writeJSONError(w, http.StatusBadRequest, "invalid_request", "unknown draft type")
	}
}

func (s *Server) discardDraft(w http.ResponseWriter, draft *Draft) {
	if err := s.store.UpdateDraftStatus(draft.ID, "discarded", ""); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal_error", "update failed")
		return
	}
	updated, _ := s.store.GetDraft(draft.ID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

func nilSafe(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

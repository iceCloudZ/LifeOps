package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func createDraftForReview(t *testing.T, srv *Server, draftType, title string) string {
	t.Helper()
	now := "2026-06-08T12:00:00Z"
	draft := &Draft{
		ID: newID(), InboxItemID: "inbox-1", DraftType: draftType,
		Title: title, Description: "test", Confidence: 0.9,
		Status: "pending", CreatedAt: now, UpdatedAt: now,
	}
	if err := srv.store.CreateDraft(draft); err != nil {
		t.Fatalf("create draft: %v", err)
	}
	return draft.ID
}

func TestConfirmTaskDraft(t *testing.T) {
	srv := testServer(t)
	draftID := createDraftForReview(t, srv, "task", "交水费")

	body := bytes.NewBufferString(`{"action":"confirm"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/drafts/"+draftID+"/review", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var task FamilyTask
	json.Unmarshal(w.Body.Bytes(), &task)
	if task.Title != "交水费" {
		t.Fatalf("expected title 交水费, got %q", task.Title)
	}
	if task.Status != "open" {
		t.Fatalf("expected status open, got %q", task.Status)
	}

	draft, _ := srv.store.GetDraft(draftID)
	if draft.Status != "confirmed" {
		t.Fatalf("expected draft status confirmed, got %q", draft.Status)
	}
	if draft.EntityID == nil || *draft.EntityID != task.ID {
		t.Fatalf("expected entity_id to match task id")
	}
}

func TestConfirmEventDraft(t *testing.T) {
	srv := testServer(t)
	draftID := createDraftForReview(t, srv, "event", "家长会")

	body := bytes.NewBufferString(`{"action":"confirm"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/drafts/"+draftID+"/review", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var event FamilyEvent
	json.Unmarshal(w.Body.Bytes(), &event)
	if event.Title != "家长会" {
		t.Fatalf("expected title 家长会, got %q", event.Title)
	}
}

func TestConfirmShoppingItemDraft(t *testing.T) {
	srv := testServer(t)
	draftID := createDraftForReview(t, srv, "shopping_item", "牛奶")

	body := bytes.NewBufferString(`{"action":"confirm"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/drafts/"+draftID+"/review", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var item ShoppingItem
	json.Unmarshal(w.Body.Bytes(), &item)
	if item.Name != "牛奶" {
		t.Fatalf("expected name 牛奶, got %q", item.Name)
	}
}

func TestConfirmNoteDraft(t *testing.T) {
	srv := testServer(t)
	draftID := createDraftForReview(t, srv, "note", "晚餐偏好")

	body := bytes.NewBufferString(`{"action":"confirm"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/drafts/"+draftID+"/review", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var note FamilyNote
	json.Unmarshal(w.Body.Bytes(), &note)
	if note.Title != "晚餐偏好" {
		t.Fatalf("expected title 晚餐偏好, got %q", note.Title)
	}
}

func TestConfirmDraftWithUpdates(t *testing.T) {
	srv := testServer(t)
	draftID := createDraftForReview(t, srv, "task", "old title")

	body := bytes.NewBufferString(`{"action":"confirm","updates":{"title":"new title","description":"new desc"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/drafts/"+draftID+"/review", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var task FamilyTask
	json.Unmarshal(w.Body.Bytes(), &task)
	if task.Title != "new title" {
		t.Fatalf("expected title 'new title', got %q", task.Title)
	}
	if task.Description != "new desc" {
		t.Fatalf("expected description 'new desc', got %q", task.Description)
	}
}

func TestDiscardDraft(t *testing.T) {
	srv := testServer(t)
	draftID := createDraftForReview(t, srv, "task", "test")

	body := bytes.NewBufferString(`{"action":"discard"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/drafts/"+draftID+"/review", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	draft, _ := srv.store.GetDraft(draftID)
	if draft.Status != "discarded" {
		t.Fatalf("expected status discarded, got %q", draft.Status)
	}
}

func TestReviewNonPendingDraft(t *testing.T) {
	srv := testServer(t)
	draftID := createDraftForReview(t, srv, "task", "test")

	srv.store.UpdateDraftStatus(draftID, "confirmed", "entity-1")

	body := bytes.NewBufferString(`{"action":"confirm"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/drafts/"+draftID+"/review", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestReviewNonexistentDraft(t *testing.T) {
	srv := testServer(t)

	body := bytes.NewBufferString(`{"action":"confirm"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/drafts/nonexistent/review", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestReviewInvalidAction(t *testing.T) {
	srv := testServer(t)
	draftID := createDraftForReview(t, srv, "task", "test")

	body := bytes.NewBufferString(`{"action":"invalid"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/drafts/"+draftID+"/review", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

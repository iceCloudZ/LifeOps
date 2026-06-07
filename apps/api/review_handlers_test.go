package main

import (
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

	req := authRequest(http.MethodPost, "/api/drafts/"+draftID+"/review", `{"action":"confirm"}`)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var task FamilyTask
	json.Unmarshal(w.Body.Bytes(), &task)
	if task.Title != "交水费" || task.Status != "open" {
		t.Fatalf("unexpected task: %+v", task)
	}

	draft, _ := srv.store.GetDraft(draftID)
	if draft.Status != "confirmed" {
		t.Fatalf("expected draft status confirmed, got %q", draft.Status)
	}
}

func TestConfirmEventDraft(t *testing.T) {
	srv := testServer(t)
	draftID := createDraftForReview(t, srv, "event", "家长会")

	req := authRequest(http.MethodPost, "/api/drafts/"+draftID+"/review", `{"action":"confirm"}`)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestConfirmShoppingItemDraft(t *testing.T) {
	srv := testServer(t)
	draftID := createDraftForReview(t, srv, "shopping_item", "牛奶")

	req := authRequest(http.MethodPost, "/api/drafts/"+draftID+"/review", `{"action":"confirm"}`)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestConfirmNoteDraft(t *testing.T) {
	srv := testServer(t)
	draftID := createDraftForReview(t, srv, "note", "晚餐偏好")

	req := authRequest(http.MethodPost, "/api/drafts/"+draftID+"/review", `{"action":"confirm"}`)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestConfirmDraftWithUpdates(t *testing.T) {
	srv := testServer(t)
	draftID := createDraftForReview(t, srv, "task", "old title")

	req := authRequest(http.MethodPost, "/api/drafts/"+draftID+"/review", `{"action":"confirm","updates":{"title":"new title","description":"new desc"}}`)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var task FamilyTask
	json.Unmarshal(w.Body.Bytes(), &task)
	if task.Title != "new title" || task.Description != "new desc" {
		t.Fatalf("unexpected task: %+v", task)
	}
}

func TestDiscardDraft(t *testing.T) {
	srv := testServer(t)
	draftID := createDraftForReview(t, srv, "task", "test")

	req := authRequest(http.MethodPost, "/api/drafts/"+draftID+"/review", `{"action":"discard"}`)
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

	req := authRequest(http.MethodPost, "/api/drafts/"+draftID+"/review", `{"action":"confirm"}`)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestReviewNonexistentDraft(t *testing.T) {
	srv := testServer(t)

	req := authRequest(http.MethodPost, "/api/drafts/nonexistent/review", `{"action":"confirm"}`)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestReviewInvalidAction(t *testing.T) {
	srv := testServer(t)
	draftID := createDraftForReview(t, srv, "task", "test")

	req := authRequest(http.MethodPost, "/api/drafts/"+draftID+"/review", `{"action":"invalid"}`)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestDraftEndpointsRequireAuth(t *testing.T) {
	srv := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/drafts/", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without auth, got %d", w.Code)
	}
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

type e2eClient struct {
	baseURL string
	token   string
	client  *http.Client
	store   *Store
}

func newE2EClient(t *testing.T) *e2eClient {
	t.Helper()
	dir := t.TempDir()
	store, err := NewStore(filepath.Join(dir, "e2e.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })

	srv := NewServer("test-token", store, nil)
	ts := httptest.NewServer(srv)
	t.Cleanup(func() { ts.Close() })

	return &e2eClient{baseURL: ts.URL, token: "test-token", client: ts.Client(), store: store}
}

func (c *e2eClient) post(path, body string) (*http.Response, map[string]interface{}) {
	req, _ := http.NewRequest(http.MethodPost, c.baseURL+path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, nil
	}
	respBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)
	return resp, result
}

func (c *e2eClient) postWebhook(body string) (*http.Response, map[string]interface{}) {
	req, _ := http.NewRequest(http.MethodPost, c.baseURL+"/api/inbox/webhook", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-LifeOps-Token", c.token)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, nil
	}
	respBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)
	return resp, result
}

func (c *e2eClient) get(path string) (*http.Response, []map[string]interface{}) {
	req, _ := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, nil
	}
	respBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	var result []map[string]interface{}
	json.Unmarshal(respBody, &result)
	return resp, result
}

func (c *e2eClient) getOne(path string) (*http.Response, map[string]interface{}) {
	req, _ := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, nil
	}
	respBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)
	return resp, result
}

func (c *e2eClient) delete(path string) *http.Response {
	req, _ := http.NewRequest(http.MethodDelete, c.baseURL+path, nil)
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, _ := c.client.Do(req)
	resp.Body.Close()
	return resp
}

// E2E: Webhook → Inbox → Manual Draft → Review → Entity

func TestE2E_WebhookToEntity(t *testing.T) {
	client := newE2EClient(t)

	// Step 1: POST message to webhook inbox
	resp, inboxItem := client.postWebhook(`{
		"source": "wechat",
		"sender": "partner",
		"content": "周五孩子要带彩笔，别忘了交水费。"
	}`)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("step 1 webhook: expected 201, got %d", resp.StatusCode)
	}
	inboxID := inboxItem["id"].(string)
	t.Logf("step 1: inbox item created, id=%s", inboxID)

	// Step 2: Simulate AI extraction — create drafts manually (AI module would do this)
	now := "2026-06-08T12:00:00Z"
	draft1ID := newID()
	client.store.CreateDraft(&Draft{
		ID: draft1ID, InboxItemID: inboxID, DraftType: "task",
		Title: "准备彩笔", Description: "周五孩子需要带彩笔",
		Confidence: 0.91, Status: "pending", CreatedAt: now, UpdatedAt: now,
	})
	draft2ID := newID()
	client.store.CreateDraft(&Draft{
		ID: draft2ID, InboxItemID: inboxID, DraftType: "task",
		Title: "交水费", Description: "提醒家庭处理水费缴纳",
		Confidence: 0.86, Status: "pending", CreatedAt: now, UpdatedAt: now,
	})
	t.Logf("step 2: 2 drafts created from AI extraction")

	// Step 3: List pending drafts
	resp, drafts := client.get("/api/drafts/?status=pending")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("step 3 list drafts: expected 200, got %d", resp.StatusCode)
	}
	if len(drafts) != 2 {
		t.Fatalf("step 3: expected 2 pending drafts, got %d", len(drafts))
	}
	t.Logf("step 3: listed %d pending drafts", len(drafts))

	// Step 4: Confirm first draft → creates a task
	resp, entity := client.post("/api/drafts/"+draft1ID+"/review", `{"action":"confirm"}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("step 4 confirm: expected 200, got %d", resp.StatusCode)
	}
	taskID := entity["id"].(string)
	if entity["title"].(string) != "准备彩笔" {
		t.Fatalf("step 4: expected title 准备彩笔, got %v", entity["title"])
	}
	t.Logf("step 4: draft confirmed → task created, id=%s", taskID)

	// Step 5: Verify task appears in entity list
	resp, tasks := client.get("/api/tasks/")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("step 5 list tasks: expected 200, got %d", resp.StatusCode)
	}
	if len(tasks) != 1 {
		t.Fatalf("step 5: expected 1 task, got %d", len(tasks))
	}
	t.Logf("step 5: task verified in entity list")

	// Step 6: Get the task
	resp, task := client.getOne("/api/tasks/" + taskID)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("step 6 get task: expected 200, got %d", resp.StatusCode)
	}
	if task["status"].(string) != "open" {
		t.Fatalf("step 6: expected status open, got %v", task["status"])
	}
	t.Logf("step 6: task detail retrieved, status=open")

	// Step 7: Update the task
	resp, _ = client.post("/api/tasks/"+taskID, "")
	// PUT via authRequest helper doesn't exist in e2eClient, use raw method
	updateBody := `{"title":"准备彩笔（美术课）","description":"周五美术课","status":"done"}`
	req, _ := http.NewRequest(http.MethodPut, client.baseURL+"/api/tasks/"+taskID, bytes.NewBufferString(updateBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+client.token)
	resp, _ = client.client.Do(req)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("step 7 update task: expected 200, got %d", resp.StatusCode)
	}
	t.Logf("step 7: task updated")

	// Step 8: Discard second draft
	resp, _ = client.post("/api/drafts/"+draft2ID+"/review", `{"action":"discard"}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("step 8 discard: expected 200, got %d", resp.StatusCode)
	}
	t.Logf("step 8: second draft discarded")

	// Step 9: Verify no more pending drafts
	resp, pendingDrafts := client.get("/api/drafts/?status=pending")
	if len(pendingDrafts) != 0 {
		t.Fatalf("step 9: expected 0 pending drafts, got %d", len(pendingDrafts))
	}

	// Step 10: Delete the task
	resp2 := client.delete("/api/tasks/" + taskID)
	if resp2.StatusCode != http.StatusNoContent {
		t.Fatalf("step 10 delete: expected 204, got %d", resp2.StatusCode)
	}

	// Step 11: Verify task is gone
	resp, _ = client.getOne("/api/tasks/" + taskID)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("step 11: expected 404 after delete, got %d", resp.StatusCode)
	}
	t.Logf("steps 9-11: cleanup verified — 0 pending drafts, task deleted")
}

func TestE2E_ShoppingItemFlow(t *testing.T) {
	client := newE2EClient(t)

	// Create a shopping_item draft via extraction simulation
	now := "2026-06-08T12:00:00Z"
	draftID := newID()
	client.store.CreateDraft(&Draft{
		ID: draftID, InboxItemID: "inbox-1", DraftType: "shopping_item",
		Title: "牛奶", Confidence: 0.95,
		Status: "pending", CreatedAt: now, UpdatedAt: now,
	})

	// Confirm → creates shopping item
	resp, item := client.post("/api/drafts/"+draftID+"/review", `{"action":"confirm"}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if item["name"].(string) != "牛奶" {
		t.Fatalf("expected name 牛奶, got %v", item["name"])
	}
	itemID := item["id"].(string)

	// Verify in shopping items list
	resp, items := client.get("/api/shopping-items/")
	if len(items) != 1 {
		t.Fatalf("expected 1 shopping item, got %d", len(items))
	}
	if items[0]["name"].(string) != "牛奶" {
		t.Fatalf("expected name 牛奶, got %v", items[0]["name"])
	}

	// Mark as bought
	updateBody := fmt.Sprintf(`{"name":"牛奶","quantity":"","status":"bought"}`)
	req, _ := http.NewRequest(http.MethodPut, client.baseURL+"/api/shopping-items/"+itemID, bytes.NewBufferString(updateBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+client.token)
	resp, _ = client.client.Do(req)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("update shopping item: expected 200, got %d", resp.StatusCode)
	}

	// Verify status changed
	resp, updated := client.getOne("/api/shopping-items/" + itemID)
	if updated["status"].(string) != "bought" {
		t.Fatalf("expected status bought, got %v", updated["status"])
	}
}

func TestE2E_AuthEnforced(t *testing.T) {
	client := newE2EClient(t)

	// No auth → 401
	req, _ := http.NewRequest(http.MethodGet, client.baseURL+"/api/tasks/", nil)
	resp, _ := client.client.Do(req)
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 without auth, got %d", resp.StatusCode)
	}

	// Wrong token → 401
	req, _ = http.NewRequest(http.MethodGet, client.baseURL+"/api/tasks/", nil)
	req.Header.Set("Authorization", "Bearer wrong")
	resp, _ = client.client.Do(req)
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 with wrong token, got %d", resp.StatusCode)
	}

	// Webhook without token → 401
	req, _ = http.NewRequest(http.MethodPost, client.baseURL+"/api/inbox/webhook", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = client.client.Do(req)
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for webhook without token, got %d", resp.StatusCode)
	}
}

func TestE2E_DataPersistsInSQLite(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "persist-e2e.db")

	// Session 1: create store, server, post a task
	store1, _ := NewStore(dbPath)
	srv1 := NewServer("token1", store1, nil)
	ts1 := httptest.NewServer(srv1)

	req, _ := http.NewRequest(http.MethodPost, ts1.URL+"/api/tasks/", bytes.NewBufferString(`{"title":"persistent task"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token1")
	resp, _ := ts1.Client().Do(req)
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create task: expected 201, got %d", resp.StatusCode)
	}
	ts1.Close()
	store1.Close()

	// Session 2: new store and server with same DB
	store2, _ := NewStore(dbPath)
	srv2 := NewServer("token1", store2, nil)
	ts2 := httptest.NewServer(srv2)
	defer ts2.Close()
	defer store2.Close()

	req, _ = http.NewRequest(http.MethodGet, ts2.URL+"/api/tasks/", nil)
	req.Header.Set("Authorization", "Bearer token1")
	resp, _ = ts2.Client().Do(req)
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	var tasks []map[string]interface{}
	json.Unmarshal(body, &tasks)
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task after restart, got %d", len(tasks))
	}
	if tasks[0]["title"].(string) != "persistent task" {
		t.Fatalf("expected title 'persistent task', got %v", tasks[0]["title"])
	}
}

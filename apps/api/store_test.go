package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func testStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func TestStoreMigrationCreatesTables(t *testing.T) {
	store := testStore(t)

	tables := []string{"inbox_items", "drafts", "family_tasks", "family_events", "shopping_items", "family_notes"}
	for _, table := range tables {
		var name string
		err := store.db.QueryRow(
			`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table,
		).Scan(&name)
		if err != nil {
			t.Fatalf("expected table %s to exist: %v", table, err)
		}
	}
}

func TestCreateAndGetInboxItem(t *testing.T) {
	store := testStore(t)

	now := time.Now().UTC().Truncate(time.Second)
	item := &InboxItem{
		ID:        "test-id-1",
		Source:    "wechat",
		Sender:    "partner",
		Content:   "test content",
		Status:    "new",
		CreatedAt: now,
	}

	if err := store.CreateInboxItem(item); err != nil {
		t.Fatalf("create inbox item: %v", err)
	}

	var got InboxItem
	var createdAt string
	err := store.db.QueryRow(
		`SELECT id, source, sender, content, status, created_at FROM inbox_items WHERE id = ?`,
		"test-id-1",
	).Scan(&got.ID, &got.Source, &got.Sender, &got.Content, &got.Status, &createdAt)
	if err != nil {
		t.Fatalf("query inbox item: %v", err)
	}
	if got.Source != "wechat" {
		t.Fatalf("expected source wechat, got %q", got.Source)
	}
	if got.Content != "test content" {
		t.Fatalf("expected content 'test content', got %q", got.Content)
	}
}

func TestCreateAndListDrafts(t *testing.T) {
	store := testStore(t)

	now := time.Now().UTC().Format(time.RFC3339)
	draft := &Draft{
		ID:          "draft-1",
		InboxItemID: "inbox-1",
		DraftType:   "task",
		Title:       "test task",
		Description: "test desc",
		Confidence:  0.9,
		Status:      "pending",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := store.CreateDraft(draft); err != nil {
		t.Fatalf("create draft: %v", err)
	}

	drafts, err := store.ListDrafts("pending")
	if err != nil {
		t.Fatalf("list drafts: %v", err)
	}
	if len(drafts) != 1 {
		t.Fatalf("expected 1 draft, got %d", len(drafts))
	}
	if drafts[0].Title != "test task" {
		t.Fatalf("expected title 'test task', got %q", drafts[0].Title)
	}
}

func TestGetDraft(t *testing.T) {
	store := testStore(t)

	now := time.Now().UTC().Format(time.RFC3339)
	draft := &Draft{
		ID:          "draft-2",
		InboxItemID: "inbox-1",
		DraftType:   "event",
		Title:       "test event",
		Confidence:  0.8,
		Status:      "pending",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	store.CreateDraft(draft)

	got, err := store.GetDraft("draft-2")
	if err != nil {
		t.Fatalf("get draft: %v", err)
	}
	if got == nil {
		t.Fatal("expected draft, got nil")
	}
	if got.Title != "test event" {
		t.Fatalf("expected title 'test event', got %q", got.Title)
	}
}

func TestGetDraftNotFound(t *testing.T) {
	store := testStore(t)

	got, err := store.GetDraft("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatal("expected nil for nonexistent draft")
	}
}

func TestUpdateDraftStatus(t *testing.T) {
	store := testStore(t)

	now := time.Now().UTC().Format(time.RFC3339)
	draft := &Draft{
		ID:          "draft-3",
		InboxItemID: "inbox-1",
		DraftType:   "task",
		Title:       "test",
		Confidence:  0.9,
		Status:      "pending",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	store.CreateDraft(draft)

	if err := store.UpdateDraftStatus("draft-3", "confirmed", "entity-1"); err != nil {
		t.Fatalf("update draft status: %v", err)
	}

	got, _ := store.GetDraft("draft-3")
	if got.Status != "confirmed" {
		t.Fatalf("expected status confirmed, got %q", got.Status)
	}
	if got.EntityID == nil || *got.EntityID != "entity-1" {
		t.Fatalf("expected entity_id entity-1, got %v", got.EntityID)
	}
}

func TestListDraftsFiltersByStatus(t *testing.T) {
	store := testStore(t)

	now := time.Now().UTC().Format(time.RFC3339)
	for i, status := range []string{"pending", "confirmed", "pending"} {
		store.CreateDraft(&Draft{
			ID:          "draft-f-" + string(rune('a'+i)),
			InboxItemID: "inbox-1",
			DraftType:   "task",
			Title:       "test",
			Confidence:  0.9,
			Status:      status,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
	}

	pending, _ := store.ListDrafts("pending")
	if len(pending) != 2 {
		t.Fatalf("expected 2 pending drafts, got %d", len(pending))
	}

	confirmed, _ := store.ListDrafts("confirmed")
	if len(confirmed) != 1 {
		t.Fatalf("expected 1 confirmed draft, got %d", len(confirmed))
	}

	all, _ := store.ListDrafts("")
	if len(all) != 3 {
		t.Fatalf("expected 3 total drafts, got %d", len(all))
	}
}

func TestCreateTask(t *testing.T) {
	store := testStore(t)

	now := time.Now().UTC().Format(time.RFC3339)
	task := &FamilyTask{
		ID:        "task-1",
		Title:     "交水费",
		Status:    "open",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := store.CreateTask(task); err != nil {
		t.Fatalf("create task: %v", err)
	}

	var title string
	err := store.db.QueryRow(`SELECT title FROM family_tasks WHERE id = ?`, "task-1").Scan(&title)
	if err != nil {
		t.Fatalf("query task: %v", err)
	}
	if title != "交水费" {
		t.Fatalf("expected title '交水费', got %q", title)
	}
}

func TestCreateEvent(t *testing.T) {
	store := testStore(t)

	now := time.Now().UTC().Format(time.RFC3339)
	event := &FamilyEvent{
		ID:        "event-1",
		Title:     "家长会",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := store.CreateEvent(event); err != nil {
		t.Fatalf("create event: %v", err)
	}
}

func TestCreateShoppingItem(t *testing.T) {
	store := testStore(t)

	now := time.Now().UTC().Format(time.RFC3339)
	item := &ShoppingItem{
		ID:        "shop-1",
		Name:      "牛奶",
		Quantity:  "2盒",
		Status:    "open",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := store.CreateShoppingItem(item); err != nil {
		t.Fatalf("create shopping item: %v", err)
	}
}

func TestCreateNote(t *testing.T) {
	store := testStore(t)

	now := time.Now().UTC().Format(time.RFC3339)
	note := &FamilyNote{
		ID:        "note-1",
		Title:     "晚餐偏好",
		Content:   "少油少辣",
		Topic:     "food",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := store.CreateNote(note); err != nil {
		t.Fatalf("create note: %v", err)
	}
}

func TestStorePersistsAcrossInstances(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "persist.db")

	now := time.Now().UTC().Format(time.RFC3339)

	s1, err := NewStore(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	s1.CreateDraft(&Draft{
		ID: "persist-1", InboxItemID: "inbox-1", DraftType: "task",
		Title: "persistent", Confidence: 0.9, Status: "pending",
		CreatedAt: now, UpdatedAt: now,
	})
	s1.Close()

	s2, err := NewStore(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer s2.Close()

	draft, err := s2.GetDraft("persist-1")
	if err != nil {
		t.Fatal(err)
	}
	if draft == nil {
		t.Fatal("expected draft to persist across store instances")
	}
	if draft.Title != "persistent" {
		t.Fatalf("expected title 'persistent', got %q", draft.Title)
	}
}

func TestMain_StoreInitFromEnv(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "env-test.db")

	os.Setenv("LIFEOPS_DB_PATH", dbPath)
	defer os.Unsetenv("LIFEOPS_DB_PATH")

	s, err := NewStore(os.Getenv("LIFEOPS_DB_PATH"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
}

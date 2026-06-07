package main

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
	mu sync.Mutex
}

func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	db.SetMaxOpenConns(1)

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS inbox_items (
			id         TEXT PRIMARY KEY,
			source     TEXT NOT NULL,
			sender     TEXT DEFAULT '',
			content    TEXT NOT NULL,
			status     TEXT NOT NULL DEFAULT 'new',
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS drafts (
			id             TEXT PRIMARY KEY,
			inbox_item_id  TEXT NOT NULL,
			draft_type     TEXT NOT NULL,
			title          TEXT NOT NULL,
			description    TEXT DEFAULT '',
			confidence     REAL NOT NULL DEFAULT 0,
			due_at         TEXT,
			assignee_hint  TEXT,
			quantity       TEXT,
			topic          TEXT,
			status         TEXT NOT NULL DEFAULT 'pending',
			entity_id      TEXT,
			created_at     DATETIME NOT NULL,
			updated_at     DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS family_tasks (
			id                  TEXT PRIMARY KEY,
			title               TEXT NOT NULL,
			description         TEXT DEFAULT '',
			assignee_member_id  TEXT,
			due_at              TEXT,
			status              TEXT NOT NULL DEFAULT 'open',
			source_inbox_item_id TEXT,
			created_at          DATETIME NOT NULL,
			updated_at          DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS family_events (
			id                   TEXT PRIMARY KEY,
			title                TEXT NOT NULL,
			description          TEXT DEFAULT '',
			starts_at            TEXT,
			ends_at              TEXT,
			participant_members  TEXT DEFAULT '',
			source_inbox_item_id TEXT,
			created_at           DATETIME NOT NULL,
			updated_at           DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS shopping_items (
			id                   TEXT PRIMARY KEY,
			name                 TEXT NOT NULL,
			quantity             TEXT DEFAULT '',
			status               TEXT NOT NULL DEFAULT 'open',
			source_inbox_item_id TEXT,
			created_at           DATETIME NOT NULL,
			updated_at           DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS family_notes (
			id                   TEXT PRIMARY KEY,
			title                TEXT NOT NULL,
			content              TEXT DEFAULT '',
			topic                TEXT DEFAULT '',
			source_inbox_item_id TEXT,
			created_at           DATETIME NOT NULL,
			updated_at           DATETIME NOT NULL
		)`,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, m := range migrations {
		if _, err := s.db.Exec(m); err != nil {
			return fmt.Errorf("exec migration: %w", err)
		}
	}
	return nil
}

// InboxItem operations

func (s *Store) CreateInboxItem(item *InboxItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(
		`INSERT INTO inbox_items (id, source, sender, content, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		item.ID, item.Source, item.Sender, item.Content, item.Status,
		item.CreatedAt.Format(time.RFC3339), item.CreatedAt.Format(time.RFC3339),
	)
	return err
}

func (s *Store) UpdateInboxItemStatus(id, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(
		`UPDATE inbox_items SET status = ?, updated_at = ? WHERE id = ?`,
		status, time.Now().UTC().Format(time.RFC3339), id,
	)
	return err
}

// Draft operations

type Draft struct {
	ID            string   `json:"id"`
	InboxItemID   string   `json:"inbox_item_id"`
	DraftType     string   `json:"draft_type"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Confidence    float64  `json:"confidence"`
	DueAt         *string  `json:"due_at"`
	AssigneeHint  *string  `json:"assignee_hint"`
	Quantity      *string  `json:"quantity"`
	Topic         *string  `json:"topic"`
	Status        string   `json:"status"`
	EntityID      *string  `json:"entity_id"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

func (s *Store) CreateDraft(draft *Draft) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(
		`INSERT INTO drafts (id, inbox_item_id, draft_type, title, description, confidence,
		 due_at, assignee_hint, quantity, topic, status, entity_id, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		draft.ID, draft.InboxItemID, draft.DraftType, draft.Title, draft.Description,
		draft.Confidence, draft.DueAt, draft.AssigneeHint, draft.Quantity, draft.Topic,
		draft.Status, draft.EntityID, draft.CreatedAt, draft.UpdatedAt,
	)
	return err
}

func (s *Store) ListDrafts(status string) ([]Draft, error) {
	query := `SELECT id, inbox_item_id, draft_type, title, description, confidence,
	          due_at, assignee_hint, quantity, topic, status, entity_id, created_at, updated_at
	          FROM drafts`
	var args []interface{}
	if status != "" {
		query += ` WHERE status = ?`
		args = append(args, status)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var drafts []Draft
	for rows.Next() {
		var d Draft
		if err := rows.Scan(&d.ID, &d.InboxItemID, &d.DraftType, &d.Title, &d.Description,
			&d.Confidence, &d.DueAt, &d.AssigneeHint, &d.Quantity, &d.Topic,
			&d.Status, &d.EntityID, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		drafts = append(drafts, d)
	}
	return drafts, rows.Err()
}

func (s *Store) GetDraft(id string) (*Draft, error) {
	var d Draft
	err := s.db.QueryRow(
		`SELECT id, inbox_item_id, draft_type, title, description, confidence,
		 due_at, assignee_hint, quantity, topic, status, entity_id, created_at, updated_at
		 FROM drafts WHERE id = ?`, id,
	).Scan(&d.ID, &d.InboxItemID, &d.DraftType, &d.Title, &d.Description,
		&d.Confidence, &d.DueAt, &d.AssigneeHint, &d.Quantity, &d.Topic,
		&d.Status, &d.EntityID, &d.CreatedAt, &d.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (s *Store) UpdateDraftStatus(id, status, entityID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(
		`UPDATE drafts SET status = ?, entity_id = ?, updated_at = ? WHERE id = ?`,
		status, entityID, time.Now().UTC().Format(time.RFC3339), id,
	)
	return err
}

func (s *Store) UpdateDraftFields(id, title, description string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(
		`UPDATE drafts SET title = ?, description = ?, updated_at = ? WHERE id = ?`,
		title, description, time.Now().UTC().Format(time.RFC3339), id,
	)
	return err
}

// Entity creation operations

type FamilyTask struct {
	ID                  string `json:"id"`
	Title               string `json:"title"`
	Description         string `json:"description"`
	AssigneeMemberID    *string `json:"assignee_member_id"`
	DueAt               *string `json:"due_at"`
	Status              string `json:"status"`
	SourceInboxItemID   *string `json:"source_inbox_item_id"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
}

type FamilyEvent struct {
	ID                  string `json:"id"`
	Title               string `json:"title"`
	Description         string `json:"description"`
	StartsAt            *string `json:"starts_at"`
	EndsAt              *string `json:"ends_at"`
	ParticipantMembers  string `json:"participant_members"`
	SourceInboxItemID   *string `json:"source_inbox_item_id"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
}

type ShoppingItem struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Quantity            string `json:"quantity"`
	Status              string `json:"status"`
	SourceInboxItemID   *string `json:"source_inbox_item_id"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
}

type FamilyNote struct {
	ID                  string `json:"id"`
	Title               string `json:"title"`
	Content             string `json:"content"`
	Topic               string `json:"topic"`
	SourceInboxItemID   *string `json:"source_inbox_item_id"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
}

func (s *Store) CreateTask(task *FamilyTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(
		`INSERT INTO family_tasks (id, title, description, assignee_member_id, due_at, status, source_inbox_item_id, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		task.ID, task.Title, task.Description, task.AssigneeMemberID, task.DueAt,
		task.Status, task.SourceInboxItemID, task.CreatedAt, task.UpdatedAt,
	)
	return err
}

func (s *Store) CreateEvent(event *FamilyEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(
		`INSERT INTO family_events (id, title, description, starts_at, ends_at, participant_members, source_inbox_item_id, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		event.ID, event.Title, event.Description, event.StartsAt, event.EndsAt,
		event.ParticipantMembers, event.SourceInboxItemID, event.CreatedAt, event.UpdatedAt,
	)
	return err
}

func (s *Store) CreateShoppingItem(item *ShoppingItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(
		`INSERT INTO shopping_items (id, name, quantity, status, source_inbox_item_id, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		item.ID, item.Name, item.Quantity, item.Status,
		item.SourceInboxItemID, item.CreatedAt, item.UpdatedAt,
	)
	return err
}

func (s *Store) CreateNote(note *FamilyNote) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(
		`INSERT INTO family_notes (id, title, content, topic, source_inbox_item_id, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		note.ID, note.Title, note.Content, note.Topic,
		note.SourceInboxItemID, note.CreatedAt, note.UpdatedAt,
	)
	return err
}

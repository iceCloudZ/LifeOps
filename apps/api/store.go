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
		`CREATE TABLE IF NOT EXISTS family_members (
			id         TEXT PRIMARY KEY,
			name       TEXT NOT NULL,
			role       TEXT NOT NULL,
			birth_date TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS finance_accounts (
			id         TEXT PRIMARY KEY,
			member_id  TEXT,
			name       TEXT NOT NULL,
			type       TEXT NOT NULL,
			balance    REAL NOT NULL DEFAULT 0,
			note       TEXT DEFAULT '',
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (member_id) REFERENCES family_members(id)
		)`,
		`CREATE TABLE IF NOT EXISTS finance_records (
			id          TEXT PRIMARY KEY,
			member_id   TEXT,
			type        TEXT NOT NULL,
			amount      REAL NOT NULL,
			currency    TEXT NOT NULL DEFAULT 'CNY',
			category    TEXT NOT NULL,
			note        TEXT DEFAULT '',
			record_date TEXT NOT NULL,
			created_at  DATETIME NOT NULL,
			updated_at  DATETIME NOT NULL,
			FOREIGN KEY (member_id) REFERENCES family_members(id)
		)`,
		`CREATE TABLE IF NOT EXISTS health_profiles (
			id         TEXT PRIMARY KEY,
			member_id  TEXT NOT NULL UNIQUE,
			summary    TEXT NOT NULL DEFAULT '',
			updated_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS health_records (
			id          TEXT PRIMARY KEY,
			member_id   TEXT NOT NULL,
			type        TEXT NOT NULL,
			metric      TEXT,
			value       TEXT,
			unit        TEXT,
			note        TEXT DEFAULT '',
			record_date TEXT NOT NULL,
			created_at  DATETIME NOT NULL,
			updated_at  DATETIME NOT NULL,
			FOREIGN KEY (member_id) REFERENCES family_members(id)
		)`,
		`CREATE TABLE IF NOT EXISTS work_statuses (
			id         TEXT PRIMARY KEY,
			member_id  TEXT NOT NULL UNIQUE,
			summary    TEXT NOT NULL DEFAULT '',
			updated_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS work_profiles (
			id                TEXT PRIMARY KEY,
			member_id         TEXT NOT NULL UNIQUE,
			employment_status TEXT DEFAULT '',
			company           TEXT DEFAULT '',
			position          TEXT DEFAULT '',
			industry          TEXT DEFAULT '',
			work_location     TEXT DEFAULT '',
			income_range      TEXT DEFAULT '',
			work_schedule     TEXT DEFAULT '',
			commute_minutes   INTEGER DEFAULT 0,
			started_at        TEXT DEFAULT '',
			note              TEXT DEFAULT '',
			updated_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (member_id) REFERENCES family_members(id)
		)`,
		`CREATE TABLE IF NOT EXISTS work_records (
			id         TEXT PRIMARY KEY,
			member_id  TEXT NOT NULL,
			type       TEXT NOT NULL,
			title      TEXT NOT NULL,
			status     TEXT NOT NULL DEFAULT 'active',
			priority   TEXT NOT NULL DEFAULT 'medium',
			project    TEXT DEFAULT '',
			due_date   TEXT,
			note       TEXT DEFAULT '',
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (member_id) REFERENCES family_members(id)
		)`,
		`CREATE TABLE IF NOT EXISTS family_statuses (
			id         TEXT PRIMARY KEY,
			summary    TEXT NOT NULL DEFAULT '',
			updated_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS family_records (
			id              TEXT PRIMARY KEY,
			member_id       TEXT,
			type            TEXT NOT NULL,
			title           TEXT NOT NULL,
			status          TEXT NOT NULL DEFAULT 'pending',
			location        TEXT DEFAULT '',
			participants    TEXT DEFAULT '[]',
			scheduled_date  TEXT,
			note            TEXT DEFAULT '',
			created_at      DATETIME NOT NULL,
			updated_at      DATETIME NOT NULL,
			FOREIGN KEY (member_id) REFERENCES family_members(id)
		)`,
		`CREATE TABLE IF NOT EXISTS knowledge_notes (
			id         TEXT PRIMARY KEY,
			domain     TEXT,
			member_id  TEXT,
			title      TEXT NOT NULL,
			content    TEXT NOT NULL,
			tags       TEXT DEFAULT '[]',
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (member_id) REFERENCES family_members(id)
		)`,
		`CREATE TABLE IF NOT EXISTS conversations (
			id         TEXT PRIMARY KEY,
			title      TEXT DEFAULT '',
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id              TEXT PRIMARY KEY,
			conversation_id TEXT NOT NULL,
			role            TEXT NOT NULL,
			content         TEXT NOT NULL,
			agent_used      TEXT,
			tokens_used     INTEGER NOT NULL DEFAULT 0,
			created_at      DATETIME NOT NULL,
			FOREIGN KEY (conversation_id) REFERENCES conversations(id)
		)`,
		`CREATE TABLE IF NOT EXISTS movement_records (
				id          TEXT PRIMARY KEY,
				member_id   TEXT NOT NULL,
				metric      TEXT,
				value       TEXT,
				unit        TEXT,
				note        TEXT DEFAULT '',
				record_date TEXT NOT NULL,
				created_at  DATETIME NOT NULL,
				updated_at  DATETIME NOT NULL,
				FOREIGN KEY (member_id) REFERENCES family_members(id)
			)`,
		`CREATE TABLE IF NOT EXISTS ai_config (
			id         INTEGER PRIMARY KEY DEFAULT 1,
			endpoint   TEXT NOT NULL DEFAULT 'https://api.openai.com/v1',
			model      TEXT NOT NULL DEFAULT 'gpt-4o-mini',
			api_key    TEXT NOT NULL DEFAULT '',
			max_tokens INTEGER NOT NULL DEFAULT 2048
		)`,
			`CREATE TABLE IF NOT EXISTS llm_usage (
				id                TEXT PRIMARY KEY,
				conversation_id   TEXT NOT NULL DEFAULT '',
				model             TEXT NOT NULL,
				lens_id           TEXT DEFAULT '',
				domain            TEXT NOT NULL DEFAULT '',
				prompt_tokens     INTEGER NOT NULL DEFAULT 0,
				completion_tokens INTEGER NOT NULL DEFAULT 0,
				total_tokens      INTEGER NOT NULL DEFAULT 0,
				cost_cents        INTEGER NOT NULL DEFAULT 0,
				latency_ms        INTEGER NOT NULL DEFAULT 0,
				created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
			)`,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, m := range migrations {
		if _, err := s.db.Exec(m); err != nil {
			return fmt.Errorf("exec migration: %w", err)
		}
	}

	// Migrate exercise records from health_records to movement_records
	s.migrateExerciseRecords()

	// Add lens columns to messages if missing
	s.db.Exec(`ALTER TABLE messages ADD COLUMN lens_id TEXT DEFAULT ''`)
	s.db.Exec(`ALTER TABLE messages ADD COLUMN lens_name TEXT DEFAULT ''`)
	s.db.Exec(`ALTER TABLE messages ADD COLUMN lens_reason TEXT DEFAULT ''`)

	return nil
}

func (s *Store) migrateExerciseRecords() {
	s.db.Exec(`INSERT OR IGNORE INTO movement_records (id, member_id, metric, value, unit, note, record_date, created_at, updated_at)
		SELECT id, member_id, metric, value, unit, note, record_date, created_at, updated_at
		FROM health_records WHERE type = 'exercise'`)
	s.db.Exec(`DELETE FROM health_records WHERE type = 'exercise' AND id IN (SELECT id FROM movement_records)`)
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

// Task CRUD

func (s *Store) ListTasks() ([]FamilyTask, error) {
	rows, err := s.db.Query(`SELECT id, title, description, assignee_member_id, due_at, status, source_inbox_item_id, created_at, updated_at FROM family_tasks ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []FamilyTask
	for rows.Next() {
		var t FamilyTask
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.AssigneeMemberID, &t.DueAt, &t.Status, &t.SourceInboxItemID, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (s *Store) GetTask(id string) (*FamilyTask, error) {
	var t FamilyTask
	err := s.db.QueryRow(`SELECT id, title, description, assignee_member_id, due_at, status, source_inbox_item_id, created_at, updated_at FROM family_tasks WHERE id = ?`, id).
		Scan(&t.ID, &t.Title, &t.Description, &t.AssigneeMemberID, &t.DueAt, &t.Status, &t.SourceInboxItemID, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *Store) UpdateTask(id, title, description string, dueAt *string, status, updatedAt string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`UPDATE family_tasks SET title=?, description=?, due_at=?, status=?, updated_at=? WHERE id=?`,
		title, description, dueAt, status, updatedAt, id)
	return err
}

func (s *Store) DeleteTask(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`DELETE FROM family_tasks WHERE id = ?`, id)
	return err
}

// Event CRUD

func (s *Store) ListEvents() ([]FamilyEvent, error) {
	rows, err := s.db.Query(`SELECT id, title, description, starts_at, ends_at, participant_members, source_inbox_item_id, created_at, updated_at FROM family_events ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []FamilyEvent
	for rows.Next() {
		var e FamilyEvent
		if err := rows.Scan(&e.ID, &e.Title, &e.Description, &e.StartsAt, &e.EndsAt, &e.ParticipantMembers, &e.SourceInboxItemID, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (s *Store) GetEvent(id string) (*FamilyEvent, error) {
	var e FamilyEvent
	err := s.db.QueryRow(`SELECT id, title, description, starts_at, ends_at, participant_members, source_inbox_item_id, created_at, updated_at FROM family_events WHERE id = ?`, id).
		Scan(&e.ID, &e.Title, &e.Description, &e.StartsAt, &e.EndsAt, &e.ParticipantMembers, &e.SourceInboxItemID, &e.CreatedAt, &e.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (s *Store) UpdateEvent(id, title, description string, startsAt, endsAt *string, updatedAt string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`UPDATE family_events SET title=?, description=?, starts_at=?, ends_at=?, updated_at=? WHERE id=?`,
		title, description, startsAt, endsAt, updatedAt, id)
	return err
}

func (s *Store) DeleteEvent(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`DELETE FROM family_events WHERE id = ?`, id)
	return err
}

// ShoppingItem CRUD

func (s *Store) ListShoppingItems() ([]ShoppingItem, error) {
	rows, err := s.db.Query(`SELECT id, name, quantity, status, source_inbox_item_id, created_at, updated_at FROM shopping_items ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ShoppingItem
	for rows.Next() {
		var i ShoppingItem
		if err := rows.Scan(&i.ID, &i.Name, &i.Quantity, &i.Status, &i.SourceInboxItemID, &i.CreatedAt, &i.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

func (s *Store) GetShoppingItem(id string) (*ShoppingItem, error) {
	var i ShoppingItem
	err := s.db.QueryRow(`SELECT id, name, quantity, status, source_inbox_item_id, created_at, updated_at FROM shopping_items WHERE id = ?`, id).
		Scan(&i.ID, &i.Name, &i.Quantity, &i.Status, &i.SourceInboxItemID, &i.CreatedAt, &i.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &i, nil
}

func (s *Store) UpdateShoppingItem(id, name, quantity, status, updatedAt string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`UPDATE shopping_items SET name=?, quantity=?, status=?, updated_at=? WHERE id=?`,
		name, quantity, status, updatedAt, id)
	return err
}

func (s *Store) DeleteShoppingItem(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`DELETE FROM shopping_items WHERE id = ?`, id)
	return err
}

// Note CRUD

func (s *Store) ListNotes() ([]FamilyNote, error) {
	rows, err := s.db.Query(`SELECT id, title, content, topic, source_inbox_item_id, created_at, updated_at FROM family_notes ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var notes []FamilyNote
	for rows.Next() {
		var n FamilyNote
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.Topic, &n.SourceInboxItemID, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

func (s *Store) GetNote(id string) (*FamilyNote, error) {
	var n FamilyNote
	err := s.db.QueryRow(`SELECT id, title, content, topic, source_inbox_item_id, created_at, updated_at FROM family_notes WHERE id = ?`, id).
		Scan(&n.ID, &n.Title, &n.Content, &n.Topic, &n.SourceInboxItemID, &n.CreatedAt, &n.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (s *Store) UpdateNote(id, title, content, topic, updatedAt string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`UPDATE family_notes SET title=?, content=?, topic=?, updated_at=? WHERE id=?`,
		title, content, topic, updatedAt, id)
	return err
}

func (s *Store) DeleteNote(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`DELETE FROM family_notes WHERE id = ?`, id)
	return err
}

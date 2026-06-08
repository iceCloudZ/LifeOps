package main

import (
	"database/sql"
	"time"
)

// FamilyMember

type FamilyMember struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Role      string  `json:"role"`
	BirthDate *string `json:"birth_date"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

func (s *Store) CreateFamilyMember(m *FamilyMember) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(
		`INSERT INTO family_members (id, name, role, birth_date, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		m.ID, m.Name, m.Role, m.BirthDate, now, now,
	)
	return err
}

func (s *Store) ListFamilyMembers() ([]FamilyMember, error) {
	rows, err := s.db.Query(
		`SELECT id, name, role, birth_date, created_at, updated_at FROM family_members ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []FamilyMember
	for rows.Next() {
		var m FamilyMember
		if err := rows.Scan(&m.ID, &m.Name, &m.Role, &m.BirthDate, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (s *Store) GetFamilyMember(id string) (*FamilyMember, error) {
	var m FamilyMember
	err := s.db.QueryRow(
		`SELECT id, name, role, birth_date, created_at, updated_at FROM family_members WHERE id = ?`, id,
	).Scan(&m.ID, &m.Name, &m.Role, &m.BirthDate, &m.CreatedAt, &m.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *Store) UpdateFamilyMember(id, name, role string, birthDate *string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(
		`UPDATE family_members SET name = ?, role = ?, birth_date = ?, updated_at = ? WHERE id = ?`,
		name, role, birthDate, time.Now().UTC().Format(time.RFC3339), id,
	)
	return err
}

func (s *Store) DeleteFamilyMember(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`DELETE FROM family_members WHERE id = ?`, id)
	return err
}

// FinanceAccount

type FinanceAccount struct {
	ID        string  `json:"id"`
	MemberID  *string `json:"member_id"`
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	Balance   float64 `json:"balance"`
	Note      string  `json:"note"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

func (s *Store) CreateFinanceAccount(a *FinanceAccount) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(
		`INSERT INTO finance_accounts (id, member_id, name, type, balance, note, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ID, a.MemberID, a.Name, a.Type, a.Balance, a.Note, now, now,
	)
	return err
}

func (s *Store) ListFinanceAccounts() ([]FinanceAccount, error) {
	rows, err := s.db.Query(
		`SELECT id, member_id, name, type, balance, note, created_at, updated_at FROM finance_accounts ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []FinanceAccount
	for rows.Next() {
		var a FinanceAccount
		if err := rows.Scan(&a.ID, &a.MemberID, &a.Name, &a.Type, &a.Balance, &a.Note, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}

func (s *Store) GetFinanceAccount(id string) (*FinanceAccount, error) {
	var a FinanceAccount
	err := s.db.QueryRow(
		`SELECT id, member_id, name, type, balance, note, created_at, updated_at FROM finance_accounts WHERE id = ?`, id,
	).Scan(&a.ID, &a.MemberID, &a.Name, &a.Type, &a.Balance, &a.Note, &a.CreatedAt, &a.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *Store) UpdateFinanceAccount(id string, memberID *string, name, accType, note string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(
		`UPDATE finance_accounts SET member_id = ?, name = ?, type = ?, note = ?, updated_at = ? WHERE id = ?`,
		memberID, name, accType, note, time.Now().UTC().Format(time.RFC3339), id,
	)
	return err
}

func (s *Store) UpdateFinanceAccountBalance(id string, balance float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(
		`UPDATE finance_accounts SET balance = ?, updated_at = ? WHERE id = ?`,
		balance, time.Now().UTC().Format(time.RFC3339), id,
	)
	return err
}

func (s *Store) DeleteFinanceAccount(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`DELETE FROM finance_accounts WHERE id = ?`, id)
	return err
}

// FinanceRecord

type FinanceRecord struct {
	ID         string  `json:"id"`
	MemberID   *string `json:"member_id"`
	Type       string  `json:"type"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
	Category   string  `json:"category"`
	Note       string  `json:"note"`
	RecordDate string  `json:"record_date"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

func (s *Store) CreateFinanceRecord(r *FinanceRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(
		`INSERT INTO finance_records (id, member_id, type, amount, currency, category, note, record_date, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.MemberID, r.Type, r.Amount, r.Currency, r.Category, r.Note, r.RecordDate, now, now,
	)
	return err
}

func (s *Store) ListFinanceRecords(memberID, recordType, category, fromDate, toDate string) ([]FinanceRecord, error) {
	query := `SELECT id, member_id, type, amount, currency, category, note, record_date, created_at, updated_at FROM finance_records`
	var args []interface{}
	var conditions []string

	if memberID != "" {
		conditions = append(conditions, `member_id = ?`)
		args = append(args, memberID)
	}
	if recordType != "" {
		conditions = append(conditions, `type = ?`)
		args = append(args, recordType)
	}
	if category != "" {
		conditions = append(conditions, `category = ?`)
		args = append(args, category)
	}
	if fromDate != "" {
		conditions = append(conditions, `record_date >= ?`)
		args = append(args, fromDate)
	}
	if toDate != "" {
		conditions = append(conditions, `record_date <= ?`)
		args = append(args, toDate)
	}
	if len(conditions) > 0 {
		query += ` WHERE ` + conditions[0]
		for _, c := range conditions[1:] {
			query += ` AND ` + c
		}
	}
	query += ` ORDER BY record_date DESC`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []FinanceRecord
	for rows.Next() {
		var r FinanceRecord
		if err := rows.Scan(&r.ID, &r.MemberID, &r.Type, &r.Amount, &r.Currency, &r.Category, &r.Note, &r.RecordDate, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

func (s *Store) GetFinanceRecord(id string) (*FinanceRecord, error) {
	var r FinanceRecord
	err := s.db.QueryRow(
		`SELECT id, member_id, type, amount, currency, category, note, record_date, created_at, updated_at FROM finance_records WHERE id = ?`, id,
	).Scan(&r.ID, &r.MemberID, &r.Type, &r.Amount, &r.Currency, &r.Category, &r.Note, &r.RecordDate, &r.CreatedAt, &r.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *Store) UpdateFinanceRecord(id string, memberID *string, recordType string, amount float64, currency, category, note, recordDate string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(
		`UPDATE finance_records SET member_id = ?, type = ?, amount = ?, currency = ?, category = ?, note = ?, record_date = ?, updated_at = ? WHERE id = ?`,
		memberID, recordType, amount, currency, category, note, recordDate, time.Now().UTC().Format(time.RFC3339), id,
	)
	return err
}

func (s *Store) DeleteFinanceRecord(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`DELETE FROM finance_records WHERE id = ?`, id)
	return err
}

// HealthProfile

type HealthProfile struct {
	ID        string `json:"id"`
	MemberID  string `json:"member_id"`
	Summary   string `json:"summary"`
	UpdatedAt string `json:"updated_at"`
}

func (s *Store) GetHealthProfile(memberID string) (*HealthProfile, error) {
	var h HealthProfile
	err := s.db.QueryRow(
		`SELECT id, member_id, summary, updated_at FROM health_profiles WHERE member_id = ?`, memberID,
	).Scan(&h.ID, &h.MemberID, &h.Summary, &h.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (s *Store) UpdateHealthProfile(memberID, summary string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(
		`INSERT INTO health_profiles (id, member_id, summary, updated_at)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(member_id) DO UPDATE SET summary = excluded.summary, updated_at = excluded.updated_at`,
		newID(), memberID, summary, now,
	)
	return err
}

func (s *Store) DeleteHealthProfile(memberID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`DELETE FROM health_profiles WHERE member_id = ?`, memberID)
	return err
}

// HealthRecord

type HealthRecord struct {
	ID         string  `json:"id"`
	MemberID   string  `json:"member_id"`
	Type       string  `json:"type"`
	Metric     *string `json:"metric"`
	Value      *string `json:"value"`
	Unit       *string `json:"unit"`
	Note       string  `json:"note"`
	RecordDate string  `json:"record_date"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

func (s *Store) CreateHealthRecord(r *HealthRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(
		`INSERT INTO health_records (id, member_id, type, metric, value, unit, note, record_date, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.MemberID, r.Type, r.Metric, r.Value, r.Unit, r.Note, r.RecordDate, now, now,
	)
	return err
}

func (s *Store) ListHealthRecords(memberID, recordType string) ([]HealthRecord, error) {
	query := `SELECT id, member_id, type, metric, value, unit, note, record_date, created_at, updated_at FROM health_records`
	var args []interface{}
	var conditions []string

	if memberID != "" {
		conditions = append(conditions, `member_id = ?`)
		args = append(args, memberID)
	}
	if recordType != "" {
		conditions = append(conditions, `type = ?`)
		args = append(args, recordType)
	}
	if len(conditions) > 0 {
		query += ` WHERE ` + conditions[0]
		for _, c := range conditions[1:] {
			query += ` AND ` + c
		}
	}
	query += ` ORDER BY record_date DESC`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []HealthRecord
	for rows.Next() {
		var r HealthRecord
		if err := rows.Scan(&r.ID, &r.MemberID, &r.Type, &r.Metric, &r.Value, &r.Unit, &r.Note, &r.RecordDate, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

func (s *Store) GetHealthRecord(id string) (*HealthRecord, error) {
	var r HealthRecord
	err := s.db.QueryRow(
		`SELECT id, member_id, type, metric, value, unit, note, record_date, created_at, updated_at FROM health_records WHERE id = ?`, id,
	).Scan(&r.ID, &r.MemberID, &r.Type, &r.Metric, &r.Value, &r.Unit, &r.Note, &r.RecordDate, &r.CreatedAt, &r.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *Store) UpdateHealthRecord(id string, memberID, recordType string, metric, value, unit *string, note, recordDate string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(
		`UPDATE health_records SET member_id = ?, type = ?, metric = ?, value = ?, unit = ?, note = ?, record_date = ?, updated_at = ? WHERE id = ?`,
		memberID, recordType, metric, value, unit, note, recordDate, time.Now().UTC().Format(time.RFC3339), id,
	)
	return err
}

func (s *Store) DeleteHealthRecord(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`DELETE FROM health_records WHERE id = ?`, id)
	return err
}

// WorkStatus

type WorkStatus struct {
	ID        string `json:"id"`
	MemberID  string `json:"member_id"`
	Summary   string `json:"summary"`
	UpdatedAt string `json:"updated_at"`
}

func (s *Store) GetWorkStatus(memberID string) (*WorkStatus, error) {
	var w WorkStatus
	err := s.db.QueryRow(
		`SELECT id, member_id, summary, updated_at FROM work_statuses WHERE member_id = ?`, memberID,
	).Scan(&w.ID, &w.MemberID, &w.Summary, &w.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (s *Store) UpdateWorkStatus(memberID, summary string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(
		`INSERT INTO work_statuses (id, member_id, summary, updated_at)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(member_id) DO UPDATE SET summary = excluded.summary, updated_at = excluded.updated_at`,
		newID(), memberID, summary, now,
	)
	return err
}

func (s *Store) DeleteWorkStatus(memberID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`DELETE FROM work_statuses WHERE member_id = ?`, memberID)
	return err
}

// WorkRecord

type WorkRecord struct {
	ID        string  `json:"id"`
	MemberID  string  `json:"member_id"`
	Type      string  `json:"type"`
	Title     string  `json:"title"`
	Status    string  `json:"status"`
	Priority  string  `json:"priority"`
	Project   string  `json:"project"`
	DueDate   *string `json:"due_date"`
	Note      string  `json:"note"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

func (s *Store) CreateWorkRecord(r *WorkRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(
		`INSERT INTO work_records (id, member_id, type, title, status, priority, project, due_date, note, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.MemberID, r.Type, r.Title, r.Status, r.Priority, r.Project, r.DueDate, r.Note, now, now,
	)
	return err
}

func (s *Store) ListWorkRecords(memberID, status string) ([]WorkRecord, error) {
	query := `SELECT id, member_id, type, title, status, priority, project, due_date, note, created_at, updated_at FROM work_records`
	var args []interface{}
	var conditions []string

	if memberID != "" {
		conditions = append(conditions, `member_id = ?`)
		args = append(args, memberID)
	}
	if status != "" {
		conditions = append(conditions, `status = ?`)
		args = append(args, status)
	}
	if len(conditions) > 0 {
		query += ` WHERE ` + conditions[0]
		for _, c := range conditions[1:] {
			query += ` AND ` + c
		}
	}
	query += ` ORDER BY created_at DESC`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []WorkRecord
	for rows.Next() {
		var r WorkRecord
		if err := rows.Scan(&r.ID, &r.MemberID, &r.Type, &r.Title, &r.Status, &r.Priority, &r.Project, &r.DueDate, &r.Note, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

func (s *Store) GetWorkRecord(id string) (*WorkRecord, error) {
	var r WorkRecord
	err := s.db.QueryRow(
		`SELECT id, member_id, type, title, status, priority, project, due_date, note, created_at, updated_at FROM work_records WHERE id = ?`, id,
	).Scan(&r.ID, &r.MemberID, &r.Type, &r.Title, &r.Status, &r.Priority, &r.Project, &r.DueDate, &r.Note, &r.CreatedAt, &r.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *Store) UpdateWorkRecord(id string, memberID, recordType, title, status, priority, project string, dueDate *string, note string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(
		`UPDATE work_records SET member_id = ?, type = ?, title = ?, status = ?, priority = ?, project = ?, due_date = ?, note = ?, updated_at = ? WHERE id = ?`,
		memberID, recordType, title, status, priority, project, dueDate, note, time.Now().UTC().Format(time.RFC3339), id,
	)
	return err
}

func (s *Store) DeleteWorkRecord(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`DELETE FROM work_records WHERE id = ?`, id)
	return err
}

// FamilyStatus

type FamilyStatus struct {
	ID        string `json:"id"`
	Summary   string `json:"summary"`
	UpdatedAt string `json:"updated_at"`
}

func (s *Store) GetFamilyStatus() (*FamilyStatus, error) {
	var f FamilyStatus
	err := s.db.QueryRow(
		`SELECT id, summary, updated_at FROM family_statuses LIMIT 1`,
	).Scan(&f.ID, &f.Summary, &f.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (s *Store) UpdateFamilyStatus(summary string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(
		`INSERT INTO family_statuses (id, summary, updated_at)
		 VALUES (?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET summary = excluded.summary, updated_at = excluded.updated_at`,
		"family", summary, now,
	)
	return err
}

// FamilyRecord

type FamilyRecord struct {
	ID             string  `json:"id"`
	MemberID       *string `json:"member_id"`
	Type           string  `json:"type"`
	Title          string  `json:"title"`
	Status         string  `json:"status"`
	Location       string  `json:"location"`
	Participants   string  `json:"participants"`
	ScheduledDate  *string `json:"scheduled_date"`
	Note           string  `json:"note"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

func (s *Store) CreateFamilyRecord(r *FamilyRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(
		`INSERT INTO family_records (id, member_id, type, title, status, location, participants, scheduled_date, note, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.MemberID, r.Type, r.Title, r.Status, r.Location, r.Participants, r.ScheduledDate, r.Note, now, now,
	)
	return err
}

func (s *Store) ListFamilyRecords(memberID, recordType, status string) ([]FamilyRecord, error) {
	query := `SELECT id, member_id, type, title, status, location, participants, scheduled_date, note, created_at, updated_at FROM family_records`
	var args []interface{}
	var conditions []string

	if memberID != "" {
		conditions = append(conditions, `member_id = ?`)
		args = append(args, memberID)
	}
	if recordType != "" {
		conditions = append(conditions, `type = ?`)
		args = append(args, recordType)
	}
	if status != "" {
		conditions = append(conditions, `status = ?`)
		args = append(args, status)
	}
	if len(conditions) > 0 {
		query += ` WHERE ` + conditions[0]
		for _, c := range conditions[1:] {
			query += ` AND ` + c
		}
	}
	query += ` ORDER BY created_at DESC`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []FamilyRecord
	for rows.Next() {
		var r FamilyRecord
		if err := rows.Scan(&r.ID, &r.MemberID, &r.Type, &r.Title, &r.Status, &r.Location, &r.Participants, &r.ScheduledDate, &r.Note, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

func (s *Store) GetFamilyRecord(id string) (*FamilyRecord, error) {
	var r FamilyRecord
	err := s.db.QueryRow(
		`SELECT id, member_id, type, title, status, location, participants, scheduled_date, note, created_at, updated_at FROM family_records WHERE id = ?`, id,
	).Scan(&r.ID, &r.MemberID, &r.Type, &r.Title, &r.Status, &r.Location, &r.Participants, &r.ScheduledDate, &r.Note, &r.CreatedAt, &r.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *Store) UpdateFamilyRecord(id string, memberID *string, recordType, title, status, location, participants string, scheduledDate *string, note string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(
		`UPDATE family_records SET member_id = ?, type = ?, title = ?, status = ?, location = ?, participants = ?, scheduled_date = ?, note = ?, updated_at = ? WHERE id = ?`,
		memberID, recordType, title, status, location, participants, scheduledDate, note, time.Now().UTC().Format(time.RFC3339), id,
	)
	return err
}

func (s *Store) DeleteFamilyRecord(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`DELETE FROM family_records WHERE id = ?`, id)
	return err
}

// KnowledgeNote

type KnowledgeNote struct {
	ID        string  `json:"id"`
	Domain    *string `json:"domain"`
	MemberID  *string `json:"member_id"`
	Title     string  `json:"title"`
	Content   string  `json:"content"`
	Tags      string  `json:"tags"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

func (s *Store) CreateKnowledgeNote(n *KnowledgeNote) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(
		`INSERT INTO knowledge_notes (id, domain, member_id, title, content, tags, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		n.ID, n.Domain, n.MemberID, n.Title, n.Content, n.Tags, now, now,
	)
	return err
}

func (s *Store) ListKnowledgeNotes(domain, memberID string) ([]KnowledgeNote, error) {
	query := `SELECT id, domain, member_id, title, content, tags, created_at, updated_at FROM knowledge_notes`
	var args []interface{}
	var conditions []string

	if domain != "" {
		conditions = append(conditions, `domain = ?`)
		args = append(args, domain)
	}
	if memberID != "" {
		conditions = append(conditions, `member_id = ?`)
		args = append(args, memberID)
	}
	if len(conditions) > 0 {
		query += ` WHERE ` + conditions[0]
		for _, c := range conditions[1:] {
			query += ` AND ` + c
		}
	}
	query += ` ORDER BY created_at DESC`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []KnowledgeNote
	for rows.Next() {
		var n KnowledgeNote
		if err := rows.Scan(&n.ID, &n.Domain, &n.MemberID, &n.Title, &n.Content, &n.Tags, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

func (s *Store) GetKnowledgeNote(id string) (*KnowledgeNote, error) {
	var n KnowledgeNote
	err := s.db.QueryRow(
		`SELECT id, domain, member_id, title, content, tags, created_at, updated_at FROM knowledge_notes WHERE id = ?`, id,
	).Scan(&n.ID, &n.Domain, &n.MemberID, &n.Title, &n.Content, &n.Tags, &n.CreatedAt, &n.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (s *Store) UpdateKnowledgeNote(id string, domain, memberID *string, title, content, tags string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(
		`UPDATE knowledge_notes SET domain = ?, member_id = ?, title = ?, content = ?, tags = ?, updated_at = ? WHERE id = ?`,
		domain, memberID, title, content, tags, time.Now().UTC().Format(time.RFC3339), id,
	)
	return err
}

func (s *Store) DeleteKnowledgeNote(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`DELETE FROM knowledge_notes WHERE id = ?`, id)
	return err
}

// AI Config

type AIConfig struct {
	Endpoint  string `json:"endpoint"`
	Model     string `json:"model"`
	APIKey    string `json:"api_key"`
	MaxTokens int    `json:"max_tokens"`
}

func (s *Store) GetAIConfig() (*AIConfig, error) {
	var c AIConfig
	err := s.db.QueryRow(
		`SELECT endpoint, model, api_key, max_tokens FROM ai_config WHERE id = 1`,
	).Scan(&c.Endpoint, &c.Model, &c.APIKey, &c.MaxTokens)
	if err == sql.ErrNoRows {
		return &AIConfig{
			Endpoint:  "https://api.openai.com/v1",
			Model:     "gpt-4o-mini",
			MaxTokens: 2048,
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Store) UpdateAIConfig(c *AIConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(
		`INSERT INTO ai_config (id, endpoint, model, api_key, max_tokens)
		 VALUES (1, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET endpoint = excluded.endpoint, model = excluded.model, api_key = excluded.api_key, max_tokens = excluded.max_tokens`,
		c.Endpoint, c.Model, c.APIKey, c.MaxTokens,
	)
	return err
}

func (s *Store) ListHealthProfiles() ([]HealthProfile, error) {
	rows, err := s.db.Query(
		`SELECT id, member_id, summary, updated_at FROM health_profiles ORDER BY updated_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var profiles []HealthProfile
	for rows.Next() {
		var h HealthProfile
		if err := rows.Scan(&h.ID, &h.MemberID, &h.Summary, &h.UpdatedAt); err != nil {
			return nil, err
		}
		profiles = append(profiles, h)
	}
	return profiles, rows.Err()
}

func (s *Store) ListWorkStatuses() ([]WorkStatus, error) {
	rows, err := s.db.Query(
		`SELECT id, member_id, summary, updated_at FROM work_statuses ORDER BY updated_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var statuses []WorkStatus
	for rows.Next() {
		var w WorkStatus
		if err := rows.Scan(&w.ID, &w.MemberID, &w.Summary, &w.UpdatedAt); err != nil {
			return nil, err
		}
		statuses = append(statuses, w)
	}
	return statuses, rows.Err()
}

// Conversation

type Conversation struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (s *Store) CreateConversation(c *Conversation) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(
		`INSERT INTO conversations (id, title, created_at, updated_at)
		 VALUES (?, ?, ?, ?)`,
		c.ID, c.Title, now, now,
	)
	return err
}

func (s *Store) ListConversations() ([]Conversation, error) {
	rows, err := s.db.Query(
		`SELECT id, title, created_at, updated_at FROM conversations ORDER BY updated_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []Conversation
	for rows.Next() {
		var c Conversation
		if err := rows.Scan(&c.ID, &c.Title, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		conversations = append(conversations, c)
	}
	return conversations, rows.Err()
}

func (s *Store) GetConversation(id string) (*Conversation, error) {
	var c Conversation
	err := s.db.QueryRow(
		`SELECT id, title, created_at, updated_at FROM conversations WHERE id = ?`, id,
	).Scan(&c.ID, &c.Title, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Store) UpdateConversation(id, title string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(
		`UPDATE conversations SET title = ?, updated_at = ? WHERE id = ?`,
		title, time.Now().UTC().Format(time.RFC3339), id,
	)
	return err
}

func (s *Store) DeleteConversation(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`DELETE FROM conversations WHERE id = ?`, id)
	return err
}

// Message

type Message struct {
	ID             string  `json:"id"`
	ConversationID string  `json:"conversation_id"`
	Role           string  `json:"role"`
	Content        string  `json:"content"`
	AgentUsed      *string `json:"agent_used"`
	TokensUsed     int     `json:"tokens_used"`
	CreatedAt      string  `json:"created_at"`
}

func (s *Store) CreateMessage(m *Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(
		`INSERT INTO messages (id, conversation_id, role, content, agent_used, tokens_used, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		m.ID, m.ConversationID, m.Role, m.Content, m.AgentUsed, m.TokensUsed,
		time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

func (s *Store) ListMessages(conversationID string) ([]Message, error) {
	query := `SELECT id, conversation_id, role, content, agent_used, tokens_used, created_at FROM messages`
	var args []interface{}

	if conversationID != "" {
		query += ` WHERE conversation_id = ?`
		args = append(args, conversationID)
	}
	query += ` ORDER BY created_at ASC`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.Role, &m.Content, &m.AgentUsed, &m.TokensUsed, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}

func (s *Store) GetMessage(id string) (*Message, error) {
	var m Message
	err := s.db.QueryRow(
		`SELECT id, conversation_id, role, content, agent_used, tokens_used, created_at FROM messages WHERE id = ?`, id,
	).Scan(&m.ID, &m.ConversationID, &m.Role, &m.Content, &m.AgentUsed, &m.TokensUsed, &m.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *Store) DeleteMessage(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`DELETE FROM messages WHERE id = ?`, id)
	return err
}

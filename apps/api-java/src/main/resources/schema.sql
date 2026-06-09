-- LifeOps initial schema

CREATE TABLE IF NOT EXISTS inbox_items (
    id         TEXT PRIMARY KEY,
    source     TEXT NOT NULL,
    sender     TEXT DEFAULT '',
    content    TEXT NOT NULL,
    status     TEXT NOT NULL DEFAULT 'new',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS drafts (
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
);

CREATE TABLE IF NOT EXISTS family_tasks (
    id                  TEXT PRIMARY KEY,
    title               TEXT NOT NULL,
    description         TEXT DEFAULT '',
    assignee_member_id  TEXT,
    due_at              TEXT,
    status              TEXT NOT NULL DEFAULT 'open',
    source_inbox_item_id TEXT,
    created_at          DATETIME NOT NULL,
    updated_at          DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS family_events (
    id                   TEXT PRIMARY KEY,
    title                TEXT NOT NULL,
    description          TEXT DEFAULT '',
    starts_at            TEXT,
    ends_at              TEXT,
    participant_members  TEXT DEFAULT '',
    source_inbox_item_id TEXT,
    created_at           DATETIME NOT NULL,
    updated_at           DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS shopping_items (
    id                   TEXT PRIMARY KEY,
    name                 TEXT NOT NULL,
    quantity             TEXT DEFAULT '',
    status               TEXT NOT NULL DEFAULT 'open',
    source_inbox_item_id TEXT,
    created_at           DATETIME NOT NULL,
    updated_at           DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS family_notes (
    id                   TEXT PRIMARY KEY,
    title                TEXT NOT NULL,
    content              TEXT DEFAULT '',
    topic                TEXT DEFAULT '',
    source_inbox_item_id TEXT,
    created_at           DATETIME NOT NULL,
    updated_at           DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS family_members (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    role       TEXT NOT NULL,
    birth_date TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS finance_accounts (
    id         TEXT PRIMARY KEY,
    member_id  TEXT,
    name       TEXT NOT NULL,
    type       TEXT NOT NULL,
    balance    REAL NOT NULL DEFAULT 0,
    note       TEXT DEFAULT '',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (member_id) REFERENCES family_members(id)
);

CREATE TABLE IF NOT EXISTS finance_records (
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
);

CREATE TABLE IF NOT EXISTS health_profiles (
    id         TEXT PRIMARY KEY,
    member_id  TEXT NOT NULL UNIQUE,
    summary    TEXT NOT NULL DEFAULT '',
    updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS health_records (
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
);

CREATE TABLE IF NOT EXISTS work_statuses (
    id         TEXT PRIMARY KEY,
    member_id  TEXT NOT NULL UNIQUE,
    summary    TEXT NOT NULL DEFAULT '',
    updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS work_profiles (
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
);

CREATE TABLE IF NOT EXISTS work_records (
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
);

CREATE TABLE IF NOT EXISTS family_statuses (
    id         TEXT PRIMARY KEY,
    summary    TEXT NOT NULL DEFAULT '',
    updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS family_records (
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
);

CREATE TABLE IF NOT EXISTS knowledge_notes (
    id         TEXT PRIMARY KEY,
    domain     TEXT,
    member_id  TEXT,
    title      TEXT NOT NULL,
    content    TEXT NOT NULL,
    tags       TEXT DEFAULT '[]',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (member_id) REFERENCES family_members(id)
);

CREATE TABLE IF NOT EXISTS conversations (
    id         TEXT PRIMARY KEY,
    title      TEXT DEFAULT '',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS messages (
    id              TEXT PRIMARY KEY,
    conversation_id TEXT NOT NULL,
    role            TEXT NOT NULL,
    content         TEXT NOT NULL,
    agent_used      TEXT,
    tokens_used     INTEGER NOT NULL DEFAULT 0,
    lens_id         TEXT DEFAULT '',
    lens_name       TEXT DEFAULT '',
    lens_reason     TEXT DEFAULT '',
    created_at      DATETIME NOT NULL,
    FOREIGN KEY (conversation_id) REFERENCES conversations(id)
);

CREATE TABLE IF NOT EXISTS movement_records (
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
);

CREATE TABLE IF NOT EXISTS ai_config (
    id         INTEGER PRIMARY KEY DEFAULT 1,
    endpoint   TEXT NOT NULL DEFAULT 'https://api.openai.com/v1',
    model      TEXT NOT NULL DEFAULT 'gpt-4o-mini',
    api_key    TEXT NOT NULL DEFAULT '',
    max_tokens INTEGER NOT NULL DEFAULT 2048
);

CREATE TABLE IF NOT EXISTS llm_usage (
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
);

CREATE TABLE IF NOT EXISTS t_chat_trace (
    id                     INTEGER PRIMARY KEY AUTOINCREMENT,
    conversation_id        TEXT NOT NULL,
    trace_no               INTEGER NOT NULL DEFAULT 1,
    input_message          TEXT NOT NULL DEFAULT '',
    output_message         TEXT DEFAULT '',
    model                  TEXT DEFAULT '',
    lens_id                TEXT DEFAULT '',
    total_prompt_tokens    INTEGER NOT NULL DEFAULT 0,
    total_completion_tokens INTEGER NOT NULL DEFAULT 0,
    total_tokens           INTEGER NOT NULL DEFAULT 0,
    total_cached_tokens    INTEGER NOT NULL DEFAULT 0,
    cost_yuan              DECIMAL(10,6) NOT NULL DEFAULT 0,
    total_latency_ms       INTEGER NOT NULL DEFAULT 0,
    llm_call_count         INTEGER NOT NULL DEFAULT 0,
    tool_call_count        INTEGER NOT NULL DEFAULT 0,
    status                 TEXT NOT NULL DEFAULT 'running',
    error_message          TEXT DEFAULT '',
    metadata               TEXT DEFAULT '',
    created_at             DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS t_chat_span (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    trace_id        INTEGER NOT NULL,
    span_no         INTEGER NOT NULL DEFAULT 0,
    span_type       TEXT NOT NULL DEFAULT '',
    span_name       TEXT NOT NULL DEFAULT '',
    parent_span_id  INTEGER,
    input_data      TEXT DEFAULT '',
    output_data     TEXT DEFAULT '',
    prompt_tokens   INTEGER NOT NULL DEFAULT 0,
    completion_tokens INTEGER NOT NULL DEFAULT 0,
    cached_tokens   INTEGER NOT NULL DEFAULT 0,
    latency_ms      INTEGER NOT NULL DEFAULT 0,
    status          TEXT NOT NULL DEFAULT 'ok',
    metadata        TEXT DEFAULT '',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (trace_id) REFERENCES t_chat_trace(id)
);

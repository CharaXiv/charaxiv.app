-- D1 schema for write buffer
-- This schema implements the write coalescing pattern:
-- Writes are buffered in D1, flushed to R2 on read

-- Write buffer table
CREATE TABLE IF NOT EXISTS write_buffer (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id TEXT NOT NULL,
    path TEXT NOT NULL,
    value TEXT NOT NULL,
    created_at TEXT DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_buffer_char ON write_buffer(character_id);

-- Character metadata (for tracking last flush, etc.)
CREATE TABLE IF NOT EXISTS character_meta (
    character_id TEXT PRIMARY KEY,
    r2_key TEXT,
    last_flush TEXT,
    created_at TEXT DEFAULT (datetime('now'))
);

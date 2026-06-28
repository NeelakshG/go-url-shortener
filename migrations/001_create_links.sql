-- Migration: 001_create_links
-- Creates the links table for Phase 1.
-- click_count lives here for Phase 1; Phase 2 will move click recording to a separate clicks table.

CREATE TABLE IF NOT EXISTS links (
    id          BIGSERIAL PRIMARY KEY,
    short_code  TEXT        NOT NULL UNIQUE,
    long_url    TEXT        NOT NULL,
    click_count BIGINT      NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_links_short_code ON links (short_code);

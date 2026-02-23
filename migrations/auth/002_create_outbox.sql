-- +goose Up
CREATE TABLE outbox (
    event_id   UUID PRIMARY KEY,
    topic      TEXT NOT NULL,
    key        BYTEA,
    value      BYTEA NOT NULL,
    is_sent    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_outbox_unsent ON outbox (created_at) WHERE NOT is_sent;

-- +goose Down
DROP TABLE IF EXISTS outbox;

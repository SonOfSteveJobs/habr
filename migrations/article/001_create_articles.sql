-- +goose Up
CREATE TABLE articles (
    id         UUID PRIMARY KEY,
    author_id  UUID NOT NULL,
    title      TEXT NOT NULL,
    content    TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_articles_created_at_id ON articles (created_at DESC, id DESC);

-- +goose Down
DROP TABLE IF EXISTS articles;

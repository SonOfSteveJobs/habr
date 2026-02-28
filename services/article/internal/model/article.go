package model

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	titleMinLen   = 1
	titleMaxLen   = 255
	contentMinLen = 1
	contentMaxLen = 1000
)

type Article struct {
	ID        uuid.UUID
	AuthorID  uuid.UUID
	Title     string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewArticle(authorID uuid.UUID, title, content string) (*Article, error) {
	titleLen := utf8.RuneCountInString(title)
	if titleLen < titleMinLen || titleLen > titleMaxLen {
		return nil, ErrInvalidTitle
	}

	contentLen := utf8.RuneCountInString(content)
	if contentLen < contentMinLen || contentLen > contentMaxLen {
		return nil, ErrInvalidContent
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	return &Article{
		ID:       id,
		AuthorID: authorID,
		Title:    title,
		Content:  content,
	}, nil
}

type ArticlePage struct {
	Articles   []*Article
	NextCursor string
}

func EncodeCursor(createdAt time.Time, id uuid.UUID) string {
	raw := fmt.Sprintf("%d:%s", createdAt.UnixMicro(), id.String())
	return base64.URLEncoding.EncodeToString([]byte(raw))
}

func DecodeCursor(cursor string) (time.Time, uuid.UUID, error) {
	data, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, uuid.Nil, ErrInvalidCursor
	}

	parts := strings.SplitN(string(data), ":", 2)
	if len(parts) != 2 {
		return time.Time{}, uuid.Nil, ErrInvalidCursor
	}

	var micros int64
	if _, err := fmt.Sscanf(parts[0], "%d", &micros); err != nil {
		return time.Time{}, uuid.Nil, ErrInvalidCursor
	}

	id, err := uuid.Parse(parts[1])
	if err != nil {
		return time.Time{}, uuid.Nil, ErrInvalidCursor
	}

	return time.UnixMicro(micros), id, nil
}

package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	titleMinLen = 1
	titleMaxLen = 255
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
	if len(title) < titleMinLen || len(title) > titleMaxLen {
		return nil, ErrInvalidTitle
	}

	if len(content) < 1 {
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

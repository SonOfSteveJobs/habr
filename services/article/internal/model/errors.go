package model

import "errors"

var (
	ErrArticleNotFound = errors.New("article not found")
	ErrInvalidTitle    = errors.New("invalid title")
	ErrInvalidContent  = errors.New("invalid content")
	ErrInvalidCursor   = errors.New("invalid cursor")
	ErrNotAuthor       = errors.New("not the author")
)

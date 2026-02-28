package model

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewArticle_Success(t *testing.T) {
	authorID := uuid.Must(uuid.NewV7())

	article, err := NewArticle(authorID, "Title", "Content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if article.ID == uuid.Nil {
		t.Error("article.ID = uuid.Nil, want valid UUID")
	}

	if article.AuthorID != authorID {
		t.Errorf("article.AuthorID = %v, want %v", article.AuthorID, authorID)
	}

	if article.Title != "Title" {
		t.Errorf("article.Title = %q, want %q", article.Title, "Title")
	}

	if article.Content != "Content" {
		t.Errorf("article.Content = %q, want %q", article.Content, "Content")
	}
}

func TestNewArticle_TitleValidation(t *testing.T) {
	authorID := uuid.Must(uuid.NewV7())

	tests := []struct {
		name  string
		title string
		ok    bool
	}{
		{"empty", "", false},
		{"one char", "a", true},
		{"max length", strings.Repeat("a", 255), true},
		{"too long", strings.Repeat("a", 256), false},
		{"max utf8 runes", strings.Repeat("я", 255), true},
		{"too many utf8 runes", strings.Repeat("я", 256), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewArticle(authorID, tt.title, "content")
			if tt.ok && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tt.ok && !errors.Is(err, ErrInvalidTitle) {
				t.Errorf("error = %v, want ErrInvalidTitle", err)
			}
		})
	}
}

func TestNewArticle_ContentValidation(t *testing.T) {
	authorID := uuid.Must(uuid.NewV7())

	tests := []struct {
		name    string
		content string
		ok      bool
	}{
		{"empty", "", false},
		{"one char", "a", true},
		{"max length", strings.Repeat("a", 1000), true},
		{"too long", strings.Repeat("a", 1001), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewArticle(authorID, "title", tt.content)
			if tt.ok && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tt.ok && !errors.Is(err, ErrInvalidContent) {
				t.Errorf("error = %v, want ErrInvalidContent", err)
			}
		})
	}
}

func TestUpdate_BothFields(t *testing.T) {
	a := &Article{}
	id := uuid.Must(uuid.NewV7())
	authorID := uuid.Must(uuid.NewV7())
	title := "New Title"
	content := "New Content"

	err := a.Update(id, authorID, &title, &content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if a.ID != id {
		t.Errorf("ID = %v, want %v", a.ID, id)
	}
	if a.AuthorID != authorID {
		t.Errorf("AuthorID = %v, want %v", a.AuthorID, authorID)
	}
	if a.Title != title {
		t.Errorf("Title = %q, want %q", a.Title, title)
	}
	if a.Content != content {
		t.Errorf("Content = %q, want %q", a.Content, content)
	}
}

func TestUpdate_TitleOnly(t *testing.T) {
	a := &Article{}
	title := "Only Title"

	err := a.Update(uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), &title, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if a.Title != title {
		t.Errorf("Title = %q, want %q", a.Title, title)
	}
	if a.Content != "" {
		t.Errorf("Content = %q, want empty", a.Content)
	}
}

func TestUpdate_ContentOnly(t *testing.T) {
	a := &Article{}
	content := "Only Content"

	err := a.Update(uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), nil, &content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if a.Title != "" {
		t.Errorf("Title = %q, want empty", a.Title)
	}
	if a.Content != content {
		t.Errorf("Content = %q, want %q", a.Content, content)
	}
}

func TestUpdate_NilBoth(t *testing.T) {
	a := &Article{}
	id := uuid.Must(uuid.NewV7())
	authorID := uuid.Must(uuid.NewV7())

	err := a.Update(id, authorID, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if a.ID != id {
		t.Errorf("ID = %v, want %v", a.ID, id)
	}
	if a.AuthorID != authorID {
		t.Errorf("AuthorID = %v, want %v", a.AuthorID, authorID)
	}
}

func TestUpdate_TitleValidation(t *testing.T) {
	tests := []struct {
		name  string
		title string
		ok    bool
	}{
		{"empty", "", false},
		{"valid", "ok", true},
		{"too long", strings.Repeat("a", 256), false},
		{"max utf8", strings.Repeat("я", 255), true},
		{"too many utf8", strings.Repeat("я", 256), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Article{}
			err := a.Update(uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), &tt.title, nil)
			if tt.ok && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tt.ok && !errors.Is(err, ErrInvalidTitle) {
				t.Errorf("error = %v, want ErrInvalidTitle", err)
			}
		})
	}
}

func TestUpdate_ContentValidation(t *testing.T) {
	tests := []struct {
		name    string
		content string
		ok      bool
	}{
		{"empty", "", false},
		{"valid", "ok", true},
		{"too long", strings.Repeat("a", 1001), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Article{}
			err := a.Update(uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), nil, &tt.content)
			if tt.ok && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tt.ok && !errors.Is(err, ErrInvalidContent) {
				t.Errorf("error = %v, want ErrInvalidContent", err)
			}
		})
	}
}

func TestCursor_RoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Microsecond)
	id := uuid.Must(uuid.NewV7())

	cursor := EncodeCursor(now, id)

	gotTime, gotID, err := DecodeCursor(cursor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !gotTime.Equal(now) {
		t.Errorf("time = %v, want %v", gotTime, now)
	}

	if gotID != id {
		t.Errorf("id = %v, want %v", gotID, id)
	}
}

func TestDecodeCursor_InvalidBase64(t *testing.T) {
	_, _, err := DecodeCursor("not-valid-base64!!!")
	if !errors.Is(err, ErrInvalidCursor) {
		t.Errorf("error = %v, want ErrInvalidCursor", err)
	}
}

func TestDecodeCursor_MissingColon(t *testing.T) {
	cursor := base64.URLEncoding.EncodeToString([]byte("no-colon-here"))
	_, _, err := DecodeCursor(cursor)
	if !errors.Is(err, ErrInvalidCursor) {
		t.Errorf("error = %v, want ErrInvalidCursor", err)
	}
}

func TestDecodeCursor_InvalidTimestamp(t *testing.T) {
	cursor := base64.URLEncoding.EncodeToString([]byte("abc:" + uuid.Must(uuid.NewV7()).String()))
	_, _, err := DecodeCursor(cursor)
	if !errors.Is(err, ErrInvalidCursor) {
		t.Errorf("error = %v, want ErrInvalidCursor", err)
	}
}

func TestDecodeCursor_InvalidUUID(t *testing.T) {
	cursor := base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%d:not-a-uuid", time.Now().UnixMicro())))
	_, _, err := DecodeCursor(cursor)
	if !errors.Is(err, ErrInvalidCursor) {
		t.Errorf("error = %v, want ErrInvalidCursor", err)
	}
}

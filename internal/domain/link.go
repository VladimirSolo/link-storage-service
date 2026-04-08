package domain

import (
	"errors"
	"time"
)

type Link struct {
	ID          int64     `json:"id"`
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
	Visits      int64     `json:"visits"`
}

var (
	ErrNotFound = errors.New("link not found")
	ErrConflict = errors.New("short code already exists")
)

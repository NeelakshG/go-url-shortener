package model

import "time"

// Link represents a shortened URL stored in the database.
type Link struct {
	ID        int64     `json:"id"`
	ShortCode string    `json:"short_code"`
	LongURL   string    `json:"long_url"`
	CreatedAt time.Time `json:"created_at"`
}

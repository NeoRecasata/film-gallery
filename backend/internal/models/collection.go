package models

import "time"

type Collection struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description"`
	CoverPhoto  *string   `json:"cover_photo"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Photos     []Photo `json:"photos,omitempty"`
	PhotoCount int     `json:"photo_count,omitempty"`
	CoverURL *string `json:"cover_url,omitempty"`
}

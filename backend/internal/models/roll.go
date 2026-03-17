package models

import "time"

type Roll struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	Slug         string     `json:"slug"`
	Description  *string    `json:"description"`
	Camera       *string    `json:"camera"`
	FilmStock    *string    `json:"film_stock"`
	Lens         *string    `json:"lens"`
	Location     *string    `json:"location"`
	ShotAt       *time.Time `json:"shot_at"`
	Published    bool       `json:"published"`
	CoverPhotoID *string    `json:"cover_photo_id"`
	SortOrder    int        `json:"sort_order"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// Populated at API level
	Photos     []Photo `json:"photos,omitempty"`
	PhotoCount int     `json:"photo_count,omitempty"`
	CoverURL   *string `json:"cover_url,omitempty"`
}

package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type PhotoVariants map[string]string

func (v PhotoVariants) Value() (driver.Value, error) {
	return json.Marshal(v)
}

func (v *PhotoVariants) Scan(src interface{}) error {
	if src == nil {
		*v = make(PhotoVariants)
		return nil
	}
	var data []byte
	switch val := src.(type) {
	case []byte:
		data = val
	case string:
		data = []byte(val)
	default:
		return fmt.Errorf("unsupported type for PhotoVariants: %T", src)
	}
	return json.Unmarshal(data, v)
}

type Photo struct {
	ID          string        `json:"id"`
	Title       *string       `json:"title"`
	Description *string       `json:"description"`
	Slug        string        `json:"slug"`
	FilmStock   *string       `json:"film_stock"`
	Camera      *string       `json:"camera"`
	Lens        *string       `json:"lens"`
	Location    *string       `json:"location"`
	TakenAt     *time.Time    `json:"taken_at"`
	RollID      string        `json:"roll_id"`
	Hidden      bool          `json:"hidden"`
	OriginalKey string        `json:"-"`
	Variants    PhotoVariants `json:"-"`
	Width       int           `json:"width"`
	Height      int           `json:"height"`
	FileSize    int64         `json:"file_size"`
	BlurHash    *string       `json:"blur_hash"`
	SortOrder   int           `json:"sort_order"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`

	// Populated at API level
	URLs      map[string]string `json:"urls,omitempty"`
	RollSlug  string            `json:"roll_slug,omitempty"`
	RollTitle string            `json:"roll_title,omitempty"`
	PrevSlug  *string           `json:"prev_slug,omitempty"`
	NextSlug  *string           `json:"next_slug,omitempty"`
}

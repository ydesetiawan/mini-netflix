package model

import "time"

type Content struct {
	ID             string    `json:"id" db:"id"`
	Title          string    `json:"title" db:"title"`
	Description    string    `json:"description" db:"description"`
	Type           string    `json:"type" db:"type"` // movie | series
	ReleaseYear    int       `json:"release_year" db:"release_year"`
	DurationMin    int       `json:"duration_min,omitempty" db:"duration_min"`
	TotalEpisodes  int       `json:"total_episodes,omitempty" db:"total_episodes"`
	ThumbnailURL   string    `json:"thumbnail_url" db:"thumbnail_url"`
	VideoURL       string    `json:"video_url" db:"video_url"`
	ViewCount      int64     `json:"view_count" db:"view_count"`
	RatingAvg      float64   `json:"rating_avg" db:"rating_avg"`
	RatingCount    int       `json:"rating_count" db:"rating_count"`
	Genres         []string  `json:"genres,omitempty"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

type CreateContentRequest struct {
	Title         string   `json:"title" binding:"required"`
	Description   string   `json:"description"`
	Type          string   `json:"type" binding:"required,oneof=movie series"`
	ReleaseYear   int      `json:"release_year"`
	DurationMin   int      `json:"duration_min"`
	TotalEpisodes int      `json:"total_episodes"`
	ThumbnailURL  string   `json:"thumbnail_url"`
	VideoURL      string   `json:"video_url"`
	GenreIDs      []int    `json:"genre_ids"`
}

type WatchEvent struct {
	UserID    string `json:"user_id"`
	ContentID string `json:"content_id"`
	Progress  int    `json:"progress"` // seconds
}

type RateRequest struct {
	Score int `json:"score" binding:"required,min=1,max=10"`
}

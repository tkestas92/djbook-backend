package model

import "time"

type ReleasePlatform struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Enabled bool   `json:"enabled"`
}

type Release struct {
	ID          string
	ProfileID   string
	Title       string
	Artist      string
	ArtworkURL  *string
	SongLinkURL string
	Platforms   []ReleasePlatform
	CreatedAt   time.Time
}

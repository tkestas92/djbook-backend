package model

import "time"

type ProfilePhoto struct {
	ID        string
	ProfileID string
	URL       string
	SortOrder int
	CreatedAt time.Time
}

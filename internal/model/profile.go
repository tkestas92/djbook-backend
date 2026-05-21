package model

import "time"

type DjProfile struct {
	ID                string
	UserID            string
	DjName            string
	Bio               *string
	Email             *string
	ManagementContact *string
	CreatedAt         time.Time
}

type ProfileGenre struct {
	ID        string
	ProfileID string
	Genre     string
}

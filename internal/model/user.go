package model

import "time"

type User struct {
	ID           string
	GoogleID     *string
	AppleID      *string
	Email        string
	Username     *string
	PasswordHash *string
	CreatedAt    time.Time
}

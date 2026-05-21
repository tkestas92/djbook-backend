package model

import "time"

type EventStatus string
type PaymentStatus string

const (
	EventStatusPending   EventStatus = "PENDING"
	EventStatusConfirmed EventStatus = "CONFIRMED"
	EventStatusCompleted EventStatus = "COMPLETED"
	EventStatusCancelled EventStatus = "CANCELLED"
)

const (
	PaymentStatusPaid   PaymentStatus = "PAID"
	PaymentStatusUnpaid PaymentStatus = "UNPAID"
)

type Event struct {
	ID            string
	ProfileID     string
	Title         string
	Venue         string
	Date          time.Time
	StartTime     string
	EndTime       *string
	Notes         *string
	AmountEur     *float64
	IsPublic      bool
	TicketsURL    *string
	EventStatus   EventStatus
	PaymentStatus PaymentStatus
	CreatedAt     time.Time
}

type EventStatusHistory struct {
	ID        string
	EventID   string
	OldStatus *string
	NewStatus string
	ChangedAt time.Time
}

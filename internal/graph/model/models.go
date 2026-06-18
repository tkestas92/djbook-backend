package model

import (
	"fmt"
	"io"
	"strconv"
)

// EventStatus represents the status of a booking event.
type EventStatus string

const (
	EventStatusPending   EventStatus = "PENDING"
	EventStatusConfirmed EventStatus = "CONFIRMED"
	EventStatusCompleted EventStatus = "COMPLETED"
	EventStatusCancelled EventStatus = "CANCELLED"
)

var AllEventStatus = []EventStatus{
	EventStatusPending,
	EventStatusConfirmed,
	EventStatusCompleted,
	EventStatusCancelled,
}

func (e EventStatus) IsValid() bool {
	switch e {
	case EventStatusPending, EventStatusConfirmed, EventStatusCompleted, EventStatusCancelled:
		return true
	}
	return false
}

func (e EventStatus) String() string { return string(e) }

func (e EventStatus) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(string(e)))
}

func (e *EventStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}
	*e = EventStatus(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid EventStatus", str)
	}
	return nil
}

// PaymentStatus represents whether a gig has been paid.
type PaymentStatus string

const (
	PaymentStatusPaid   PaymentStatus = "PAID"
	PaymentStatusUnpaid PaymentStatus = "UNPAID"
)

var AllPaymentStatus = []PaymentStatus{
	PaymentStatusPaid,
	PaymentStatusUnpaid,
}

func (e PaymentStatus) IsValid() bool {
	switch e {
	case PaymentStatusPaid, PaymentStatusUnpaid:
		return true
	}
	return false
}

func (e PaymentStatus) String() string { return string(e) }

func (e PaymentStatus) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(string(e)))
}

func (e *PaymentStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}
	*e = PaymentStatus(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid PaymentStatus", str)
	}
	return nil
}

// Period is used for finance summary filtering.
type Period string

const (
	PeriodMonth   Period = "MONTH"
	PeriodQuarter Period = "QUARTER"
	PeriodYear    Period = "YEAR"
)

var AllPeriod = []Period{
	PeriodMonth,
	PeriodQuarter,
	PeriodYear,
}

func (e Period) IsValid() bool {
	switch e {
	case PeriodMonth, PeriodQuarter, PeriodYear:
		return true
	}
	return false
}

func (e Period) String() string { return string(e) }

func (e Period) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(string(e)))
}

func (e *Period) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}
	*e = Period(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid Period", str)
	}
	return nil
}

// Input types

type DjProfileInput struct {
	DjName            string   `json:"djName"`
	Bio               *string  `json:"bio,omitempty"`
	Email             *string  `json:"email,omitempty"`
	ManagementContact *string  `json:"managementContact,omitempty"`
	Genres            []string `json:"genres"`
}

type EventInput struct {
	Title         string        `json:"title"`
	Venue         string        `json:"venue"`
	Date          string        `json:"date"`
	StartTime     string        `json:"startTime"`
	EndTime       *string       `json:"endTime,omitempty"`
	Notes         *string       `json:"notes,omitempty"`
	AmountEur     *float64      `json:"amountEur,omitempty"`
	IsPublic      *bool         `json:"isPublic,omitempty"`
	TicketsURL    *string       `json:"ticketsUrl,omitempty"`
	EventStatus   EventStatus   `json:"eventStatus"`
	PaymentStatus PaymentStatus `json:"paymentStatus"`
}

type EventFilter struct {
	Month         *string        `json:"month,omitempty"`
	EventStatus   *EventStatus   `json:"eventStatus,omitempty"`
	PaymentStatus *PaymentStatus `json:"paymentStatus,omitempty"`
}

type SocialLinkInput struct {
	Platform string `json:"platform"`
	URL      string `json:"url"`
}

type ReleaseInput struct {
	Title       string                  `json:"title"`
	Artist      string                  `json:"artist"`
	ArtworkURL  *string                 `json:"artworkUrl,omitempty"`
	SongLinkURL string                  `json:"songLinkUrl"`
	Platforms   []*ReleasePlatformInput `json:"platforms"`
}

type ReleasePlatformInput struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Enabled bool   `json:"enabled"`
}

// Response types

type User struct {
	ID        string       `json:"id"`
	Email     string       `json:"email"`
	Profiles  []*DjProfile `json:"profiles"`
	CreatedAt string       `json:"createdAt"`
}

type DjProfile struct {
	ID                string          `json:"id"`
	DjName            string          `json:"djName"`
	Bio               *string         `json:"bio,omitempty"`
	Email             *string         `json:"email,omitempty"`
	ManagementContact *string         `json:"managementContact,omitempty"`
	Genres            []string        `json:"genres"`
	Photos            []*ProfilePhoto `json:"photos"`
	SocialLinks       []*SocialLink   `json:"socialLinks"`
	Events            []*Event        `json:"events"`
	UpcomingEvents    []*Event        `json:"upcomingEvents"`
	CreatedAt         string          `json:"createdAt"`
}

type Event struct {
	ID            string                `json:"id"`
	Title         string                `json:"title"`
	Venue         string                `json:"venue"`
	Date          string                `json:"date"`
	StartTime     string                `json:"startTime"`
	EndTime       *string               `json:"endTime,omitempty"`
	Notes         *string               `json:"notes,omitempty"`
	AmountEur     *float64              `json:"amountEur,omitempty"`
	IsPublic      bool                  `json:"isPublic"`
	TicketsURL    *string               `json:"ticketsUrl,omitempty"`
	EventStatus   EventStatus           `json:"eventStatus"`
	PaymentStatus PaymentStatus         `json:"paymentStatus"`
	StatusHistory []*EventStatusHistory `json:"statusHistory"`
	CreatedAt     string                `json:"createdAt"`
}

type EventStatusHistory struct {
	ID        string  `json:"id"`
	OldStatus *string `json:"oldStatus,omitempty"`
	NewStatus string  `json:"newStatus"`
	ChangedAt string  `json:"changedAt"`
}

type ProfilePhoto struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	SortOrder int    `json:"sortOrder"`
	CreatedAt string `json:"createdAt"`
}

type SocialLink struct {
	ID       string `json:"id"`
	Platform string `json:"platform"`
	URL      string `json:"url"`
}

type Release struct {
	ID          string             `json:"id"`
	Title       string             `json:"title"`
	Artist      string             `json:"artist"`
	ArtworkURL  *string            `json:"artworkUrl,omitempty"`
	SongLinkURL string             `json:"songLinkUrl"`
	Platforms   []*ReleasePlatform `json:"platforms"`
	CreatedAt   string             `json:"createdAt"`
}

type ReleasePlatform struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Enabled bool   `json:"enabled"`
}

type FinanceSummary struct {
	ThisMonthEarnings   float64          `json:"thisMonthEarnings"`
	ThisMonthEventCount int              `json:"thisMonthEventCount"`
	ThisYearEarnings    float64          `json:"thisYearEarnings"`
	ThisYearEventCount  int              `json:"thisYearEventCount"`
	UnpaidAmount        float64          `json:"unpaidAmount"`
	UnpaidEventCount    int              `json:"unpaidEventCount"`
	MonthlyChart        []*MonthlyEarning `json:"monthlyChart"`
	BestMonth           *MonthlyEarning  `json:"bestMonth,omitempty"`
}

type MonthlyEarning struct {
	Month      string  `json:"month"`
	Earnings   float64 `json:"earnings"`
	EventCount int     `json:"eventCount"`
}

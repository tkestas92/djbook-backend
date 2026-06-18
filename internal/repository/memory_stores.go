package repository

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/djbook/backend/internal/model"
)

var errNotFound = fmt.Errorf("not found")

// MemoryUserStore is an in-memory UserStore for tests.
type MemoryUserStore struct {
	mu         sync.Mutex
	byID       map[string]*model.User
	byUsername map[string]string
	byEmail    map[string]string
	byGoogleID map[string]string
	byAppleID  map[string]string
}

func NewMemoryUserStore() *MemoryUserStore {
	return &MemoryUserStore{
		byID:       make(map[string]*model.User),
		byUsername: make(map[string]string),
		byEmail:    make(map[string]string),
		byGoogleID: make(map[string]string),
		byAppleID:  make(map[string]string),
	}
}

func (s *MemoryUserStore) Create(_ context.Context, user *model.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	copyUser := *user
	s.byID[user.ID] = &copyUser
	s.byEmail[user.Email] = user.ID
	if user.Username != nil {
		s.byUsername[*user.Username] = user.ID
	}
	if user.GoogleID != nil {
		s.byGoogleID[*user.GoogleID] = user.ID
	}
	if user.AppleID != nil {
		s.byAppleID[*user.AppleID] = user.ID
	}
	return nil
}

func (s *MemoryUserStore) GetByID(_ context.Context, id string) (*model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.byID[id]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	copyUser := *user
	return &copyUser, nil
}

func (s *MemoryUserStore) GetByGoogleID(_ context.Context, googleID string) (*model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, ok := s.byGoogleID[googleID]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	copyUser := *s.byID[id]
	return &copyUser, nil
}

func (s *MemoryUserStore) GetByAppleID(_ context.Context, appleID string) (*model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, ok := s.byAppleID[appleID]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	copyUser := *s.byID[id]
	return &copyUser, nil
}

func (s *MemoryUserStore) GetByEmail(_ context.Context, email string) (*model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, ok := s.byEmail[email]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	copyUser := *s.byID[id]
	return &copyUser, nil
}

func (s *MemoryUserStore) GetByUsername(_ context.Context, username string) (*model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, ok := s.byUsername[username]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	copyUser := *s.byID[id]
	return &copyUser, nil
}

// MemoryProfileStore is an in-memory ProfileStore for tests.
type MemoryProfileStore struct {
	mu       sync.Mutex
	profiles map[string]*model.DjProfile
	genres   map[string][]string
}

func NewMemoryProfileStore() *MemoryProfileStore {
	return &MemoryProfileStore{
		profiles: make(map[string]*model.DjProfile),
		genres:   make(map[string][]string),
	}
}

func (s *MemoryProfileStore) Create(_ context.Context, p *model.DjProfile) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	copyProfile := *p
	copyProfile.CreatedAt = time.Now()
	s.profiles[p.ID] = &copyProfile
	return nil
}

func (s *MemoryProfileStore) Update(_ context.Context, p *model.DjProfile) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.profiles[p.ID]; !ok {
		return fmt.Errorf("profile not found")
	}
	copyProfile := *p
	s.profiles[p.ID] = &copyProfile
	return nil
}

func (s *MemoryProfileStore) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.profiles[id]; !ok {
		return fmt.Errorf("profile not found")
	}
	delete(s.profiles, id)
	delete(s.genres, id)
	return nil
}

func (s *MemoryProfileStore) GetByID(_ context.Context, id string) (*model.DjProfile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.profiles[id]
	if !ok {
		return nil, fmt.Errorf("profile not found")
	}
	copyProfile := *p
	return &copyProfile, nil
}

func (s *MemoryProfileStore) ListByUserID(_ context.Context, userID string) ([]*model.DjProfile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var profiles []*model.DjProfile
	for _, p := range s.profiles {
		if p.UserID == userID {
			copyProfile := *p
			profiles = append(profiles, &copyProfile)
		}
	}
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].CreatedAt.Before(profiles[j].CreatedAt)
	})
	return profiles, nil
}

func (s *MemoryProfileStore) GetOwnerID(_ context.Context, profileID string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.profiles[profileID]
	if !ok {
		return "", fmt.Errorf("profile not found")
	}
	return p.UserID, nil
}

func (s *MemoryProfileStore) SetGenres(_ context.Context, profileID string, genres []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.profiles[profileID]; !ok {
		return fmt.Errorf("profile not found")
	}
	s.genres[profileID] = append([]string(nil), genres...)
	return nil
}

func (s *MemoryProfileStore) GetGenres(_ context.Context, profileID string) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	genres, ok := s.genres[profileID]
	if !ok {
		return []string{}, nil
	}
	return append([]string(nil), genres...), nil
}

func (s *MemoryProfileStore) AddPhoto(_ context.Context, _ *model.ProfilePhoto) error {
	return nil
}

func (s *MemoryProfileStore) DeletePhoto(_ context.Context, _ string) error {
	return nil
}

func (s *MemoryProfileStore) GetPhotos(_ context.Context, _ string) ([]*model.ProfilePhoto, error) {
	return []*model.ProfilePhoto{}, nil
}

func (s *MemoryProfileStore) GetPhotoByID(_ context.Context, _ string) (*model.ProfilePhoto, error) {
	return nil, errNotFound
}

func (s *MemoryProfileStore) ReorderPhotos(_ context.Context, _ string, _ []string) error {
	return nil
}

func (s *MemoryProfileStore) UpsertSocialLinkTx(_ context.Context, link *model.SocialLink) (*model.SocialLink, error) {
	return link, nil
}

func (s *MemoryProfileStore) DeleteSocialLink(_ context.Context, _ string) error {
	return nil
}

func (s *MemoryProfileStore) GetSocialLinks(_ context.Context, _ string) ([]*model.SocialLink, error) {
	return []*model.SocialLink{}, nil
}

// MemoryEventStore is an in-memory EventStore for tests.
type MemoryEventStore struct {
	mu      sync.Mutex
	events  map[string]*model.Event
	history map[string][]*model.EventStatusHistory
	owners  map[string]string
}

func NewMemoryEventStore() *MemoryEventStore {
	return &MemoryEventStore{
		events:  make(map[string]*model.Event),
		history: make(map[string][]*model.EventStatusHistory),
		owners:  make(map[string]string),
	}
}

func (s *MemoryEventStore) SetOwner(eventID, userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.owners[eventID] = userID
}

func (s *MemoryEventStore) Create(_ context.Context, e *model.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	copyEvent := *e
	copyEvent.CreatedAt = time.Now()
	s.events[e.ID] = &copyEvent
	return nil
}

func (s *MemoryEventStore) Update(_ context.Context, e *model.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.events[e.ID]; !ok {
		return fmt.Errorf("event not found")
	}
	copyEvent := *e
	createdAt := s.events[e.ID].CreatedAt
	copyEvent.CreatedAt = createdAt
	s.events[e.ID] = &copyEvent
	return nil
}

func (s *MemoryEventStore) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.events[id]; !ok {
		return fmt.Errorf("event not found")
	}
	delete(s.events, id)
	delete(s.history, id)
	delete(s.owners, id)
	return nil
}

func (s *MemoryEventStore) GetByID(_ context.Context, id string) (*model.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.events[id]
	if !ok {
		return nil, fmt.Errorf("event not found")
	}
	copyEvent := *e
	return &copyEvent, nil
}

func (s *MemoryEventStore) GetCurrentStatus(_ context.Context, id string) (model.EventStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.events[id]
	if !ok {
		return "", fmt.Errorf("event not found")
	}
	return e.EventStatus, nil
}

func (s *MemoryEventStore) List(_ context.Context, filter EventFilter) ([]*model.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var events []*model.Event
	for _, e := range s.events {
		if e.ProfileID != filter.ProfileID {
			continue
		}
		if filter.EventStatus != nil && e.EventStatus != *filter.EventStatus {
			continue
		}
		if filter.PaymentStatus != nil && e.PaymentStatus != *filter.PaymentStatus {
			continue
		}
		copyEvent := *e
		events = append(events, &copyEvent)
	}
	sort.Slice(events, func(i, j int) bool {
		return events[i].Date.After(events[j].Date)
	})
	return events, nil
}

func (s *MemoryEventStore) ListUpcoming(_ context.Context, profileID string) ([]*model.Event, error) {
	today := time.Now().Truncate(24 * time.Hour)
	filterPending := model.EventStatusPending
	filterConfirmed := model.EventStatusConfirmed

	all, err := s.List(context.Background(), EventFilter{ProfileID: profileID})
	if err != nil {
		return nil, err
	}
	var upcoming []*model.Event
	for _, e := range all {
		if e.Date.Before(today) {
			continue
		}
		if e.EventStatus != filterPending && e.EventStatus != filterConfirmed {
			continue
		}
		upcoming = append(upcoming, e)
	}
	sort.Slice(upcoming, func(i, j int) bool {
		return upcoming[i].Date.Before(upcoming[j].Date)
	})
	return upcoming, nil
}

func (s *MemoryEventStore) AddStatusHistory(_ context.Context, h *model.EventStatusHistory) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	copyHistory := *h
	copyHistory.ChangedAt = time.Now()
	s.history[h.EventID] = append(s.history[h.EventID], &copyHistory)
	return nil
}

func (s *MemoryEventStore) GetStatusHistory(_ context.Context, eventID string) ([]*model.EventStatusHistory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	history := s.history[eventID]
	if history == nil {
		return []*model.EventStatusHistory{}, nil
	}
	out := make([]*model.EventStatusHistory, len(history))
	for i, h := range history {
		copyHistory := *h
		out[i] = &copyHistory
	}
	return out, nil
}

func (s *MemoryEventStore) GetOwnerID(_ context.Context, eventID string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	userID, ok := s.owners[eventID]
	if !ok {
		return "", fmt.Errorf("event not found")
	}
	return userID, nil
}

// ValidateEventInput checks required event fields before persistence.
func ValidateEventInput(input model.Event) error {
	if strings.TrimSpace(input.Title) == "" {
		return fmt.Errorf("title is required")
	}
	if input.Date.IsZero() {
		return fmt.Errorf("date is required")
	}
	return nil
}

package service

import (
	"context"
	"fmt"
	"time"

	"github.com/djbook/backend/internal/model"
	"github.com/djbook/backend/internal/repository"
	"github.com/google/uuid"
)

type EventService struct {
	repo *repository.EventRepository
}

func NewEventService(repo *repository.EventRepository) *EventService {
	return &EventService{repo: repo}
}

func (s *EventService) Create(ctx context.Context, profileID string, input model.Event) (*model.Event, error) {
	input.ID = uuid.New().String()
	input.ProfileID = profileID

	if err := s.repo.Create(ctx, &input); err != nil {
		return nil, fmt.Errorf("create event: %w", err)
	}

	// Record initial status in history
	history := &model.EventStatusHistory{
		ID:        uuid.New().String(),
		EventID:   input.ID,
		OldStatus: nil,
		NewStatus: string(input.EventStatus),
	}
	if err := s.repo.AddStatusHistory(ctx, history); err != nil {
		return nil, fmt.Errorf("add status history: %w", err)
	}

	return s.repo.GetByID(ctx, input.ID)
}

func (s *EventService) Update(ctx context.Context, id string, input model.Event) (*model.Event, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	statusChanged := existing.EventStatus != input.EventStatus

	input.ID = id
	input.ProfileID = existing.ProfileID
	if err := s.repo.Update(ctx, &input); err != nil {
		return nil, fmt.Errorf("update event: %w", err)
	}

	if statusChanged {
		oldStatus := string(existing.EventStatus)
		history := &model.EventStatusHistory{
			ID:        uuid.New().String(),
			EventID:   id,
			OldStatus: &oldStatus,
			NewStatus: string(input.EventStatus),
		}
		if err := s.repo.AddStatusHistory(ctx, history); err != nil {
			return nil, fmt.Errorf("add status history: %w", err)
		}
	}

	return s.repo.GetByID(ctx, id)
}

func (s *EventService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *EventService) GetByID(ctx context.Context, id string) (*model.Event, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *EventService) List(ctx context.Context, profileID string, filter *repository.EventFilter) ([]*model.Event, error) {
	f := repository.EventFilter{ProfileID: profileID}
	if filter != nil {
		f.Month = filter.Month
		f.EventStatus = filter.EventStatus
		f.PaymentStatus = filter.PaymentStatus
	}
	events, err := s.repo.List(ctx, f)
	if err != nil {
		return nil, err
	}
	if events == nil {
		return []*model.Event{}, nil
	}
	return events, nil
}

func (s *EventService) ListUpcoming(ctx context.Context, profileID string) ([]*model.Event, error) {
	events, err := s.repo.ListUpcoming(ctx, profileID)
	if err != nil {
		return nil, err
	}
	if events == nil {
		return []*model.Event{}, nil
	}
	return events, nil
}

func (s *EventService) GetStatusHistory(ctx context.Context, eventID string) ([]*model.EventStatusHistory, error) {
	history, err := s.repo.GetStatusHistory(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if history == nil {
		return []*model.EventStatusHistory{}, nil
	}
	return history, nil
}

func (s *EventService) GetOwnerID(ctx context.Context, eventID string) (string, error) {
	return s.repo.GetOwnerID(ctx, eventID)
}

// AssertOwner verifies that userID owns the given event.
func (s *EventService) AssertOwner(ctx context.Context, eventID, userID string) error {
	ownerID, err := s.repo.GetOwnerID(ctx, eventID)
	if err != nil {
		return err
	}
	if ownerID != userID {
		return fmt.Errorf("forbidden: you do not own this event")
	}
	return nil
}

// ParseDate parses a date string in YYYY-MM-DD format.
func ParseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

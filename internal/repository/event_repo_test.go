package repository_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/djbook/backend/internal/model"
	"github.com/djbook/backend/internal/repository"
	"github.com/djbook/backend/internal/service"
)

func newEventService() (*service.EventService, *repository.MemoryEventStore) {
	store := repository.NewMemoryEventStore()
	return service.NewEventService(store), store
}

func validEventInput() model.Event {
	return model.Event{
		Title:         "Warehouse Night",
		Venue:         "Main Hall",
		Date:          time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC),
		StartTime:     "22:00",
		EventStatus:   model.EventStatusPending,
		PaymentStatus: model.PaymentStatusUnpaid,
	}
}

func TestCreateEvent_Success(t *testing.T) {
	svc, _ := newEventService()
	ctx := context.Background()
	profileID := "profile-1"
	input := validEventInput()

	created, err := svc.Create(ctx, profileID, input)
	if err != nil {
		t.Fatalf("create event: %v", err)
	}

	if created.ID == "" {
		t.Fatal("expected generated event id")
	}
	if created.ProfileID != profileID {
		t.Fatalf("profileID = %q, want %q", created.ProfileID, profileID)
	}
	if created.Title != input.Title {
		t.Fatalf("title = %q, want %q", created.Title, input.Title)
	}
	if !created.Date.Equal(input.Date) {
		t.Fatalf("date = %v, want %v", created.Date, input.Date)
	}
	if created.EventStatus != model.EventStatusPending {
		t.Fatalf("eventStatus = %q, want %q", created.EventStatus, model.EventStatusPending)
	}
}

func TestCreateEvent_MissingRequiredFields(t *testing.T) {
	svc, _ := newEventService()
	ctx := context.Background()
	profileID := "profile-1"

	t.Run("missing title", func(t *testing.T) {
		input := validEventInput()
		input.Title = "   "
		_, err := svc.Create(ctx, profileID, input)
		if err == nil {
			t.Fatal("expected validation error for missing title")
		}
		if !strings.Contains(err.Error(), "title is required") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("missing date", func(t *testing.T) {
		input := validEventInput()
		input.Date = time.Time{}
		_, err := svc.Create(ctx, profileID, input)
		if err == nil {
			t.Fatal("expected validation error for missing date")
		}
		if !strings.Contains(err.Error(), "date is required") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestUpdateEventStatus_ValidTransition(t *testing.T) {
	svc, store := newEventService()
	ctx := context.Background()
	profileID := "profile-1"

	created, err := svc.Create(ctx, profileID, validEventInput())
	if err != nil {
		t.Fatalf("create event: %v", err)
	}
	if created.EventStatus != model.EventStatusPending {
		t.Fatalf("initial status = %q, want PENDING", created.EventStatus)
	}

	updatedInput := validEventInput()
	updatedInput.EventStatus = model.EventStatusConfirmed

	updated, err := svc.Update(ctx, created.ID, updatedInput)
	if err != nil {
		t.Fatalf("update event: %v", err)
	}
	if updated.EventStatus != model.EventStatusConfirmed {
		t.Fatalf("status = %q, want CONFIRMED", updated.EventStatus)
	}

	history, err := store.GetStatusHistory(ctx, created.ID)
	if err != nil {
		t.Fatalf("get status history: %v", err)
	}
	if len(history) < 2 {
		t.Fatalf("expected at least 2 history entries, got %d", len(history))
	}
	last := history[len(history)-1]
	if last.OldStatus == nil || *last.OldStatus != string(model.EventStatusPending) {
		t.Fatalf("unexpected old status in history: %+v", last)
	}
	if last.NewStatus != string(model.EventStatusConfirmed) {
		t.Fatalf("new status = %q, want CONFIRMED", last.NewStatus)
	}
}

func TestDeleteEvent_Success(t *testing.T) {
	svc, _ := newEventService()
	ctx := context.Background()
	profileID := "profile-1"

	created, err := svc.Create(ctx, profileID, validEventInput())
	if err != nil {
		t.Fatalf("create event: %v", err)
	}

	events, err := svc.List(ctx, profileID, nil)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event before delete, got %d", len(events))
	}

	if err := svc.Delete(ctx, created.ID); err != nil {
		t.Fatalf("delete event: %v", err)
	}

	events, err = svc.List(ctx, profileID, nil)
	if err != nil {
		t.Fatalf("list events after delete: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected 0 events after delete, got %d", len(events))
	}

	if _, err := svc.GetByID(ctx, created.ID); err == nil {
		t.Fatal("expected get by id to fail after delete")
	}
}

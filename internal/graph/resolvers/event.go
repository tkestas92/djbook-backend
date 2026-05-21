package resolvers

import (
	"context"
	"fmt"

	gmodel "github.com/djbook/backend/internal/graph/model"
	"github.com/djbook/backend/internal/auth"
	"github.com/djbook/backend/internal/model"
	"github.com/djbook/backend/internal/service"
)

func isPublicFromInput(input *bool) bool {
	if input != nil {
		return *input
	}
	return false
}

func (m *mutationResolver) CreateEvent(ctx context.Context, profileID string, input gmodel.EventInput) (*gmodel.Event, error) {
	userID, err := auth.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	if err := m.ProfileService.AssertOwner(ctx, profileID, userID); err != nil {
		return nil, err
	}

	date, err := service.ParseDate(input.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format (expected YYYY-MM-DD): %w", err)
	}

	dbEvent := model.Event{
		Title:         input.Title,
		Venue:         input.Venue,
		Date:          date,
		StartTime:     input.StartTime,
		EndTime:       input.EndTime,
		Notes:         input.Notes,
		AmountEur:     input.AmountEur,
		IsPublic:      isPublicFromInput(input.IsPublic),
		TicketsURL:    input.TicketsURL,
		EventStatus:   model.EventStatus(input.EventStatus),
		PaymentStatus: model.PaymentStatus(input.PaymentStatus),
	}

	created, err := m.EventService.Create(ctx, profileID, dbEvent)
	if err != nil {
		return nil, fmt.Errorf("create event: %w", err)
	}
	return toGQLEvent(ctx, created, m.Resolver)
}

func (m *mutationResolver) UpdateEvent(ctx context.Context, id string, input gmodel.EventInput) (*gmodel.Event, error) {
	userID, err := auth.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	if err := m.EventService.AssertOwner(ctx, id, userID); err != nil {
		return nil, err
	}

	date, err := service.ParseDate(input.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format (expected YYYY-MM-DD): %w", err)
	}

	dbEvent := model.Event{
		Title:         input.Title,
		Venue:         input.Venue,
		Date:          date,
		StartTime:     input.StartTime,
		EndTime:       input.EndTime,
		Notes:         input.Notes,
		AmountEur:     input.AmountEur,
		IsPublic:      isPublicFromInput(input.IsPublic),
		TicketsURL:    input.TicketsURL,
		EventStatus:   model.EventStatus(input.EventStatus),
		PaymentStatus: model.PaymentStatus(input.PaymentStatus),
	}

	updated, err := m.EventService.Update(ctx, id, dbEvent)
	if err != nil {
		return nil, fmt.Errorf("update event: %w", err)
	}
	return toGQLEvent(ctx, updated, m.Resolver)
}

func (m *mutationResolver) DeleteEvent(ctx context.Context, id string) (bool, error) {
	userID, err := auth.RequireUserID(ctx)
	if err != nil {
		return false, err
	}
	if err := m.EventService.AssertOwner(ctx, id, userID); err != nil {
		return false, err
	}
	if err := m.EventService.Delete(ctx, id); err != nil {
		return false, err
	}
	return true, nil
}

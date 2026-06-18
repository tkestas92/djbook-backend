package resolvers

import (
	"context"
	"fmt"

	"github.com/djbook/backend/internal/graph"
	gmodel "github.com/djbook/backend/internal/graph/model"
	"github.com/djbook/backend/internal/model"
	"github.com/djbook/backend/internal/repository"
	"github.com/djbook/backend/internal/service"
)

func profileFromInput(input gmodel.DjProfileInput, bio *string) model.DjProfile {
	return model.DjProfile{
		DjName:            input.DjName,
		Bio:               bio,
		Email:             input.Email,
		ManagementContact: input.ManagementContact,
	}
}

// Resolver embeds the root graph.Resolver and implements graph.ResolverRoot.
type Resolver struct {
	*graph.Resolver
}

// NewResolver creates a new Resolver with all services.
func NewResolver(
	userSvc *service.UserService,
	profileSvc *service.ProfileService,
	eventSvc *service.EventService,
	releaseSvc *service.ReleaseService,
	financeSvc *service.FinanceService,
) *Resolver {
	return &Resolver{
		Resolver: &graph.Resolver{
			UserService:    userSvc,
			ProfileService: profileSvc,
			EventService:   eventSvc,
			ReleaseService: releaseSvc,
			FinanceService: financeSvc,
		},
	}
}

// toGQLProfile converts a DB profile to a GraphQL profile, loading all sub-resources.
func toGQLProfile(ctx context.Context, p *model.DjProfile, r *graph.Resolver) (*gmodel.DjProfile, error) {
	genres, err := r.ProfileService.GetGenres(ctx, p.ID)
	if err != nil {
		return nil, fmt.Errorf("load genres: %w", err)
	}

	dbPhotos, err := r.ProfileService.GetPhotos(ctx, p.ID)
	if err != nil {
		return nil, fmt.Errorf("load photos: %w", err)
	}
	photos := make([]*gmodel.ProfilePhoto, 0, len(dbPhotos))
	for _, ph := range dbPhotos {
		photos = append(photos, &gmodel.ProfilePhoto{
			ID:        ph.ID,
			URL:       ph.URL,
			SortOrder: ph.SortOrder,
			CreatedAt: ph.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	dbLinks, err := r.ProfileService.GetSocialLinks(ctx, p.ID)
	if err != nil {
		return nil, fmt.Errorf("load social links: %w", err)
	}
	links := make([]*gmodel.SocialLink, 0, len(dbLinks))
	for _, l := range dbLinks {
		links = append(links, &gmodel.SocialLink{
			ID:       l.ID,
			Platform: l.Platform,
			URL:      l.URL,
		})
	}

	dbEvents, err := r.EventService.List(ctx, p.ID, nil)
	if err != nil {
		return nil, fmt.Errorf("load events: %w", err)
	}
	events, err := toGQLEvents(ctx, dbEvents, r)
	if err != nil {
		return nil, err
	}

	dbUpcoming, err := r.EventService.ListUpcoming(ctx, p.ID)
	if err != nil {
		return nil, fmt.Errorf("load upcoming events: %w", err)
	}
	upcoming, err := toGQLEvents(ctx, dbUpcoming, r)
	if err != nil {
		return nil, err
	}

	bio := (*string)(nil)
	if p.Bio != nil {
		bio = p.Bio
	}

	return &gmodel.DjProfile{
		ID:                p.ID,
		DjName:            p.DjName,
		Bio:               bio,
		Email:             p.Email,
		ManagementContact: p.ManagementContact,
		Genres:            genres,
		Photos:            photos,
		SocialLinks:       links,
		Events:            events,
		UpcomingEvents:    upcoming,
		CreatedAt:         p.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}

func toGQLEvents(ctx context.Context, events []*model.Event, r *graph.Resolver) ([]*gmodel.Event, error) {
	result := make([]*gmodel.Event, 0, len(events))
	for _, e := range events {
		gEvent, err := toGQLEvent(ctx, e, r)
		if err != nil {
			return nil, err
		}
		result = append(result, gEvent)
	}
	return result, nil
}

func toGQLEvent(ctx context.Context, e *model.Event, r *graph.Resolver) (*gmodel.Event, error) {
	history, err := r.EventService.GetStatusHistory(ctx, e.ID)
	if err != nil {
		return nil, fmt.Errorf("load status history: %w", err)
	}
	gHistory := make([]*gmodel.EventStatusHistory, 0, len(history))
	for _, h := range history {
		gHistory = append(gHistory, &gmodel.EventStatusHistory{
			ID:        h.ID,
			OldStatus: h.OldStatus,
			NewStatus: h.NewStatus,
			ChangedAt: h.ChangedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	gEvent := &gmodel.Event{
		ID:            e.ID,
		Title:         e.Title,
		Venue:         e.Venue,
		Date:          e.Date.Format("2006-01-02"),
		StartTime:     e.StartTime,
		EndTime:       e.EndTime,
		Notes:         e.Notes,
		AmountEur:     e.AmountEur,
		IsPublic:      e.IsPublic,
		TicketsURL:    e.TicketsURL,
		EventStatus:   gmodel.EventStatus(e.EventStatus),
		PaymentStatus: gmodel.PaymentStatus(e.PaymentStatus),
		StatusHistory: gHistory,
		CreatedAt:     e.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	return gEvent, nil
}

func toRepoFilter(f *gmodel.EventFilter) *repository.EventFilter {
	if f == nil {
		return nil
	}
	rf := &repository.EventFilter{}
	rf.Month = f.Month
	if f.EventStatus != nil {
		s := model.EventStatus(*f.EventStatus)
		rf.EventStatus = &s
	}
	if f.PaymentStatus != nil {
		ps := model.PaymentStatus(*f.PaymentStatus)
		rf.PaymentStatus = &ps
	}
	return rf
}

func toServicePeriod(p *gmodel.Period) service.Period {
	if p == nil {
		return service.PeriodYear
	}
	switch *p {
	case gmodel.PeriodMonth:
		return service.PeriodMonth
	case gmodel.PeriodQuarter:
		return service.PeriodQuarter
	default:
		return service.PeriodYear
	}
}

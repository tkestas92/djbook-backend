package resolvers

import (
	"context"
	"fmt"

	"github.com/djbook/backend/internal/auth"
	"github.com/djbook/backend/internal/graph"
	gmodel "github.com/djbook/backend/internal/graph/model"
)

// Ensure Resolver implements graph.QueryResolver (partial).
var _ graph.QueryResolver = (*queryResolver)(nil)

type queryResolver struct{ *graph.Resolver }

func (r *Resolver) Query() graph.QueryResolver { return &queryResolver{r.Resolver} }

func (q *queryResolver) Me(ctx context.Context) (*gmodel.User, error) {
	userID, err := auth.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}

	u, err := q.UserService.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	profiles, err := q.ProfileService.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	gProfiles := make([]*gmodel.DjProfile, 0, len(profiles))
	for _, p := range profiles {
		gp, err := toGQLProfile(ctx, p, q.Resolver)
		if err != nil {
			return nil, err
		}
		gProfiles = append(gProfiles, gp)
	}

	return &gmodel.User{
		ID:        u.ID,
		Email:     u.Email,
		Profiles:  gProfiles,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}

func (q *queryResolver) DjProfile(ctx context.Context, id string) (*gmodel.DjProfile, error) {
	p, err := q.ProfileService.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("profile not found: %w", err)
	}
	return toGQLProfile(ctx, p, q.Resolver)
}

func (q *queryResolver) Events(ctx context.Context, profileID string, filter *gmodel.EventFilter) ([]*gmodel.Event, error) {
	repoFilter := toRepoFilter(filter)
	events, err := q.EventService.List(ctx, profileID, repoFilter)
	if err != nil {
		return nil, err
	}
	return toGQLEvents(ctx, events, q.Resolver)
}

func (q *queryResolver) FinanceSummary(ctx context.Context, profileID string, period *gmodel.Period) (*gmodel.FinanceSummary, error) {
	_, err := auth.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}

	p := toServicePeriod(period)
	summary, err := q.FinanceService.GetSummary(ctx, profileID, p)
	if err != nil {
		return nil, err
	}

	result := &gmodel.FinanceSummary{
		ThisMonthEarnings:   summary.ThisMonthEarnings,
		ThisMonthEventCount: summary.ThisMonthEventCount,
		ThisYearEarnings:    summary.ThisYearEarnings,
		ThisYearEventCount:  summary.ThisYearEventCount,
		UnpaidAmount:        summary.UnpaidAmount,
		UnpaidEventCount:    summary.UnpaidEventCount,
		MonthlyChart:        make([]*gmodel.MonthlyEarning, 0),
	}
	for _, m := range summary.MonthlyChart {
		result.MonthlyChart = append(result.MonthlyChart, &gmodel.MonthlyEarning{
			Month:      m.Month,
			Earnings:   m.Earnings,
			EventCount: m.EventCount,
		})
	}
	if summary.BestMonth != nil {
		result.BestMonth = &gmodel.MonthlyEarning{
			Month:      summary.BestMonth.Month,
			Earnings:   summary.BestMonth.Earnings,
			EventCount: summary.BestMonth.EventCount,
		}
	}
	return result, nil
}

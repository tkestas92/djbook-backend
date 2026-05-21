package service

import (
	"context"
	"fmt"

	"github.com/djbook/backend/internal/repository"
)

// FinanceSummaryResult aggregates all finance metrics.
type FinanceSummaryResult struct {
	ThisMonthEarnings   float64
	ThisMonthEventCount int
	ThisYearEarnings    float64
	ThisYearEventCount  int
	UnpaidAmount        float64
	UnpaidEventCount    int
	MonthlyChart        []*MonthlyEarningResult
	BestMonth           *MonthlyEarningResult
}

type MonthlyEarningResult struct {
	Month      string
	Earnings   float64
	EventCount int
}

type FinanceService struct {
	repo *repository.FinanceRepository
}

func NewFinanceService(repo *repository.FinanceRepository) *FinanceService {
	return &FinanceService{repo: repo}
}

type Period string

const (
	PeriodMonth   Period = "MONTH"
	PeriodQuarter Period = "QUARTER"
	PeriodYear    Period = "YEAR"
)

func (s *FinanceService) GetSummary(ctx context.Context, profileID string, period Period) (*FinanceSummaryResult, error) {
	result := &FinanceSummaryResult{}

	// This month (always shown regardless of period)
	monthSummary, err := s.repo.GetThisMonthSummary(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("month summary: %w", err)
	}
	result.ThisMonthEarnings = monthSummary.TotalEarnings
	result.ThisMonthEventCount = monthSummary.EventCount

	// Year (always shown)
	yearSummary, err := s.repo.GetThisYearSummary(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("year summary: %w", err)
	}
	result.ThisYearEarnings = yearSummary.TotalEarnings
	result.ThisYearEventCount = yearSummary.EventCount

	// Unpaid
	unpaid, err := s.repo.GetUnpaidSummary(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("unpaid summary: %w", err)
	}
	result.UnpaidAmount = unpaid.TotalEarnings
	result.UnpaidEventCount = unpaid.EventCount

	// Monthly chart — number of months depends on period
	months := 12
	switch period {
	case PeriodMonth:
		months = 1
	case PeriodQuarter:
		months = 3
	case PeriodYear:
		months = 12
	}

	chart, err := s.repo.GetMonthlyChart(ctx, profileID, months)
	if err != nil {
		return nil, fmt.Errorf("monthly chart: %w", err)
	}
	for _, m := range chart {
		result.MonthlyChart = append(result.MonthlyChart, &MonthlyEarningResult{
			Month:      m.Month,
			Earnings:   m.Earnings,
			EventCount: m.EventCount,
		})
	}
	if result.MonthlyChart == nil {
		result.MonthlyChart = []*MonthlyEarningResult{}
	}

	best, err := s.repo.GetBestMonth(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("best month: %w", err)
	}
	if best != nil {
		result.BestMonth = &MonthlyEarningResult{
			Month:      best.Month,
			Earnings:   best.Earnings,
			EventCount: best.EventCount,
		}
	}

	return result, nil
}

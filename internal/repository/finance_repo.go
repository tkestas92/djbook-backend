package repository

import (
	"context"
	"database/sql"
)

type FinanceRepository struct {
	db *sql.DB
}

func NewFinanceRepository(db *sql.DB) *FinanceRepository {
	return &FinanceRepository{db: db}
}

type PeriodSummary struct {
	TotalEarnings float64
	EventCount    int
}

func (r *FinanceRepository) GetThisMonthSummary(ctx context.Context, profileID string) (PeriodSummary, error) {
	return r.querySummary(ctx, profileID, `
		WHERE profile_id = ?
		  AND YEAR(date) = YEAR(CURDATE())
		  AND MONTH(date) = MONTH(CURDATE())
		  AND event_status = 'COMPLETED'
	`)
}

func (r *FinanceRepository) GetThisYearSummary(ctx context.Context, profileID string) (PeriodSummary, error) {
	return r.querySummary(ctx, profileID, `
		WHERE profile_id = ?
		  AND YEAR(date) = YEAR(CURDATE())
		  AND event_status = 'COMPLETED'
	`)
}

func (r *FinanceRepository) GetQuarterSummary(ctx context.Context, profileID string) (PeriodSummary, error) {
	return r.querySummary(ctx, profileID, `
		WHERE profile_id = ?
		  AND YEAR(date) = YEAR(CURDATE())
		  AND QUARTER(date) = QUARTER(CURDATE())
		  AND event_status = 'COMPLETED'
	`)
}

func (r *FinanceRepository) GetUnpaidSummary(ctx context.Context, profileID string) (PeriodSummary, error) {
	var s PeriodSummary
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(amount_eur), 0), COUNT(*)
		FROM events
		WHERE profile_id = ?
		  AND payment_status = 'UNPAID'
		  AND event_status = 'COMPLETED'
	`, profileID).Scan(&s.TotalEarnings, &s.EventCount)
	return s, err
}

type MonthlyEarning struct {
	Month      string
	Earnings   float64
	EventCount int
}

func (r *FinanceRepository) GetMonthlyChart(ctx context.Context, profileID string, months int) ([]*MonthlyEarning, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT DATE_FORMAT(date, '%Y-%m') AS month,
		       COALESCE(SUM(amount_eur), 0) AS earnings,
		       COUNT(*) AS event_count
		FROM events
		WHERE profile_id = ?
		  AND event_status = 'COMPLETED'
		  AND date >= DATE_SUB(CURDATE(), INTERVAL ? MONTH)
		GROUP BY month
		ORDER BY month
	`, profileID, months)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*MonthlyEarning
	for rows.Next() {
		m := &MonthlyEarning{}
		if err := rows.Scan(&m.Month, &m.Earnings, &m.EventCount); err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, rows.Err()
}

func (r *FinanceRepository) GetBestMonth(ctx context.Context, profileID string) (*MonthlyEarning, error) {
	m := &MonthlyEarning{}
	err := r.db.QueryRowContext(ctx, `
		SELECT DATE_FORMAT(date, '%Y-%m') AS month,
		       COALESCE(SUM(amount_eur), 0) AS earnings,
		       COUNT(*) AS event_count
		FROM events
		WHERE profile_id = ?
		  AND event_status = 'COMPLETED'
		GROUP BY month
		ORDER BY earnings DESC
		LIMIT 1
	`, profileID).Scan(&m.Month, &m.Earnings, &m.EventCount)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return m, err
}

func (r *FinanceRepository) querySummary(ctx context.Context, profileID, condition string) (PeriodSummary, error) {
	var s PeriodSummary
	err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount_eur), 0), COUNT(*) FROM events `+condition,
		profileID,
	).Scan(&s.TotalEarnings, &s.EventCount)
	return s, err
}

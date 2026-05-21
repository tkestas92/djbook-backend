package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/djbook/backend/internal/model"
)

type EventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Create(ctx context.Context, e *model.Event) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO events (id, profile_id, title, venue, date, start_time, end_time, notes, amount_eur, is_public, tickets_url, event_status, payment_status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.ID, e.ProfileID, e.Title, e.Venue, e.Date.Format("2006-01-02"),
		e.StartTime, e.EndTime, e.Notes, e.AmountEur, e.IsPublic, e.TicketsURL, e.EventStatus, e.PaymentStatus,
	)
	return err
}

func (r *EventRepository) Update(ctx context.Context, e *model.Event) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE events SET title=?, venue=?, date=?, start_time=?, end_time=?, notes=?, amount_eur=?, is_public=?, tickets_url=?, event_status=?, payment_status=?
		 WHERE id=?`,
		e.Title, e.Venue, e.Date.Format("2006-01-02"),
		e.StartTime, e.EndTime, e.Notes, e.AmountEur, e.IsPublic, e.TicketsURL, e.EventStatus, e.PaymentStatus,
		e.ID,
	)
	return err
}

func (r *EventRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM events WHERE id = ?`, id)
	return err
}

func (r *EventRepository) GetByID(ctx context.Context, id string) (*model.Event, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, profile_id, title, venue, date, start_time, end_time, notes, amount_eur, is_public, tickets_url, event_status, payment_status, created_at
		 FROM events WHERE id = ?`, id,
	)
	return r.scanEvent(row)
}

func (r *EventRepository) GetCurrentStatus(ctx context.Context, id string) (model.EventStatus, error) {
	var status model.EventStatus
	err := r.db.QueryRowContext(ctx,
		`SELECT event_status FROM events WHERE id = ?`, id,
	).Scan(&status)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("event not found")
	}
	return status, err
}

type EventFilter struct {
	ProfileID     string
	Month         *string
	EventStatus   *model.EventStatus
	PaymentStatus *model.PaymentStatus
}

func (r *EventRepository) List(ctx context.Context, filter EventFilter) ([]*model.Event, error) {
	query := `SELECT id, profile_id, title, venue, date, start_time, end_time, notes, amount_eur, is_public, tickets_url, event_status, payment_status, created_at
	          FROM events WHERE profile_id = ?`
	args := []interface{}{filter.ProfileID}

	if filter.Month != nil {
		query += ` AND DATE_FORMAT(date, '%Y-%m') = ?`
		args = append(args, *filter.Month)
	}
	if filter.EventStatus != nil {
		query += ` AND event_status = ?`
		args = append(args, *filter.EventStatus)
	}
	if filter.PaymentStatus != nil {
		query += ` AND payment_status = ?`
		args = append(args, *filter.PaymentStatus)
	}
	query += ` ORDER BY date DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanEvents(rows)
}

func (r *EventRepository) ListUpcoming(ctx context.Context, profileID string) ([]*model.Event, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, profile_id, title, venue, date, start_time, end_time, notes, amount_eur, is_public, tickets_url, event_status, payment_status, created_at
		 FROM events
		 WHERE profile_id = ? AND date >= CURDATE() AND event_status IN ('PENDING','CONFIRMED')
		 ORDER BY date ASC`,
		profileID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanEvents(rows)
}

// Status history

func (r *EventRepository) AddStatusHistory(ctx context.Context, h *model.EventStatusHistory) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO event_status_history (id, event_id, old_status, new_status) VALUES (?, ?, ?, ?)`,
		h.ID, h.EventID, h.OldStatus, h.NewStatus,
	)
	return err
}

func (r *EventRepository) GetStatusHistory(ctx context.Context, eventID string) ([]*model.EventStatusHistory, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, event_id, old_status, new_status, changed_at FROM event_status_history WHERE event_id = ? ORDER BY changed_at`,
		eventID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*model.EventStatusHistory
	for rows.Next() {
		h := &model.EventStatusHistory{}
		if err := rows.Scan(&h.ID, &h.EventID, &h.OldStatus, &h.NewStatus, &h.ChangedAt); err != nil {
			return nil, err
		}
		history = append(history, h)
	}
	return history, rows.Err()
}

// GetOwnerID returns the user_id that owns the given event (via profile).
func (r *EventRepository) GetOwnerID(ctx context.Context, eventID string) (string, error) {
	var userID string
	err := r.db.QueryRowContext(ctx,
		`SELECT p.user_id FROM events e JOIN dj_profiles p ON e.profile_id = p.id WHERE e.id = ?`,
		eventID,
	).Scan(&userID)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("event not found")
	}
	return userID, err
}

func (r *EventRepository) scanEvent(row *sql.Row) (*model.Event, error) {
	e := &model.Event{}
	err := row.Scan(
		&e.ID, &e.ProfileID, &e.Title, &e.Venue, &e.Date,
		&e.StartTime, &e.EndTime, &e.Notes, &e.AmountEur,
		&e.IsPublic, &e.TicketsURL, &e.EventStatus, &e.PaymentStatus, &e.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("event not found")
	}
	return e, err
}

func (r *EventRepository) scanEvents(rows *sql.Rows) ([]*model.Event, error) {
	var events []*model.Event
	for rows.Next() {
		e := &model.Event{}
		if err := rows.Scan(
			&e.ID, &e.ProfileID, &e.Title, &e.Venue, &e.Date,
			&e.StartTime, &e.EndTime, &e.Notes, &e.AmountEur,
			&e.IsPublic, &e.TicketsURL, &e.EventStatus, &e.PaymentStatus, &e.CreatedAt,
		); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// BuildInPlaceholder builds a MySQL IN clause like (?,?,?) with n placeholders.
func BuildInPlaceholder(n int) string {
	if n == 0 {
		return "(NULL)"
	}
	return "(" + strings.Repeat("?,", n-1) + "?)"
}

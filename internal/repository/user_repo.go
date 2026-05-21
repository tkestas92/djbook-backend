package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/djbook/backend/internal/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

const userSelectColumns = `id, google_id, apple_id, email, username, password_hash, created_at`

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, google_id, apple_id, email, username, password_hash) VALUES (?, ?, ?, ?, ?, ?)`,
		user.ID, user.GoogleID, user.AppleID, user.Email, user.Username, user.PasswordHash,
	)
	return err
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	return r.scanUser(r.db.QueryRowContext(ctx,
		`SELECT `+userSelectColumns+` FROM users WHERE id = ?`, id,
	))
}

func (r *UserRepository) GetByGoogleID(ctx context.Context, googleID string) (*model.User, error) {
	return r.scanUser(r.db.QueryRowContext(ctx,
		`SELECT `+userSelectColumns+` FROM users WHERE google_id = ?`, googleID,
	))
}

func (r *UserRepository) GetByAppleID(ctx context.Context, appleID string) (*model.User, error) {
	return r.scanUser(r.db.QueryRowContext(ctx,
		`SELECT `+userSelectColumns+` FROM users WHERE apple_id = ?`, appleID,
	))
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return r.scanUser(r.db.QueryRowContext(ctx,
		`SELECT `+userSelectColumns+` FROM users WHERE email = ?`, email,
	))
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	return r.scanUser(r.db.QueryRowContext(ctx,
		`SELECT `+userSelectColumns+` FROM users WHERE username = ?`, username,
	))
}

func (r *UserRepository) scanUser(row *sql.Row) (*model.User, error) {
	u := &model.User{}
	err := row.Scan(&u.ID, &u.GoogleID, &u.AppleID, &u.Email, &u.Username, &u.PasswordHash, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	return u, err
}

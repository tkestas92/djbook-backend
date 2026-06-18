package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/djbook/backend/internal/model"
	"github.com/djbook/backend/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUsernameTaken       = errors.New("username already taken")
	ErrInvalidCredentials  = errors.New("invalid credentials")
)

type UserService struct {
	repo repository.UserStore
}

func NewUserService(repo repository.UserStore) *UserService {
	return &UserService{repo: repo}
}

// GetOrCreateByGoogle finds or creates a user for a Google identity.
func (s *UserService) GetOrCreateByGoogle(ctx context.Context, googleID, email string) (*model.User, error) {
	user, err := s.repo.GetByGoogleID(ctx, googleID)
	if err == nil {
		return user, nil
	}

	// Check if user already exists with this email (different provider)
	existing, err := s.repo.GetByEmail(ctx, email)
	if err == nil {
		// Email already exists — link the Google ID
		existing.GoogleID = &googleID
		return existing, nil
	}

	newUser := &model.User{
		ID:       uuid.New().String(),
		GoogleID: &googleID,
		Email:    email,
	}
	if err := s.repo.Create(ctx, newUser); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return newUser, nil
}

// GetOrCreateByApple finds or creates a user for an Apple identity.
func (s *UserService) GetOrCreateByApple(ctx context.Context, appleID, email string) (*model.User, error) {
	user, err := s.repo.GetByAppleID(ctx, appleID)
	if err == nil {
		return user, nil
	}

	existing, err := s.repo.GetByEmail(ctx, email)
	if err == nil {
		existing.AppleID = &appleID
		return existing, nil
	}

	newUser := &model.User{
		ID:      uuid.New().String(),
		AppleID: &appleID,
		Email:   email,
	}
	if err := s.repo.Create(ctx, newUser); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return newUser, nil
}

func (s *UserService) GetByID(ctx context.Context, id string) (*model.User, error) {
	return s.repo.GetByID(ctx, id)
}

// AuthEmail builds a placeholder email for username/password users.
func AuthEmail(username string) string {
	return strings.ToLower(username) + "@auth.djbook"
}

// Register creates a new user with a bcrypt-hashed password.
func (s *UserService) Register(ctx context.Context, username, password string) (*model.User, error) {
	if _, err := s.repo.GetByUsername(ctx, username); err == nil {
		return nil, ErrUsernameTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	hashStr := string(hash)

	newUser := &model.User{
		ID:           uuid.New().String(),
		Email:        AuthEmail(username),
		Username:     &username,
		PasswordHash: &hashStr,
	}
	if err := s.repo.Create(ctx, newUser); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return newUser, nil
}

// Authenticate verifies username/password credentials.
func (s *UserService) Authenticate(ctx context.Context, username, password string) (*model.User, error) {
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if user.PasswordHash == nil {
		return nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

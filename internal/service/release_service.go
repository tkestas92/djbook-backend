package service

import (
	"context"
	"fmt"

	"github.com/djbook/backend/internal/model"
	"github.com/djbook/backend/internal/repository"
	"github.com/google/uuid"
)

type ReleaseService struct {
	repo *repository.ReleaseRepository
}

func NewReleaseService(repo *repository.ReleaseRepository) *ReleaseService {
	return &ReleaseService{repo: repo}
}

func (s *ReleaseService) Create(ctx context.Context, profileID string, input model.Release) (*model.Release, error) {
	input.ID = uuid.New().String()
	input.ProfileID = profileID
	if input.Platforms == nil {
		input.Platforms = []model.ReleasePlatform{}
	}

	if err := s.repo.Create(ctx, &input); err != nil {
		return nil, fmt.Errorf("create release: %w", err)
	}
	return s.repo.GetByID(ctx, input.ID)
}

func (s *ReleaseService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *ReleaseService) List(ctx context.Context, profileID string) ([]*model.Release, error) {
	releases, err := s.repo.List(ctx, profileID)
	if err != nil {
		return nil, err
	}
	if releases == nil {
		return []*model.Release{}, nil
	}
	return releases, nil
}

func (s *ReleaseService) GetOwnerID(ctx context.Context, releaseID string) (string, error) {
	return s.repo.GetOwnerID(ctx, releaseID)
}

func (s *ReleaseService) AssertOwner(ctx context.Context, releaseID, userID string) error {
	ownerID, err := s.repo.GetOwnerID(ctx, releaseID)
	if err != nil {
		return err
	}
	if ownerID != userID {
		return fmt.Errorf("forbidden: you do not own this release")
	}
	return nil
}

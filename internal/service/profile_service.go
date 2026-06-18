package service

import (
	"context"
	"fmt"

	"github.com/djbook/backend/internal/model"
	"github.com/djbook/backend/internal/repository"
	"github.com/google/uuid"
)

type ProfileService struct {
	repo repository.ProfileStore
}

func NewProfileService(repo repository.ProfileStore) *ProfileService {
	return &ProfileService{repo: repo}
}

func (s *ProfileService) Create(ctx context.Context, userID string, input *model.DjProfile, genres []string) (*model.DjProfile, error) {
	input.ID = uuid.New().String()
	input.UserID = userID

	if err := s.repo.Create(ctx, input); err != nil {
		return nil, fmt.Errorf("create profile: %w", err)
	}
	if err := s.repo.SetGenres(ctx, input.ID, genres); err != nil {
		return nil, fmt.Errorf("set genres: %w", err)
	}
	return s.GetByID(ctx, input.ID)
}

func (s *ProfileService) Update(ctx context.Context, id string, input *model.DjProfile, genres []string) (*model.DjProfile, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	existing.DjName = input.DjName
	existing.Bio = input.Bio
	existing.Email = input.Email
	existing.ManagementContact = input.ManagementContact

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("update profile: %w", err)
	}
	if err := s.repo.SetGenres(ctx, id, genres); err != nil {
		return nil, fmt.Errorf("update genres: %w", err)
	}
	return s.GetByID(ctx, id)
}

func (s *ProfileService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *ProfileService) GetByID(ctx context.Context, id string) (*model.DjProfile, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProfileService) ListByUserID(ctx context.Context, userID string) ([]*model.DjProfile, error) {
	return s.repo.ListByUserID(ctx, userID)
}

func (s *ProfileService) GetOwnerID(ctx context.Context, profileID string) (string, error) {
	return s.repo.GetOwnerID(ctx, profileID)
}

func (s *ProfileService) GetGenres(ctx context.Context, profileID string) ([]string, error) {
	genres, err := s.repo.GetGenres(ctx, profileID)
	if err != nil {
		return nil, err
	}
	if genres == nil {
		return []string{}, nil
	}
	return genres, nil
}

// Photo management

func (s *ProfileService) AddPhoto(ctx context.Context, profileID, url string) (*model.ProfilePhoto, error) {
	photos, err := s.repo.GetPhotos(ctx, profileID)
	if err != nil {
		return nil, err
	}
	photo := &model.ProfilePhoto{
		ID:        uuid.New().String(),
		ProfileID: profileID,
		URL:       url,
		SortOrder: len(photos),
	}
	if err := s.repo.AddPhoto(ctx, photo); err != nil {
		return nil, fmt.Errorf("add photo: %w", err)
	}
	return photo, nil
}

func (s *ProfileService) DeletePhoto(ctx context.Context, id string) error {
	return s.repo.DeletePhoto(ctx, id)
}

func (s *ProfileService) GetPhotos(ctx context.Context, profileID string) ([]*model.ProfilePhoto, error) {
	photos, err := s.repo.GetPhotos(ctx, profileID)
	if err != nil {
		return nil, err
	}
	if photos == nil {
		return []*model.ProfilePhoto{}, nil
	}
	return photos, nil
}

func (s *ProfileService) GetPhotoByID(ctx context.Context, id string) (*model.ProfilePhoto, error) {
	return s.repo.GetPhotoByID(ctx, id)
}

func (s *ProfileService) ReorderPhotos(ctx context.Context, profileID string, photoIDs []string) error {
	return s.repo.ReorderPhotos(ctx, profileID, photoIDs)
}

// Social links

func (s *ProfileService) UpsertSocialLink(ctx context.Context, profileID string, platform, url string) (*model.SocialLink, error) {
	link := &model.SocialLink{
		ID:        uuid.New().String(),
		ProfileID: profileID,
		Platform:  platform,
		URL:       url,
	}
	result, err := s.repo.UpsertSocialLinkTx(ctx, link)
	if err != nil {
		return nil, fmt.Errorf("upsert social link: %w", err)
	}
	return result, nil
}

func (s *ProfileService) DeleteSocialLink(ctx context.Context, id string) error {
	return s.repo.DeleteSocialLink(ctx, id)
}

func (s *ProfileService) GetSocialLinks(ctx context.Context, profileID string) ([]*model.SocialLink, error) {
	links, err := s.repo.GetSocialLinks(ctx, profileID)
	if err != nil {
		return nil, err
	}
	if links == nil {
		return []*model.SocialLink{}, nil
	}
	return links, nil
}

// AssertOwner verifies that userID owns the given profile.
func (s *ProfileService) AssertOwner(ctx context.Context, profileID, userID string) error {
	ownerID, err := s.repo.GetOwnerID(ctx, profileID)
	if err != nil {
		return err
	}
	if ownerID != userID {
		return fmt.Errorf("forbidden: you do not own this profile")
	}
	return nil
}

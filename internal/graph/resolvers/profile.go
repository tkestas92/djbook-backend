package resolvers

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/99designs/gqlgen/graphql"
	"github.com/djbook/backend/internal/auth"
	"github.com/djbook/backend/internal/graph"
	gmodel "github.com/djbook/backend/internal/graph/model"
	"github.com/google/uuid"
)

var _ graph.MutationResolver = (*mutationResolver)(nil)

type mutationResolver struct{ *graph.Resolver }

func (r *Resolver) Mutation() graph.MutationResolver { return &mutationResolver{r.Resolver} }

func (m *mutationResolver) CreateDjProfile(ctx context.Context, input gmodel.DjProfileInput) (*gmodel.DjProfile, error) {
	userID, err := auth.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}

	p := profileFromInput(input, input.Bio)
	dbProfile, err := m.ProfileService.Create(ctx, userID, &p, input.Genres)
	if err != nil {
		return nil, fmt.Errorf("create profile: %w", err)
	}
	return toGQLProfile(ctx, dbProfile, m.Resolver)
}

func (m *mutationResolver) UpdateDjProfile(ctx context.Context, id string, input gmodel.DjProfileInput) (*gmodel.DjProfile, error) {
	userID, err := auth.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	if err := m.ProfileService.AssertOwner(ctx, id, userID); err != nil {
		return nil, err
	}

	p := profileFromInput(input, input.Bio)
	dbProfile, err := m.ProfileService.Update(ctx, id, &p, input.Genres)
	if err != nil {
		return nil, fmt.Errorf("update profile: %w", err)
	}
	return toGQLProfile(ctx, dbProfile, m.Resolver)
}

func (m *mutationResolver) DeleteDjProfile(ctx context.Context, id string) (bool, error) {
	userID, err := auth.RequireUserID(ctx)
	if err != nil {
		return false, err
	}
	if err := m.ProfileService.AssertOwner(ctx, id, userID); err != nil {
		return false, err
	}
	if err := m.ProfileService.Delete(ctx, id); err != nil {
		return false, err
	}
	return true, nil
}

func (m *mutationResolver) UploadPhoto(ctx context.Context, profileID string, file graphql.Upload) (*gmodel.ProfilePhoto, error) {
	userID, err := auth.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	if err := m.ProfileService.AssertOwner(ctx, profileID, userID); err != nil {
		return nil, err
	}

	// Read file content
	content, err := io.ReadAll(file.File)
	if err != nil {
		return nil, fmt.Errorf("read upload: %w", err)
	}
	_ = content

	// Build a unique filename and URL (in a real app, upload to S3/GCS here)
	ext := filepath.Ext(file.Filename)
	filename := uuid.New().String() + ext
	url := "/uploads/" + filename

	photo, err := m.ProfileService.AddPhoto(ctx, profileID, url)
	if err != nil {
		return nil, fmt.Errorf("add photo: %w", err)
	}
	return &gmodel.ProfilePhoto{
		ID:        photo.ID,
		URL:       photo.URL,
		SortOrder: photo.SortOrder,
		CreatedAt: photo.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}

func (m *mutationResolver) DeletePhoto(ctx context.Context, id string) (bool, error) {
	userID, err := auth.RequireUserID(ctx)
	if err != nil {
		return false, err
	}
	// Verify ownership via photo's profile
	photo, err := m.ProfileService.GetPhotoByID(ctx, id)
	if err != nil {
		return false, fmt.Errorf("photo not found: %w", err)
	}
	if err := m.ProfileService.AssertOwner(ctx, photo.ProfileID, userID); err != nil {
		return false, err
	}
	if err := m.ProfileService.DeletePhoto(ctx, id); err != nil {
		return false, err
	}
	return true, nil
}

func (m *mutationResolver) ReorderPhotos(ctx context.Context, profileID string, photoIDs []string) (bool, error) {
	userID, err := auth.RequireUserID(ctx)
	if err != nil {
		return false, err
	}
	if err := m.ProfileService.AssertOwner(ctx, profileID, userID); err != nil {
		return false, err
	}
	if err := m.ProfileService.ReorderPhotos(ctx, profileID, photoIDs); err != nil {
		return false, err
	}
	return true, nil
}

func (m *mutationResolver) UpsertSocialLink(ctx context.Context, profileID string, input gmodel.SocialLinkInput) (*gmodel.SocialLink, error) {
	userID, err := auth.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	if err := m.ProfileService.AssertOwner(ctx, profileID, userID); err != nil {
		return nil, err
	}
	link, err := m.ProfileService.UpsertSocialLink(ctx, profileID, input.Platform, input.URL)
	if err != nil {
		return nil, err
	}
	return &gmodel.SocialLink{
		ID:       link.ID,
		Platform: link.Platform,
		URL:      link.URL,
	}, nil
}

func (m *mutationResolver) DeleteSocialLink(ctx context.Context, id string) (bool, error) {
	userID, err := auth.RequireUserID(ctx)
	if err != nil {
		return false, err
	}
	_ = userID // ownership check can be added if social_link has profile_id lookup
	if err := m.ProfileService.DeleteSocialLink(ctx, id); err != nil {
		return false, err
	}
	return true, nil
}

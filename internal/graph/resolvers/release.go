package resolvers

import (
	"context"
	"fmt"

	"github.com/djbook/backend/internal/auth"
	gmodel "github.com/djbook/backend/internal/graph/model"
	"github.com/djbook/backend/internal/model"
)

func (q *queryResolver) Releases(ctx context.Context, profileID string) ([]*gmodel.Release, error) {
	releases, err := q.ReleaseService.List(ctx, profileID)
	if err != nil {
		return nil, err
	}
	return toGQLReleases(releases), nil
}

func (m *mutationResolver) CreateRelease(ctx context.Context, profileID string, input gmodel.ReleaseInput) (*gmodel.Release, error) {
	userID, err := auth.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	if err := m.ProfileService.AssertOwner(ctx, profileID, userID); err != nil {
		return nil, err
	}

	dbRelease := model.Release{
		Title:       input.Title,
		Artist:      input.Artist,
		ArtworkURL:  input.ArtworkURL,
		SongLinkURL: input.SongLinkURL,
		Platforms:   releasePlatformsFromInput(input.Platforms),
	}

	created, err := m.ReleaseService.Create(ctx, profileID, dbRelease)
	if err != nil {
		return nil, fmt.Errorf("create release: %w", err)
	}
	return toGQLRelease(created), nil
}

func (m *mutationResolver) DeleteRelease(ctx context.Context, id string) (bool, error) {
	userID, err := auth.RequireUserID(ctx)
	if err != nil {
		return false, err
	}
	if err := m.ReleaseService.AssertOwner(ctx, id, userID); err != nil {
		return false, err
	}
	if err := m.ReleaseService.Delete(ctx, id); err != nil {
		return false, err
	}
	return true, nil
}

func releasePlatformsFromInput(platforms []*gmodel.ReleasePlatformInput) []model.ReleasePlatform {
	result := make([]model.ReleasePlatform, 0, len(platforms))
	for _, p := range platforms {
		result = append(result, model.ReleasePlatform{
			Name:    p.Name,
			URL:     p.URL,
			Enabled: p.Enabled,
		})
	}
	return result
}

func toGQLReleases(releases []*model.Release) []*gmodel.Release {
	result := make([]*gmodel.Release, 0, len(releases))
	for _, r := range releases {
		result = append(result, toGQLRelease(r))
	}
	return result
}

func toGQLRelease(r *model.Release) *gmodel.Release {
	platforms := make([]*gmodel.ReleasePlatform, 0, len(r.Platforms))
	for _, p := range r.Platforms {
		platforms = append(platforms, &gmodel.ReleasePlatform{
			Name:    p.Name,
			URL:     p.URL,
			Enabled: p.Enabled,
		})
	}

	return &gmodel.Release{
		ID:          r.ID,
		Title:       r.Title,
		Artist:      r.Artist,
		ArtworkURL:  r.ArtworkURL,
		SongLinkURL: r.SongLinkURL,
		Platforms:   platforms,
		CreatedAt:   r.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

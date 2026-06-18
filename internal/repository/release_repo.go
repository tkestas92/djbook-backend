package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/djbook/backend/internal/model"
)

type ReleaseRepository struct {
	db *sql.DB
}

func NewReleaseRepository(db *sql.DB) *ReleaseRepository {
	return &ReleaseRepository{db: db}
}

func (r *ReleaseRepository) Create(ctx context.Context, release *model.Release) error {
	platformsJSON, err := json.Marshal(release.Platforms)
	if err != nil {
		return fmt.Errorf("marshal platforms: %w", err)
	}

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO releases (id, profile_id, title, artist, artwork_url, song_link_url, platforms_json)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		release.ID, release.ProfileID, release.Title, release.Artist,
		release.ArtworkURL, release.SongLinkURL, string(platformsJSON),
	)
	return err
}

func (r *ReleaseRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM releases WHERE id = ?`, id)
	return err
}

func (r *ReleaseRepository) GetByID(ctx context.Context, id string) (*model.Release, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, profile_id, title, artist, artwork_url, song_link_url, platforms_json, created_at
		 FROM releases WHERE id = ?`, id,
	)
	return r.scanRelease(row)
}

func (r *ReleaseRepository) List(ctx context.Context, profileID string) ([]*model.Release, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, profile_id, title, artist, artwork_url, song_link_url, platforms_json, created_at
		 FROM releases WHERE profile_id = ? ORDER BY created_at DESC`,
		profileID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanReleases(rows)
}

func (r *ReleaseRepository) GetOwnerID(ctx context.Context, releaseID string) (string, error) {
	var userID string
	err := r.db.QueryRowContext(ctx,
		`SELECT p.user_id FROM releases r JOIN dj_profiles p ON r.profile_id = p.id WHERE r.id = ?`,
		releaseID,
	).Scan(&userID)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("release not found")
	}
	return userID, err
}

func (r *ReleaseRepository) scanRelease(row *sql.Row) (*model.Release, error) {
	release := &model.Release{}
	var platformsJSON sql.NullString
	err := row.Scan(
		&release.ID, &release.ProfileID, &release.Title, &release.Artist,
		&release.ArtworkURL, &release.SongLinkURL, &platformsJSON, &release.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("release not found")
	}
	if err != nil {
		return nil, err
	}
	if err := decodePlatforms(platformsJSON, &release.Platforms); err != nil {
		return nil, err
	}
	return release, nil
}

func (r *ReleaseRepository) scanReleases(rows *sql.Rows) ([]*model.Release, error) {
	var releases []*model.Release
	for rows.Next() {
		release := &model.Release{}
		var platformsJSON sql.NullString
		if err := rows.Scan(
			&release.ID, &release.ProfileID, &release.Title, &release.Artist,
			&release.ArtworkURL, &release.SongLinkURL, &platformsJSON, &release.CreatedAt,
		); err != nil {
			return nil, err
		}
		if err := decodePlatforms(platformsJSON, &release.Platforms); err != nil {
			return nil, err
		}
		releases = append(releases, release)
	}
	return releases, rows.Err()
}

func decodePlatforms(platformsJSON sql.NullString, platforms *[]model.ReleasePlatform) error {
	if !platformsJSON.Valid || platformsJSON.String == "" {
		*platforms = []model.ReleasePlatform{}
		return nil
	}
	if err := json.Unmarshal([]byte(platformsJSON.String), platforms); err != nil {
		return fmt.Errorf("unmarshal platforms: %w", err)
	}
	return nil
}

package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/djbook/backend/internal/model"
)

type ProfileRepository struct {
	db *sql.DB
}

func NewProfileRepository(db *sql.DB) *ProfileRepository {
	return &ProfileRepository{db: db}
}

func (r *ProfileRepository) Create(ctx context.Context, p *model.DjProfile) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO dj_profiles (id, user_id, dj_name, bio, email, management_contact) VALUES (?, ?, ?, ?, ?, ?)`,
		p.ID, p.UserID, p.DjName, p.Bio, p.Email, p.ManagementContact,
	)
	return err
}

func (r *ProfileRepository) Update(ctx context.Context, p *model.DjProfile) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE dj_profiles SET dj_name = ?, bio = ?, email = ?, management_contact = ? WHERE id = ?`,
		p.DjName, p.Bio, p.Email, p.ManagementContact, p.ID,
	)
	return err
}

func (r *ProfileRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM dj_profiles WHERE id = ?`, id)
	return err
}

func (r *ProfileRepository) GetByID(ctx context.Context, id string) (*model.DjProfile, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, dj_name, bio, email, management_contact, created_at FROM dj_profiles WHERE id = ?`, id,
	)
	return r.scanProfile(row)
}

func (r *ProfileRepository) ListByUserID(ctx context.Context, userID string) ([]*model.DjProfile, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, dj_name, bio, email, management_contact, created_at FROM dj_profiles WHERE user_id = ?`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []*model.DjProfile
	for rows.Next() {
		p := &model.DjProfile{}
		if err := rows.Scan(&p.ID, &p.UserID, &p.DjName, &p.Bio, &p.Email, &p.ManagementContact, &p.CreatedAt); err != nil {
			return nil, err
		}
		profiles = append(profiles, p)
	}
	return profiles, rows.Err()
}

func (r *ProfileRepository) GetOwnerID(ctx context.Context, profileID string) (string, error) {
	var userID string
	err := r.db.QueryRowContext(ctx,
		`SELECT user_id FROM dj_profiles WHERE id = ?`, profileID,
	).Scan(&userID)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("profile not found")
	}
	return userID, err
}

// Genres

func (r *ProfileRepository) SetGenres(ctx context.Context, profileID string, genres []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err := tx.ExecContext(ctx, `DELETE FROM profile_genres WHERE profile_id = ?`, profileID); err != nil {
		return err
	}
	for _, genre := range genres {
		id := newUUID()
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO profile_genres (id, profile_id, genre) VALUES (?, ?, ?)`,
			id, profileID, genre,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *ProfileRepository) GetGenres(ctx context.Context, profileID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT genre FROM profile_genres WHERE profile_id = ? ORDER BY id`, profileID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []string
	for rows.Next() {
		var g string
		if err := rows.Scan(&g); err != nil {
			return nil, err
		}
		genres = append(genres, g)
	}
	return genres, rows.Err()
}

// Photos

func (r *ProfileRepository) AddPhoto(ctx context.Context, photo *model.ProfilePhoto) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO profile_photos (id, profile_id, url, sort_order) VALUES (?, ?, ?, ?)`,
		photo.ID, photo.ProfileID, photo.URL, photo.SortOrder,
	)
	return err
}

func (r *ProfileRepository) DeletePhoto(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM profile_photos WHERE id = ?`, id)
	return err
}

func (r *ProfileRepository) GetPhotos(ctx context.Context, profileID string) ([]*model.ProfilePhoto, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, profile_id, url, sort_order, created_at FROM profile_photos WHERE profile_id = ? ORDER BY sort_order`,
		profileID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var photos []*model.ProfilePhoto
	for rows.Next() {
		p := &model.ProfilePhoto{}
		if err := rows.Scan(&p.ID, &p.ProfileID, &p.URL, &p.SortOrder, &p.CreatedAt); err != nil {
			return nil, err
		}
		photos = append(photos, p)
	}
	return photos, rows.Err()
}

func (r *ProfileRepository) GetPhotoByID(ctx context.Context, id string) (*model.ProfilePhoto, error) {
	p := &model.ProfilePhoto{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, profile_id, url, sort_order, created_at FROM profile_photos WHERE id = ?`, id,
	).Scan(&p.ID, &p.ProfileID, &p.URL, &p.SortOrder, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("photo not found")
	}
	return p, err
}

func (r *ProfileRepository) ReorderPhotos(ctx context.Context, profileID string, photoIDs []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	for i, photoID := range photoIDs {
		if _, err := tx.ExecContext(ctx,
			`UPDATE profile_photos SET sort_order = ? WHERE id = ? AND profile_id = ?`,
			i, photoID, profileID,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// Social links

// UpsertSocialLink inserts or updates a social link using ON DUPLICATE KEY UPDATE.
func (r *ProfileRepository) UpsertSocialLink(ctx context.Context, link *model.SocialLink) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO social_links (id, profile_id, platform, url)
		 VALUES (?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE url = VALUES(url)`,
		link.ID, link.ProfileID, link.Platform, link.URL,
	)
	return err
}

func (r *ProfileRepository) UpsertSocialLinkTx(ctx context.Context, link *model.SocialLink) (*model.SocialLink, error) {
	existing := &model.SocialLink{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, profile_id, platform, url FROM social_links WHERE profile_id = ? AND url = ?`,
		link.ProfileID, link.URL,
	).Scan(&existing.ID, &existing.ProfileID, &existing.Platform, &existing.URL)
	if err == nil {
		return existing, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	if _, err := r.db.ExecContext(ctx,
		`INSERT INTO social_links (id, profile_id, platform, url) VALUES (?, ?, ?, ?)`,
		link.ID, link.ProfileID, link.Platform, link.URL,
	); err != nil {
		return nil, err
	}
	return link, nil
}

func (r *ProfileRepository) DeleteSocialLink(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM social_links WHERE id = ?`, id)
	return err
}

func (r *ProfileRepository) GetSocialLinks(ctx context.Context, profileID string) ([]*model.SocialLink, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, profile_id, platform, url FROM social_links WHERE profile_id = ?`, profileID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []*model.SocialLink
	for rows.Next() {
		l := &model.SocialLink{}
		if err := rows.Scan(&l.ID, &l.ProfileID, &l.Platform, &l.URL); err != nil {
			return nil, err
		}
		links = append(links, l)
	}
	return links, rows.Err()
}

func (r *ProfileRepository) scanProfile(row *sql.Row) (*model.DjProfile, error) {
	p := &model.DjProfile{}
	err := row.Scan(&p.ID, &p.UserID, &p.DjName, &p.Bio, &p.Email, &p.ManagementContact, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("profile not found")
	}
	return p, err
}

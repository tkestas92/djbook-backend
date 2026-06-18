package repository

import (
	"context"

	"github.com/djbook/backend/internal/model"
)

// UserStore defines persistence operations for users.
type UserStore interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id string) (*model.User, error)
	GetByGoogleID(ctx context.Context, googleID string) (*model.User, error)
	GetByAppleID(ctx context.Context, appleID string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
}

// ProfileStore defines persistence operations for DJ profiles.
type ProfileStore interface {
	Create(ctx context.Context, p *model.DjProfile) error
	Update(ctx context.Context, p *model.DjProfile) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*model.DjProfile, error)
	ListByUserID(ctx context.Context, userID string) ([]*model.DjProfile, error)
	GetOwnerID(ctx context.Context, profileID string) (string, error)
	SetGenres(ctx context.Context, profileID string, genres []string) error
	GetGenres(ctx context.Context, profileID string) ([]string, error)
	AddPhoto(ctx context.Context, photo *model.ProfilePhoto) error
	DeletePhoto(ctx context.Context, id string) error
	GetPhotos(ctx context.Context, profileID string) ([]*model.ProfilePhoto, error)
	GetPhotoByID(ctx context.Context, id string) (*model.ProfilePhoto, error)
	ReorderPhotos(ctx context.Context, profileID string, photoIDs []string) error
	UpsertSocialLinkTx(ctx context.Context, link *model.SocialLink) (*model.SocialLink, error)
	DeleteSocialLink(ctx context.Context, id string) error
	GetSocialLinks(ctx context.Context, profileID string) ([]*model.SocialLink, error)
}

// EventStore defines persistence operations for events.
type EventStore interface {
	Create(ctx context.Context, e *model.Event) error
	Update(ctx context.Context, e *model.Event) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*model.Event, error)
	GetCurrentStatus(ctx context.Context, id string) (model.EventStatus, error)
	List(ctx context.Context, filter EventFilter) ([]*model.Event, error)
	ListUpcoming(ctx context.Context, profileID string) ([]*model.Event, error)
	AddStatusHistory(ctx context.Context, h *model.EventStatusHistory) error
	GetStatusHistory(ctx context.Context, eventID string) ([]*model.EventStatusHistory, error)
	GetOwnerID(ctx context.Context, eventID string) (string, error)
}

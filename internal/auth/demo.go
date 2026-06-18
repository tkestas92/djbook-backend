package auth

import (
	"context"
	"errors"

	"github.com/djbook/backend/internal/model"
	"github.com/djbook/backend/internal/service"
)

const DemoUsername = "kantrybes"

var ErrDemoWriteForbidden = errors.New("Demo account changes are not saved. Sign up to save your data!")

var demoBlockedMutations = map[string]struct{}{
	"createEvent":      {},
	"updateEvent":      {},
	"deleteEvent":      {},
	"createDjProfile":  {},
	"updateDjProfile":  {},
	"uploadPhoto":      {},
	"upsertSocialLink": {},
	"deleteSocialLink": {},
	"createRelease":    {},
	"deleteRelease":    {},
}

func IsDemoUser(user *model.User) bool {
	return user.Username != nil && *user.Username == DemoUsername
}

func IsBlockedDemoMutation(fieldName string) bool {
	_, ok := demoBlockedMutations[fieldName]
	return ok
}

// CheckDemoWrite returns ErrDemoWriteForbidden when the authenticated user is the demo account.
func CheckDemoWrite(ctx context.Context, userSvc *service.UserService) error {
	userID := GetUserID(ctx)
	if userID == "" {
		return nil
	}

	user, err := userSvc.GetByID(ctx, userID)
	if err != nil {
		return nil
	}

	if IsDemoUser(user) {
		return ErrDemoWriteForbidden
	}
	return nil
}

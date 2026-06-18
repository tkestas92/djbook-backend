package repository_test

import (
	"context"
	"testing"

	"github.com/djbook/backend/internal/model"
	"github.com/djbook/backend/internal/repository"
	"github.com/djbook/backend/internal/service"
)

func TestCreateProfile_DuplicateDjName(t *testing.T) {
	// The dj_profiles schema has no unique constraint on dj_name per user.
	// Duplicate DJ names for the same user are currently allowed.
	store := repository.NewMemoryProfileStore()
	svc := service.NewProfileService(store)
	ctx := context.Background()
	userID := "user-1"
	djName := "Same Name"

	first, err := svc.Create(ctx, userID, &model.DjProfile{DjName: djName}, nil)
	if err != nil {
		t.Fatalf("first create: %v", err)
	}

	second, err := svc.Create(ctx, userID, &model.DjProfile{DjName: djName}, nil)
	if err != nil {
		t.Fatalf("second create with duplicate djName should succeed, got error: %v", err)
	}
	if second.ID == first.ID {
		t.Fatal("expected distinct profile ids for duplicate djName")
	}
	if second.DjName != djName {
		t.Fatalf("djName = %q, want %q", second.DjName, djName)
	}

	profiles, err := svc.ListByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("list profiles: %v", err)
	}
	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles for same user, got %d", len(profiles))
	}
}

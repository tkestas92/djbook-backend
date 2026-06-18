package auth

import (
	"context"
	"errors"
)

var ErrDemoWriteForbidden = errors.New("Demo account changes are not saved. Sign up to save your data!")

// CheckDemoWrite returns ErrDemoWriteForbidden when the session was issued via the demo login flow.
func CheckDemoWrite(ctx context.Context) error {
	if IsDemoSession(ctx) {
		return ErrDemoWriteForbidden
	}
	return nil
}

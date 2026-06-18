package graph

import (
	"github.com/djbook/backend/internal/service"
)

// Resolver is the root GraphQL resolver that holds all services.
type Resolver struct {
	UserService    *service.UserService
	ProfileService *service.ProfileService
	EventService   *service.EventService
	ReleaseService *service.ReleaseService
	FinanceService *service.FinanceService
}

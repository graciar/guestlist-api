package user

import (
	"github.com/graciar/guestlist-api/internal/event"
)

type User struct {
	ID    string
	Name  string
	Email string
}

type UpdateUserParams struct {
	ID    string
	Name  string
	Email string
}

type DeleteUserParams struct {
	ID string
}

type UserEvents struct {
	User   User
	Events []event.Event
}

type Service interface {
	// GetUser(ctx context.Context, id string) (User, error)
	// UpdateUser(ctx context.Context, user User) (User, error)
	// DeleteUser(ctx context.Context, id string) error
	// GetUserEventStats(ctx context.Context, id string) (User, error)
}

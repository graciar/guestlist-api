package auth

import "context"

// Custom unexported type for context keys avoids collisions
type contextKey string

const userIDKey contextKey = "userId"
const userRoleKey contextKey = "userRole"

// WithUserID injects the user ID into the context (used by Middleware)
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func WithUserRole(ctx context.Context, userRole string) context.Context {
	return context.WithValue(ctx, userRoleKey, userRole)
}

// UserIDFromContext retrieves the user ID from the context (used by Service)
func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}

func UserRoleFromContext(ctx context.Context) (string, bool) {
	userRole, ok := ctx.Value(userRoleKey).(string)
	return userRole, ok
}

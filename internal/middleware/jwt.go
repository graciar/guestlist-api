package middleware

import (
	"net/http"
	"strings"

	"github.com/graciar/guestlist-api/internal/auth"
)

func Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		// Split "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Call your existing validation function
		claims, err := auth.ValidateToken(tokenString, "access")
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// Inject the User ID (assuming your SignedDetails struct has an ID or UserID field)
		ctx := auth.WithUserID(r.Context(), claims.ID)
		ctx = auth.WithUserRole(ctx, claims.Role)

		// Pass the enriched context down the line
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

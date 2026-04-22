package middleware

import (
	"context"
	"net/http"

	"github.com/username/app-recrutamento-ia/internal/auth"
)

type contextKey string

const ClaimsContextKey contextKey = "user_claims"

// RequireAuth protects routes checking for a valid JWT in the cookie.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Attempt to read JWT from HttpOnly cookie
		cookie, err := r.Cookie("jwt_token")
		if err != nil {
			// Not authenticated
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		claims, err := auth.ValidateToken(cookie.Value)
		if err != nil {
			// Token invalid or expired
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// Inject claims into context
		ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

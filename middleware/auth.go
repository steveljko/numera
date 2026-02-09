package middleware

import (
	"context"
	"net/http"
	"numera/pkg/session"
)

// RequireAuth prevents unauthorized users from accessing protected routes.
func RequireAuth(sessionMgr *session.Session) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := sessionMgr.GetUserID(r)

			if userID == 0 {
				if r.Header.Get("HX-Redirect") == "true" {
					w.Header().Set("HX-Redirect", "/login")
				} else {
					http.Redirect(w, r, "/login", http.StatusSeeOther)
				}
				return
			}

			ctx := context.WithValue(r.Context(), "USER_ID", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireGuest redirects authenticated users away from auth pages
func RequireGuest(sessionMgr *session.Session) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := sessionMgr.GetUserID(r)

			if userID != 0 {
				http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

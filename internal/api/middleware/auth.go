package middleware

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/virajchitnis/linux-webui/internal/auth"
)

type contextKey string

const SessionKey contextKey = "session"

type Options struct {
	DB             *sql.DB
	TimeoutMinutes int
}

func RequireAuth(opts Options) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := auth.SessionIDFromRequest(r)
			if id == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			session, err := auth.GetSession(opts.DB, id, opts.TimeoutMinutes)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), SessionKey, session)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, ok := r.Context().Value(SessionKey).(*auth.Session)
		if !ok || session.Role != "admin" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RequireCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		if !auth.ValidateCSRF(r) {
			http.Error(w, "invalid csrf token", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func SessionFromContext(ctx context.Context) *auth.Session {
	s, _ := ctx.Value(SessionKey).(*auth.Session)
	return s
}

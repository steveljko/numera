package middleware

import (
	"context"
	"github.com/sirupsen/logrus"
	"net/http"
)

type loggerKey struct{}

// WithLogger injects the logger into the request context.
func WithLogger(log *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			entry := log.WithFields(logrus.Fields{
				"path":   r.URL.Path,
				"method": r.Method,
				"ip":     r.RemoteAddr,
			})

			ctx := context.WithValue(r.Context(), loggerKey{}, entry)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetLogger retrieves the logger from the context.
func GetLogger(ctx context.Context) *logrus.Entry {
	if entry, ok := ctx.Value(loggerKey{}).(*logrus.Entry); ok {
		return entry
	}
	return logrus.NewEntry(logrus.New())
}

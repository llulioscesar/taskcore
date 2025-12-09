package server

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
)

type ctxKey string

const (
	CtxUserID    ctxKey = "user_id"
	CtxRequestID ctxKey = "request_id"
)

// Middleware is a function that wraps an http.Handler
type Middleware func(http.Handler) http.Handler

// Chain applies middlewares in order
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// RequestID adds a unique request ID to context
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = uuid.New().String()
		}
		ctx := context.WithValue(r.Context(), CtxRequestID, id)
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Logger logs request details
func Logger(log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(ww, r)

			log.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.status,
				"duration", time.Since(start),
				"request_id", r.Context().Value(CtxRequestID),
			)
		})
	}
}

// Recoverer recovers from panics
func Recoverer(log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Error("panic recovered",
						"error", err,
						"stack", string(debug.Stack()),
						"request_id", r.Context().Value(CtxRequestID),
					)
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// CORS adds CORS headers
func CORS(allowedOrigins []string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowed := false
			for _, o := range allowedOrigins {
				if o == "*" || o == origin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
				w.Header().Set("Access-Control-Max-Age", "86400")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(CtxUserID).(uuid.UUID)
	return id, ok
}

// GetRequestID extracts request ID from context
func GetRequestID(ctx context.Context) string {
	id, _ := ctx.Value(CtxRequestID).(string)
	return id
}

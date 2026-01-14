package wide_events

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
)

type contextKey struct{}

// WideEvent holds the attributes collected during a request lifecycle.
type WideEvent struct {
	mu    sync.Mutex
	attrs []slog.Attr
}

// Add appends new attributes to the event.
func (e *WideEvent) Add(attrs ...slog.Attr) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.attrs = append(e.attrs, attrs...)
}

// NewContext returns a new context containing a WideEvent.
func NewContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey{}, &WideEvent{})
}

// FromContext retrieves the WideEvent from the context.
func FromContext(ctx context.Context) *WideEvent {
	if e, ok := ctx.Value(contextKey{}).(*WideEvent); ok {
		return e
	}
	return nil
}

// Record is a helper to add attributes to the context's WideEvent.
func Record(ctx context.Context, attrs ...slog.Attr) {
	if e := FromContext(ctx); e != nil {
		e.Add(attrs...)
	}
}

func WideEventMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Initialize the event container in the context
		ctx := NewContext(r.Context())
		event := FromContext(ctx)

		// 2. Add initial request metadata
		event.Add(
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("request_id", r.Header.Get("X-Request-ID")),
			slog.Group()
		)

		// Create a custom response writer to capture status code
		sw := &statusWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// To emit log line:
		slog.LogAttrs(ctx, slog.LevelInfo, "request_completed", event.attrs...)

		next.ServeHTTP(sw, r.WithContext(ctx))

	})
}

type statusWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

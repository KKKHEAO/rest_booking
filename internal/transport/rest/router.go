package rest

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/KKKHEAO/rest_booking/internal/service"
)

// API держит зависимости HTTP-слоя.
type API struct {
	svc *service.Service
	log *slog.Logger
}

// NewRouter собирает chi-роутер со всеми ручками.
func NewRouter(svc *service.Service, log *slog.Logger) http.Handler {
	a := &API{svc: svc, log: log}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(requestLogger(log))

	r.Route("/bookings", func(r chi.Router) {
		r.Post("/", a.createBooking)
		r.Get("/", a.listBookings)
		r.Get("/{id}", a.getBooking)
		r.Delete("/{id}", a.cancelBooking)
	})

	r.Route("/tables", func(r chi.Router) {
		r.Post("/", a.createTable)
		r.Get("/", a.listTables)
		r.Get("/{id}", a.getTable)
	})

	return r
}

// requestLogger логирует каждый запрос через slog.
func requestLogger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			log.InfoContext(r.Context(), "http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"bytes", ww.BytesWritten(),
				"duration", time.Since(start).String(),
				"request_id", middleware.GetReqID(r.Context()),
			)
		})
	}
}

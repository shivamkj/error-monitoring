package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shivam/error-monitoring/backend/internal/config"
	"github.com/shivam/error-monitoring/backend/internal/handlers"
	"github.com/shivam/error-monitoring/backend/internal/handlers/dashboard"
	"github.com/shivam/error-monitoring/backend/internal/handlers/ingest"
	"github.com/shivam/error-monitoring/backend/internal/middleware"
)

func New(pool *pgxpool.Pool, cfg *config.Config) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logging)
	r.Use(middleware.CORS(cfg.AllowedOrigins))

	ingestHandler := ingest.NewHandler(pool)
	authHandler := dashboard.NewAuthHandler(pool, cfg.JWTSecret)
	projectsHandler := dashboard.NewProjectsHandler(pool, cfg)
	issuesHandler := dashboard.NewIssuesHandler(pool)
	eventsHandler := dashboard.NewEventsHandler(pool)

	// Health check
	r.Get("/api/health", handlers.HealthCheck(pool))

	// Sentry-compatible ingestion endpoints
	r.Post("/api/{projectID}/envelope/", ingestHandler.HandleEnvelope)
	r.Post("/api/{projectID}/store/", ingestHandler.HandleStore)
	// Without trailing slash
	r.Post("/api/{projectID}/envelope", ingestHandler.HandleEnvelope)
	r.Post("/api/{projectID}/store", ingestHandler.HandleStore)

	// Dashboard API
	r.Route("/api/dashboard", func(r chi.Router) {
		// Public auth routes
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth(cfg.JWTSecret))

			r.Get("/auth/me", authHandler.Me)

			r.Get("/projects", projectsHandler.List)
			r.Post("/projects", projectsHandler.Create)
			r.Get("/projects/{id}", projectsHandler.Get)
			r.Delete("/projects/{id}", projectsHandler.Delete)

			r.Get("/projects/{projectID}/issues", issuesHandler.List)
			r.Get("/issues/{id}", issuesHandler.Get)
			r.Put("/issues/{id}/status", issuesHandler.UpdateStatus)
			r.Delete("/issues/{id}", issuesHandler.Delete)

			r.Get("/issues/{id}/events", eventsHandler.ListForIssue)
			r.Get("/events/{id}", eventsHandler.Get)
		})
	})

	return r
}

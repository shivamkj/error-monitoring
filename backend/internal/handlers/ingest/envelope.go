package ingest

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shivam/error-monitoring/backend/internal/processing"
)

type Handler struct {
	pool     *pgxpool.Pool
	pipeline *processing.Pipeline
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{
		pool:     pool,
		pipeline: processing.NewPipeline(pool),
	}
}

func (h *Handler) HandleEnvelope(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "projectID")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid project id"}`, http.StatusBadRequest)
		return
	}

	sentryKey := ExtractSentryKey(r)
	if sentryKey != "" {
		if _, err := ValidateProjectKey(r.Context(), h.pool, sentryKey, projectID); err != nil {
			http.Error(w, `{"error":"invalid project key"}`, http.StatusUnauthorized)
			return
		}
	}

	var reader io.Reader = r.Body
	if r.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, `{"error":"invalid gzip"}`, http.StatusBadRequest)
			return
		}
		defer gz.Close()
		reader = gz
	}

	body, err := io.ReadAll(io.LimitReader(reader, 1<<20)) // 1MB limit
	if err != nil {
		http.Error(w, `{"error":"failed to read body"}`, http.StatusBadRequest)
		return
	}

	event, err := processing.ParseEnvelope(body)
	if err != nil {
		log.Printf("envelope parse error: %v", err)
		http.Error(w, `{"error":"failed to parse envelope"}`, http.StatusBadRequest)
		return
	}

	if err := h.pipeline.Process(r.Context(), projectID, event); err != nil {
		log.Printf("pipeline error: %v", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"id": event.EventID})
}

func (h *Handler) HandleStore(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "projectID")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid project id"}`, http.StatusBadRequest)
		return
	}

	sentryKey := ExtractSentryKey(r)
	if sentryKey != "" {
		if _, err := ValidateProjectKey(r.Context(), h.pool, sentryKey, projectID); err != nil {
			http.Error(w, `{"error":"invalid project key"}`, http.StatusUnauthorized)
			return
		}
	}

	var reader io.Reader = r.Body
	if r.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, `{"error":"invalid gzip"}`, http.StatusBadRequest)
			return
		}
		defer gz.Close()
		reader = gz
	}

	body, err := io.ReadAll(io.LimitReader(reader, 1<<20))
	if err != nil {
		http.Error(w, `{"error":"failed to read body"}`, http.StatusBadRequest)
		return
	}

	event, err := processing.ParseStorePayload(body)
	if err != nil {
		log.Printf("store parse error: %v", err)
		http.Error(w, `{"error":"failed to parse event"}`, http.StatusBadRequest)
		return
	}

	if err := h.pipeline.Process(r.Context(), projectID, event); err != nil {
		log.Printf("pipeline error: %v", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"id": event.EventID})
}

package dashboard

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shivam/error-monitoring/backend/internal/middleware"
	"github.com/shivam/error-monitoring/backend/internal/models"
)

type EventsHandler struct {
	pool *pgxpool.Pool
}

func NewEventsHandler(pool *pgxpool.Pool) *EventsHandler {
	return &EventsHandler{pool: pool}
}

func (h *EventsHandler) ListForIssue(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	issueID := chi.URLParam(r, "id")

	var exists bool
	h.pool.QueryRow(r.Context(), `
		SELECT EXISTS(
			SELECT 1 FROM issues i JOIN projects p ON p.id = i.project_id
			WHERE i.id = $1 AND p.user_id = $2
		)`, issueID, userID).Scan(&exists)

	if !exists {
		http.Error(w, `{"error":"issue not found"}`, http.StatusNotFound)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}
	offset := (page - 1) * perPage

	rows, err := h.pool.Query(r.Context(), `
		SELECT id, event_id, issue_id, project_id, timestamp, level, platform,
			   ip_address, user_data, request_data, breadcrumbs, contexts, tags,
			   exception, message, environment, release_tag, server_name, created_at
		FROM events WHERE issue_id = $1
		ORDER BY timestamp DESC LIMIT $2 OFFSET $3`,
		issueID, perPage, offset)
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var e models.Event
		var ipAddr *string
		if err := rows.Scan(&e.ID, &e.EventID, &e.IssueID, &e.ProjectID, &e.Timestamp,
			&e.Level, &e.Platform, &ipAddr, &e.UserData, &e.RequestData,
			&e.Breadcrumbs, &e.Contexts, &e.Tags, &e.Exception, &e.Message,
			&e.Environment, &e.ReleaseTag, &e.ServerName, &e.CreatedAt); err != nil {
			continue
		}
		if ipAddr != nil {
			e.IPAddress = *ipAddr
		}
		events = append(events, e)
	}

	if events == nil {
		events = []models.Event{}
	}

	var total int
	h.pool.QueryRow(r.Context(), "SELECT COUNT(*) FROM events WHERE issue_id = $1", issueID).Scan(&total)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events": events,
		"total":  total,
		"page":   page,
	})
}

func (h *EventsHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	eventID := chi.URLParam(r, "id")

	var e models.Event
	var ipAddr *string
	err := h.pool.QueryRow(r.Context(), `
		SELECT e.id, e.event_id, e.issue_id, e.project_id, e.timestamp, e.level, e.platform,
			   e.ip_address, e.user_data, e.request_data, e.breadcrumbs, e.contexts, e.tags,
			   e.exception, e.message, e.environment, e.release_tag, e.server_name, e.raw_payload, e.created_at
		FROM events e
		JOIN projects p ON p.id = e.project_id
		WHERE e.id = $1 AND p.user_id = $2`,
		eventID, userID,
	).Scan(&e.ID, &e.EventID, &e.IssueID, &e.ProjectID, &e.Timestamp, &e.Level, &e.Platform,
		&ipAddr, &e.UserData, &e.RequestData, &e.Breadcrumbs, &e.Contexts, &e.Tags,
		&e.Exception, &e.Message, &e.Environment, &e.ReleaseTag, &e.ServerName, &e.RawPayload, &e.CreatedAt)

	if err != nil {
		http.Error(w, `{"error":"event not found"}`, http.StatusNotFound)
		return
	}
	if ipAddr != nil {
		e.IPAddress = *ipAddr
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(e)
}

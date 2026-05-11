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

type IssuesHandler struct {
	pool *pgxpool.Pool
}

func NewIssuesHandler(pool *pgxpool.Pool) *IssuesHandler {
	return &IssuesHandler{pool: pool}
}

func (h *IssuesHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	projectID := chi.URLParam(r, "projectID")

	var exists bool
	h.pool.QueryRow(r.Context(),
		"SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1 AND user_id = $2)",
		projectID, userID).Scan(&exists)
	if !exists {
		http.Error(w, `{"error":"project not found"}`, http.StatusNotFound)
		return
	}

	status := r.URL.Query().Get("status")
	sort := r.URL.Query().Get("sort")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	orderBy := "last_seen DESC"
	switch sort {
	case "first_seen":
		orderBy = "first_seen DESC"
	case "event_count":
		orderBy = "event_count DESC"
	case "level":
		orderBy = "level ASC"
	}

	offset := (page - 1) * perPage

	var query string
	var args []interface{}

	if status != "" {
		query = `SELECT id, project_id, fingerprint, title, culprit, level, platform, status,
				 first_seen, last_seen, event_count, browsers, os_names, devices, urls, created_at, updated_at
				 FROM issues WHERE project_id = $1 AND status = $2 ORDER BY ` + orderBy + ` LIMIT $3 OFFSET $4`
		args = []interface{}{projectID, status, perPage, offset}
	} else {
		query = `SELECT id, project_id, fingerprint, title, culprit, level, platform, status,
				 first_seen, last_seen, event_count, browsers, os_names, devices, urls, created_at, updated_at
				 FROM issues WHERE project_id = $1 ORDER BY ` + orderBy + ` LIMIT $2 OFFSET $3`
		args = []interface{}{projectID, perPage, offset}
	}

	rows, err := h.pool.Query(r.Context(), query, args...)
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var issues []models.Issue
	for rows.Next() {
		var issue models.Issue
		if err := rows.Scan(&issue.ID, &issue.ProjectID, &issue.Fingerprint, &issue.Title,
			&issue.Culprit, &issue.Level, &issue.Platform, &issue.Status,
			&issue.FirstSeen, &issue.LastSeen, &issue.EventCount,
			&issue.Browsers, &issue.OsNames, &issue.Devices, &issue.URLs,
			&issue.CreatedAt, &issue.UpdatedAt); err != nil {
			continue
		}
		issues = append(issues, issue)
	}

	if issues == nil {
		issues = []models.Issue{}
	}

	var total int
	if status != "" {
		h.pool.QueryRow(r.Context(),
			"SELECT COUNT(*) FROM issues WHERE project_id = $1 AND status = $2",
			projectID, status).Scan(&total)
	} else {
		h.pool.QueryRow(r.Context(),
			"SELECT COUNT(*) FROM issues WHERE project_id = $1",
			projectID).Scan(&total)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"issues": issues,
		"total":  total,
		"page":   page,
	})
}

func (h *IssuesHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	issueID := chi.URLParam(r, "id")

	var issue models.Issue
	err := h.pool.QueryRow(r.Context(), `
		SELECT i.id, i.project_id, i.fingerprint, i.title, i.culprit, i.level, i.platform, i.status,
			   i.first_seen, i.last_seen, i.event_count, i.browsers, i.os_names, i.devices, i.urls,
			   i.created_at, i.updated_at
		FROM issues i
		JOIN projects p ON p.id = i.project_id
		WHERE i.id = $1 AND p.user_id = $2`,
		issueID, userID,
	).Scan(&issue.ID, &issue.ProjectID, &issue.Fingerprint, &issue.Title,
		&issue.Culprit, &issue.Level, &issue.Platform, &issue.Status,
		&issue.FirstSeen, &issue.LastSeen, &issue.EventCount,
		&issue.Browsers, &issue.OsNames, &issue.Devices, &issue.URLs,
		&issue.CreatedAt, &issue.UpdatedAt)

	if err != nil {
		http.Error(w, `{"error":"issue not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(issue)
}

func (h *IssuesHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	issueID := chi.URLParam(r, "id")

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	validStatuses := map[string]bool{"unresolved": true, "resolved": true, "ignored": true}
	if !validStatuses[req.Status] {
		http.Error(w, `{"error":"invalid status"}`, http.StatusBadRequest)
		return
	}

	result, err := h.pool.Exec(r.Context(), `
		UPDATE issues SET status = $1, updated_at = NOW()
		FROM projects
		WHERE issues.id = $2 AND issues.project_id = projects.id AND projects.user_id = $3`,
		req.Status, issueID, userID)

	if err != nil || result.RowsAffected() == 0 {
		http.Error(w, `{"error":"issue not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": req.Status})
}

func (h *IssuesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	issueID := chi.URLParam(r, "id")

	result, err := h.pool.Exec(r.Context(), `
		DELETE FROM issues
		USING projects
		WHERE issues.id = $1 AND issues.project_id = projects.id AND projects.user_id = $2`,
		issueID, userID)

	if err != nil || result.RowsAffected() == 0 {
		http.Error(w, `{"error":"issue not found"}`, http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

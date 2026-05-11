package dashboard

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shivam/error-monitoring/backend/internal/config"
	"github.com/shivam/error-monitoring/backend/internal/middleware"
	"github.com/shivam/error-monitoring/backend/internal/models"
)

type ProjectsHandler struct {
	pool *pgxpool.Pool
	cfg  *config.Config
}

func NewProjectsHandler(pool *pgxpool.Pool, cfg *config.Config) *ProjectsHandler {
	return &ProjectsHandler{pool: pool, cfg: cfg}
}

type createProjectRequest struct {
	Name     string `json:"name"`
	Platform string `json:"platform"`
}

type projectResponse struct {
	models.Project
	DSN string `json:"dsn"`
}

func (h *ProjectsHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)

	rows, err := h.pool.Query(r.Context(),
		`SELECT id, name, slug, platform, public_key, user_id, created_at, updated_at
		 FROM projects WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var projects []projectResponse
	for rows.Next() {
		var p models.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Platform, &p.PublicKey, &p.UserID, &p.CreatedAt, &p.UpdatedAt); err != nil {
			continue
		}
		projects = append(projects, projectResponse{Project: p, DSN: h.buildDSN(p)})
	}

	if projects == nil {
		projects = []projectResponse{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

func (h *ProjectsHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)

	var req createProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}

	if req.Platform == "" {
		req.Platform = "javascript"
	}

	publicKey := generatePublicKey()
	slug := slugify(req.Name)

	var p models.Project
	err := h.pool.QueryRow(r.Context(),
		`INSERT INTO projects (name, slug, platform, public_key, user_id)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, name, slug, platform, public_key, user_id, created_at, updated_at`,
		req.Name, slug, req.Platform, publicKey, userID,
	).Scan(&p.ID, &p.Name, &p.Slug, &p.Platform, &p.PublicKey, &p.UserID, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		http.Error(w, `{"error":"failed to create project"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(projectResponse{Project: p, DSN: h.buildDSN(p)})
}

func (h *ProjectsHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	projectID := chi.URLParam(r, "id")

	var p models.Project
	err := h.pool.QueryRow(r.Context(),
		`SELECT id, name, slug, platform, public_key, user_id, created_at, updated_at
		 FROM projects WHERE id = $1 AND user_id = $2`, projectID, userID,
	).Scan(&p.ID, &p.Name, &p.Slug, &p.Platform, &p.PublicKey, &p.UserID, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		http.Error(w, `{"error":"project not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projectResponse{Project: p, DSN: h.buildDSN(p)})
}

func (h *ProjectsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	projectID := chi.URLParam(r, "id")

	result, err := h.pool.Exec(r.Context(),
		"DELETE FROM projects WHERE id = $1 AND user_id = $2", projectID, userID)
	if err != nil || result.RowsAffected() == 0 {
		http.Error(w, `{"error":"project not found"}`, http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProjectsHandler) buildDSN(p models.Project) string {
	baseURL := strings.TrimRight(h.cfg.BaseURL, "/")
	baseURL = strings.TrimPrefix(baseURL, "http://")
	baseURL = strings.TrimPrefix(baseURL, "https://")

	protocol := "http"
	if strings.Contains(h.cfg.BaseURL, "https://") {
		protocol = "https"
	}

	return protocol + "://" + p.PublicKey + "@" + baseURL + "/" + p.ID.String()
}

func generatePublicKey() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func slugify(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	var result []byte
	for _, c := range []byte(slug) {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			result = append(result, c)
		}
	}
	return string(result)
}

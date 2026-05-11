package ingest

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func ExtractSentryKey(r *http.Request) string {
	authHeader := r.Header.Get("X-Sentry-Auth")
	if authHeader != "" {
		for _, part := range strings.Split(authHeader, ",") {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "sentry_key=") {
				return strings.TrimPrefix(part, "sentry_key=")
			}
			if strings.HasPrefix(part, "Sentry sentry_key=") {
				return strings.TrimPrefix(part, "Sentry sentry_key=")
			}
		}
	}

	if key := r.URL.Query().Get("sentry_key"); key != "" {
		return key
	}

	return ""
}

func ValidateProjectKey(ctx context.Context, pool *pgxpool.Pool, publicKey string, projectID uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID
	err := pool.QueryRow(ctx,
		"SELECT id FROM projects WHERE public_key = $1 AND id = $2",
		publicKey, projectID,
	).Scan(&id)
	return id, err
}

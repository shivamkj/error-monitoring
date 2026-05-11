package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Issue struct {
	ID          uuid.UUID       `json:"id"`
	ProjectID   uuid.UUID       `json:"project_id"`
	Fingerprint string          `json:"fingerprint"`
	Title       string          `json:"title"`
	Culprit     string          `json:"culprit"`
	Level       string          `json:"level"`
	Platform    string          `json:"platform"`
	Status      string          `json:"status"`
	FirstSeen   time.Time       `json:"first_seen"`
	LastSeen    time.Time       `json:"last_seen"`
	EventCount  int             `json:"event_count"`
	Browsers    json.RawMessage `json:"browsers"`
	OsNames     json.RawMessage `json:"os_names"`
	Devices     json.RawMessage `json:"devices"`
	URLs        json.RawMessage `json:"urls"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

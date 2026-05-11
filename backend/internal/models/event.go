package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID          uuid.UUID       `json:"id"`
	EventID     string          `json:"event_id"`
	IssueID     uuid.UUID       `json:"issue_id"`
	ProjectID   uuid.UUID       `json:"project_id"`
	Timestamp   time.Time       `json:"timestamp"`
	Level       string          `json:"level"`
	Platform    string          `json:"platform"`
	IPAddress   string          `json:"ip_address,omitempty"`
	UserData    json.RawMessage `json:"user_data,omitempty"`
	RequestData json.RawMessage `json:"request_data,omitempty"`
	Breadcrumbs json.RawMessage `json:"breadcrumbs,omitempty"`
	Contexts    json.RawMessage `json:"contexts,omitempty"`
	Tags        json.RawMessage `json:"tags,omitempty"`
	Exception   json.RawMessage `json:"exception,omitempty"`
	Message     string          `json:"message,omitempty"`
	Environment string          `json:"environment,omitempty"`
	ReleaseTag  string          `json:"release_tag,omitempty"`
	ServerName  string          `json:"server_name,omitempty"`
	RawPayload  json.RawMessage `json:"raw_payload,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

package processing

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Pipeline struct {
	pool       *pgxpool.Pool
	aggregator *Aggregator
}

func NewPipeline(pool *pgxpool.Pool) *Pipeline {
	return &Pipeline{
		pool:       pool,
		aggregator: NewAggregator(pool),
	}
}

func (p *Pipeline) Process(ctx context.Context, projectID uuid.UUID, event *ParsedEvent) error {
	normalized := Normalize(event)
	fingerprint := GenerateFingerprint(event, normalized)
	title := buildTitle(event)
	culprit := buildCulprit(event)

	issueID, err := p.upsertIssue(ctx, projectID, fingerprint, title, culprit, event)
	if err != nil {
		return fmt.Errorf("upsert issue: %w", err)
	}

	if err := p.aggregator.UpdateAggregates(ctx, issueID, event); err != nil {
		return fmt.Errorf("update aggregates: %w", err)
	}

	if err := p.storeEvent(ctx, issueID, projectID, event); err != nil {
		return fmt.Errorf("store event: %w", err)
	}

	return nil
}

func (p *Pipeline) upsertIssue(ctx context.Context, projectID uuid.UUID, fingerprint, title, culprit string, event *ParsedEvent) (uuid.UUID, error) {
	var issueID uuid.UUID

	err := p.pool.QueryRow(ctx, `
		INSERT INTO issues (project_id, fingerprint, title, culprit, level, platform, first_seen, last_seen, event_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7, 1)
		ON CONFLICT (project_id, fingerprint) DO UPDATE SET
			last_seen = EXCLUDED.last_seen,
			event_count = issues.event_count + 1,
			status = CASE
				WHEN issues.status = 'resolved' THEN 'reappeared'
				ELSE issues.status
			END,
			updated_at = NOW()
		RETURNING id
	`, projectID, fingerprint, title, culprit, event.Level, event.Platform, event.Timestamp).Scan(&issueID)

	return issueID, err
}

func (p *Pipeline) storeEvent(ctx context.Context, issueID, projectID uuid.UUID, event *ParsedEvent) error {
	userData, _ := json.Marshal(event.User)
	requestData, _ := json.Marshal(event.Request)
	breadcrumbs, _ := json.Marshal(event.Breadcrumbs)
	contexts, _ := json.Marshal(event.Contexts)
	tags, _ := json.Marshal(event.Tags)
	exception, _ := json.Marshal(event.Exception)

	ipAddress := ""
	if event.User != nil {
		ipAddress = event.User.IPAddress
	}

	var ipArg interface{}
	if ipAddress != "" {
		ipArg = ipAddress
	}

	_, err := p.pool.Exec(ctx, `
		INSERT INTO events (event_id, issue_id, project_id, timestamp, level, platform,
			ip_address, user_data, request_data, breadcrumbs, contexts, tags, exception,
			message, environment, release_tag, server_name, raw_payload)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`, event.EventID, issueID, projectID, event.Timestamp, event.Level, event.Platform,
		ipArg, userData, requestData, breadcrumbs, contexts, tags, exception,
		event.Message, event.Environment, event.Release, event.ServerName, event.RawPayload)

	return err
}

func buildTitle(event *ParsedEvent) string {
	if event.Exception != nil && len(event.Exception.Values) > 0 {
		last := event.Exception.Values[len(event.Exception.Values)-1]
		if last.Type != "" && last.Value != "" {
			title := last.Type + ": " + last.Value
			if len(title) > 1024 {
				title = title[:1024]
			}
			return title
		}
		if last.Type != "" {
			return last.Type
		}
	}
	if event.Message != "" {
		if len(event.Message) > 1024 {
			return event.Message[:1024]
		}
		return event.Message
	}
	return "Unknown Error"
}

func buildCulprit(event *ParsedEvent) string {
	if event.Exception != nil && len(event.Exception.Values) > 0 {
		last := event.Exception.Values[len(event.Exception.Values)-1]
		if last.Stacktrace != nil && len(last.Stacktrace.Frames) > 0 {
			frame := last.Stacktrace.Frames[len(last.Stacktrace.Frames)-1]
			if frame.Filename != "" {
				culprit := frame.Filename
				if frame.Function != "" {
					culprit += " in " + frame.Function
				}
				return culprit
			}
		}
	}
	if event.Transaction != "" {
		return event.Transaction
	}
	return ""
}

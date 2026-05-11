package processing

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Aggregator struct {
	pool *pgxpool.Pool
}

func NewAggregator(pool *pgxpool.Pool) *Aggregator {
	return &Aggregator{pool: pool}
}

func (a *Aggregator) UpdateAggregates(ctx context.Context, issueID uuid.UUID, event *ParsedEvent) error {
	browser := extractBrowser(event)
	osName := extractOS(event)
	device := extractDevice(event)
	url := extractURL(event)

	row := a.pool.QueryRow(ctx, "SELECT browsers, os_names, devices, urls FROM issues WHERE id = $1", issueID)

	var browsersRaw, osRaw, devicesRaw, urlsRaw []byte
	if err := row.Scan(&browsersRaw, &osRaw, &devicesRaw, &urlsRaw); err != nil {
		return err
	}

	browsers := jsonbIncrement(browsersRaw, browser)
	osNames := jsonbIncrement(osRaw, osName)
	devices := jsonbIncrement(devicesRaw, device)
	urls := jsonbIncrement(urlsRaw, url)

	_, err := a.pool.Exec(ctx,
		"UPDATE issues SET browsers = $1, os_names = $2, devices = $3, urls = $4 WHERE id = $5",
		browsers, osNames, devices, urls, issueID,
	)
	return err
}

func jsonbIncrement(raw []byte, key string) []byte {
	if key == "" {
		return raw
	}
	m := make(map[string]int)
	json.Unmarshal(raw, &m)
	m[key]++
	result, _ := json.Marshal(m)
	return result
}

func extractBrowser(event *ParsedEvent) string {
	if event.Contexts == nil {
		return ""
	}
	browser, ok := event.Contexts["browser"]
	if !ok {
		return ""
	}
	bMap, ok := browser.(map[string]interface{})
	if !ok {
		return ""
	}
	name, _ := bMap["name"].(string)
	version, _ := bMap["version"].(string)
	if name == "" {
		return ""
	}
	if version != "" {
		return name + " " + version
	}
	return name
}

func extractOS(event *ParsedEvent) string {
	if event.Contexts == nil {
		return ""
	}
	os, ok := event.Contexts["os"]
	if !ok {
		return ""
	}
	oMap, ok := os.(map[string]interface{})
	if !ok {
		return ""
	}
	name, _ := oMap["name"].(string)
	version, _ := oMap["version"].(string)
	if name == "" {
		return ""
	}
	if version != "" {
		return name + " " + version
	}
	return name
}

func extractDevice(event *ParsedEvent) string {
	if event.Contexts == nil {
		return ""
	}
	device, ok := event.Contexts["device"]
	if !ok {
		return ""
	}
	dMap, ok := device.(map[string]interface{})
	if !ok {
		return ""
	}
	family, _ := dMap["family"].(string)
	if family != "" {
		return family
	}
	model, _ := dMap["model"].(string)
	return model
}

func extractURL(event *ParsedEvent) string {
	if event.Request == nil {
		return ""
	}
	return event.Request.URL
}

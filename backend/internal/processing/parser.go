package processing

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"
)

type ParsedEvent struct {
	EventID     string                 `json:"event_id"`
	Timestamp   time.Time              `json:"timestamp"`
	Level       string                 `json:"level"`
	Platform    string                 `json:"platform"`
	Logger      string                 `json:"logger"`
	Transaction string                 `json:"transaction"`
	ServerName  string                 `json:"server_name"`
	Release     string                 `json:"release"`
	Environment string                 `json:"environment"`
	Message     string                 `json:"message"`
	Exception   *ExceptionData         `json:"exception"`
	Breadcrumbs []Breadcrumb           `json:"breadcrumbs"`
	User        *UserData              `json:"user"`
	Request     *RequestData           `json:"request"`
	Contexts    map[string]interface{} `json:"contexts"`
	Tags        map[string]string      `json:"tags"`
	Fingerprint []string               `json:"fingerprint"`
	Extra       map[string]interface{} `json:"extra"`
	RawPayload  json.RawMessage        `json:"-"`
}

type ExceptionData struct {
	Values []ExceptionValue `json:"values"`
}

type ExceptionValue struct {
	Type       string      `json:"type"`
	Value      string      `json:"value"`
	Module     string      `json:"module"`
	Stacktrace *Stacktrace `json:"stacktrace"`
	Mechanism  *Mechanism  `json:"mechanism"`
}

type Stacktrace struct {
	Frames []Frame `json:"frames"`
}

type Frame struct {
	Filename string `json:"filename"`
	Function string `json:"function"`
	Module   string `json:"module"`
	Lineno   int    `json:"lineno"`
	Colno    int    `json:"colno"`
	AbsPath  string `json:"abs_path"`
	InApp    *bool  `json:"in_app"`
}

type Mechanism struct {
	Type    string `json:"type"`
	Handled *bool  `json:"handled"`
}

type Breadcrumb struct {
	Timestamp interface{}            `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Category  string                 `json:"category"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
}

type UserData struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	IPAddress string `json:"ip_address"`
}

type RequestData struct {
	URL         string            `json:"url"`
	Method      string            `json:"method"`
	Headers     map[string]string `json:"headers"`
	QueryString string            `json:"query_string"`
	Data        interface{}       `json:"data"`
	Env         map[string]string `json:"env"`
}

func ParseEnvelope(data []byte) (*ParsedEvent, error) {
	lines := bytes.Split(data, []byte("\n"))
	if len(lines) < 3 {
		return nil, ErrInvalidEnvelope
	}

	var event *ParsedEvent
	i := 1 // skip envelope header (first line)

	for i < len(lines) {
		if len(bytes.TrimSpace(lines[i])) == 0 {
			i++
			continue
		}

		var itemHeader struct {
			Type   string `json:"type"`
			Length int    `json:"length"`
		}
		if err := json.Unmarshal(lines[i], &itemHeader); err != nil {
			i++
			continue
		}
		i++

		if i >= len(lines) {
			break
		}

		if itemHeader.Type == "event" || itemHeader.Type == "transaction" {
			var payload []byte
			if itemHeader.Length > 0 && i < len(lines) {
				payload = lines[i]
			} else {
				var remaining [][]byte
				for j := i; j < len(lines); j++ {
					remaining = append(remaining, lines[j])
				}
				payload = bytes.Join(remaining, []byte("\n"))
				payload = bytes.TrimRight(payload, "\n")
			}

			parsed, err := parseEventPayload(payload)
			if err == nil {
				event = parsed
			}
			break
		}

		i++
	}

	if event == nil {
		return nil, ErrNoEventInEnvelope
	}
	return event, nil
}

func ParseStorePayload(data []byte) (*ParsedEvent, error) {
	return parseEventPayload(data)
}

func parseEventPayload(data []byte) (*ParsedEvent, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	event := &ParsedEvent{
		RawPayload: data,
	}

	if v, ok := raw["event_id"]; ok {
		json.Unmarshal(v, &event.EventID)
	}
	event.EventID = strings.ReplaceAll(event.EventID, "-", "")

	if v, ok := raw["timestamp"]; ok {
		parseTimestamp(v, &event.Timestamp)
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	jsonString(raw, "level", &event.Level)
	jsonString(raw, "platform", &event.Platform)
	jsonString(raw, "logger", &event.Logger)
	jsonString(raw, "transaction", &event.Transaction)
	jsonString(raw, "server_name", &event.ServerName)
	jsonString(raw, "release", &event.Release)
	jsonString(raw, "environment", &event.Environment)

	if v, ok := raw["message"]; ok {
		json.Unmarshal(v, &event.Message)
	}
	if v, ok := raw["logentry"]; ok {
		var le struct {
			Formatted string `json:"formatted"`
			Message   string `json:"message"`
		}
		json.Unmarshal(v, &le)
		if le.Formatted != "" && event.Message == "" {
			event.Message = le.Formatted
		} else if le.Message != "" && event.Message == "" {
			event.Message = le.Message
		}
	}

	if v, ok := raw["exception"]; ok {
		var exc ExceptionData
		if err := json.Unmarshal(v, &exc); err == nil {
			event.Exception = &exc
		}
	}

	if v, ok := raw["breadcrumbs"]; ok {
		var bc struct {
			Values []Breadcrumb `json:"values"`
		}
		if err := json.Unmarshal(v, &bc); err == nil {
			event.Breadcrumbs = bc.Values
		} else {
			var bcList []Breadcrumb
			if err := json.Unmarshal(v, &bcList); err == nil {
				event.Breadcrumbs = bcList
			}
		}
	}

	if v, ok := raw["user"]; ok {
		var u UserData
		if err := json.Unmarshal(v, &u); err == nil {
			event.User = &u
		}
	}

	if v, ok := raw["request"]; ok {
		var req RequestData
		if err := json.Unmarshal(v, &req); err == nil {
			event.Request = &req
		}
	}

	if v, ok := raw["contexts"]; ok {
		var ctx map[string]interface{}
		if err := json.Unmarshal(v, &ctx); err == nil {
			event.Contexts = ctx
		}
	}

	if v, ok := raw["tags"]; ok {
		var tags map[string]string
		if err := json.Unmarshal(v, &tags); err == nil {
			event.Tags = tags
		} else {
			var tagList [][]string
			if err := json.Unmarshal(v, &tagList); err == nil {
				event.Tags = make(map[string]string)
				for _, pair := range tagList {
					if len(pair) == 2 {
						event.Tags[pair[0]] = pair[1]
					}
				}
			}
		}
	}

	if v, ok := raw["fingerprint"]; ok {
		json.Unmarshal(v, &event.Fingerprint)
	}

	if v, ok := raw["extra"]; ok {
		json.Unmarshal(v, &event.Extra)
	}

	if event.Level == "" {
		event.Level = "error"
	}

	return event, nil
}

func parseTimestamp(data json.RawMessage, t *time.Time) {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		if parsed, err := time.Parse(time.RFC3339Nano, s); err == nil {
			*t = parsed
			return
		}
		if parsed, err := time.Parse("2006-01-02T15:04:05", s); err == nil {
			*t = parsed
			return
		}
	}
	var f float64
	if err := json.Unmarshal(data, &f); err == nil {
		sec := int64(f)
		nsec := int64((f - float64(sec)) * 1e9)
		*t = time.Unix(sec, nsec)
	}
}

func jsonString(raw map[string]json.RawMessage, key string, dest *string) {
	if v, ok := raw[key]; ok {
		json.Unmarshal(v, dest)
	}
}

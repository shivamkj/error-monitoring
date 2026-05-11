package processing

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

func GenerateFingerprint(event *ParsedEvent, normalized NormalizedData) string {
	if len(event.Fingerprint) > 0 {
		hasDefault := false
		for _, f := range event.Fingerprint {
			if f == "{{ default }}" {
				hasDefault = true
				break
			}
		}
		if !hasDefault {
			input := strings.Join(event.Fingerprint, "||")
			return sha256Hex(input)
		}
	}

	if len(normalized.Frames) > 0 {
		var parts []string
		for _, f := range normalized.Frames {
			parts = append(parts, fmt.Sprintf("%s:%s:%d", f.Filename, f.Function, f.Lineno))
		}
		input := strings.Join(parts, "\n")
		return sha256Hex(input)
	}

	if event.Exception != nil && len(event.Exception.Values) > 0 {
		last := event.Exception.Values[len(event.Exception.Values)-1]
		if last.Type != "" {
			input := fmt.Sprintf("%s||%s", last.Type, normalized.ExceptionValue)
			return sha256Hex(input)
		}
	}

	if normalized.Message != "" {
		return sha256Hex(normalized.Message)
	}

	return sha256Hex("unknown-error")
}

func sha256Hex(input string) string {
	h := sha256.Sum256([]byte(input))
	return hex.EncodeToString(h[:])
}

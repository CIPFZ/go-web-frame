package audit

import (
	"bytes"
	"encoding/json"
	"strings"
)

// MaxBodySize bounds each queued request/response body, including legacy records.
const MaxBodySize = 8 * 1024

// SanitizeBody only retains complete JSON. Partial, oversized and non-JSON bodies
// are omitted rather than risking leakage from a truncated sensitive field.
func SanitizeBody(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	if len(body) > MaxBodySize {
		return "[body omitted: too large]"
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if !json.Valid(body) || decoder.Decode(&value) != nil {
		return "[body omitted: not complete JSON]"
	}
	redact(value)
	encoded, err := json.Marshal(value)
	if err != nil || len(encoded) > MaxBodySize {
		return "[body omitted: too large]"
	}
	return string(encoded)
}

func redact(value any) {
	switch v := value.(type) {
	case map[string]any:
		for key, child := range v {
			k := strings.ToLower(strings.NewReplacer("_", "", "-", "").Replace(key))
			sensitive := false
			for _, word := range []string{"password", "passwd", "token", "secret", "authorization", "cookie", "credential", "privatekey", "accesskey"} {
				if strings.Contains(k, word) {
					sensitive = true
					break
				}
			}
			if sensitive {
				v[key] = "[REDACTED]"
			} else {
				redact(child)
			}
		}
	case []any:
		for _, child := range v {
			redact(child)
		}
	}
}

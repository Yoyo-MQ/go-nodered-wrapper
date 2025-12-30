package util

import "encoding/json"

// MarshalPayload safely marshals a payload value to JSON, ensuring it never becomes "null".
// If the payload is nil or empty, it returns an empty object "{}" instead of "null".
// This is used when sending data to Node-RED inject nodes via __user_inject_props__.
func MarshalPayload(payload interface{}) (string, error) {
	if payload == nil {
		return "{}", nil
	}

	// Handle empty maps and slices
	switch v := payload.(type) {
	case map[string]interface{}:
		if len(v) == 0 {
			return "{}", nil
		}
	case []interface{}:
		if len(v) == 0 {
			return "[]", nil
		}
	}

	// Marshal the payload
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	// Safety check: never return "null" - use empty object instead
	payloadValue := string(payloadJSON)
	if payloadValue == "null" {
		return "{}", nil
	}

	return payloadValue, nil
}

// GetString safely retrieves a string from a map[string]interface{}
func GetString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// GetFloat64 safely retrieves a float64 from a map[string]interface{}, handling int conversion
func GetFloat64(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	if v, ok := m[key].(int); ok {
		return float64(v)
	}
	return 0
}

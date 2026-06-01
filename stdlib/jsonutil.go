package stdlib

import (
	"encoding/json"
	"fmt"
)

// JSON utilities for Kylix

// JsonEncode converts a value to JSON string
func JsonEncode(value interface{}) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("JsonEncode: %w", err)
	}
	return string(data), nil
}

// JsonEncodePretty converts a value to indented JSON string
func JsonEncodePretty(value interface{}) (string, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JsonEncodePretty: %w", err)
	}
	return string(data), nil
}

// JsonDecode parses a JSON string into a generic value
func JsonDecode(jsonStr string) (interface{}, error) {
	var result interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, fmt.Errorf("JsonDecode: %w", err)
	}
	return result, nil
}

// JsonDecodeMap parses a JSON string into a map
func JsonDecodeMap(jsonStr string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, fmt.Errorf("JsonDecodeMap: %w", err)
	}
	return result, nil
}

// JsonDecodeArray parses a JSON string into an array
func JsonDecodeArray(jsonStr string) ([]interface{}, error) {
	var result []interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, fmt.Errorf("JsonDecodeArray: %w", err)
	}
	return result, nil
}

// JsonGetString gets a string value from a JSON map by key
func JsonGetString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// JsonGetInt gets an integer value from a JSON map by key
func JsonGetInt(m map[string]interface{}, key string) int64 {
	if v, ok := m[key]; ok {
		switch n := v.(type) {
		case float64:
			return int64(n)
		case int64:
			return n
		case int:
			return int64(n)
		}
	}
	return 0
}

// JsonGetFloat gets a float value from a JSON map by key
func JsonGetFloat(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return 0
}

// JsonGetBool gets a boolean value from a JSON map by key
func JsonGetBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// JsonGetMap gets a nested map from a JSON map by key
func JsonGetMap(m map[string]interface{}, key string) map[string]interface{} {
	if v, ok := m[key]; ok {
		if nested, ok := v.(map[string]interface{}); ok {
			return nested
		}
	}
	return nil
}

// JsonGetArray gets an array from a JSON map by key
func JsonGetArray(m map[string]interface{}, key string) []interface{} {
	if v, ok := m[key]; ok {
		if arr, ok := v.([]interface{}); ok {
			return arr
		}
	}
	return nil
}

// JsonHasKey checks if a JSON map contains a key
func JsonHasKey(m map[string]interface{}, key string) bool {
	_, ok := m[key]
	return ok
}

// JsonIsValid checks if a string is valid JSON
func JsonIsValid(jsonStr string) bool {
	return json.Valid([]byte(jsonStr))
}

// JsonReadFile reads and parses a JSON file
func JsonReadFile(path string) (interface{}, error) {
	content, err := ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("JsonReadFile: %w", err)
	}
	return JsonDecode(content)
}

// JsonWriteFile writes a value as JSON to a file
func JsonWriteFile(path string, value interface{}) error {
	content, err := JsonEncodePretty(value)
	if err != nil {
		return fmt.Errorf("JsonWriteFile: %w", err)
	}
	return WriteFile(path, content)
}

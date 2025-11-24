package main

import (
	"encoding/json"
	"strings"
)

func containsAny(target string) bool {
	for _, member := range FILTER_LOG {
		if strings.Contains(target, member) {
			return true
		}
	}
	return false
}

func parseKeyValueLog(s string) (map[string]string, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "{")
	s = strings.TrimSuffix(s, "}")

	parts := strings.Split(s, ",")
	result := map[string]string{}

	for _, p := range parts {
		kv := strings.SplitN(strings.TrimSpace(p), "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := kv[0]
		value := kv[1]
		result[key] = value
	}

	return result, nil
}

func ConvertToApiLog(raw string) (*LogEntry, error) {
	kv, err := parseKeyValueLog(raw)
	if err != nil {
		return nil, err
	}

	// convert to JSON
	jsonBytes, err := json.Marshal(kv)
	if err != nil {
		return nil, err
	}

	var log LogEntry
	if err := json.Unmarshal(jsonBytes, &log); err != nil {
		return nil, err
	}

	return &log, nil
}

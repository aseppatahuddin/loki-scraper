package lokiparser

import (
	"encoding/json"

	"github.com/prakasa1904/loki-scraper/internal/model"
)

// convert map to structed model
func Parse(raw string) (*model.LogEntry, error) {
	kv, err := parseKeyValueLog(raw)
	if err != nil {
		return nil, err
	}

	// convert to JSON
	jsonBytes, err := json.Marshal(kv)
	if err != nil {
		return nil, err
	}

	var log model.LogEntry
	if err := json.Unmarshal(jsonBytes, &log); err != nil {
		return nil, err
	}

	return &log, nil
}

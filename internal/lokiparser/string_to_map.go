package lokiparser

import (
	"strings"
)

// check if log contains member of array
func IfLogContains(log string, array []string) bool {
	for _, member := range array {
		if strings.Contains(log, member) {
			return true
		}
	}
	return false
}

// parse string {} to key value map
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

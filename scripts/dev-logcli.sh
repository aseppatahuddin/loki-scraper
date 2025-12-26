#!/bin/bash

export LOKI_ADDR=http://127.0.0.1:3100
export LOKI_BINARY="./scripts/logcli"
export LOKI_QUERY="{container=\"splp-gw\"} |= \"LogCounterMetric\""
export LOKI_START_DATE="2025-11-15T22:00:00Z"
export LOKI_END_DATE="2025-11-15T22:01:00Z"
export LOKI_LIMIT="100"
export LOKI_DEBUG="true"

# # Clickhouse config
export CLICKHOUSE_HOST="localhost:8123"
export CLICKHOUSE_DATABASE="loki_log"
export CLICKHOUSE_TABLE="log_entry_logcli"
export CLICKHOUSE_USER="default"
export CLICKHOUSE_PASSWORD="MyStrongPassword123!"


go run cmd/with-logcli/*.go

# ./scripts/logcli query --from="2025-11-15T22:00:00Z" --to="2025-11-15T22:01:00Z" --batch 100 --output=jsonl --limit 0 "{container=\"splp-gw\"} |= \"LogCounterMetric\""
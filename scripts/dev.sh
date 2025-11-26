#!/bin/bash

# http://127.0.0.1:3100/loki/api/v1/query_range
export LOKI_URL="http://127.0.0.1:3100/loki/api/v1/query_range"
export LOKI_QUERY="{container=\"splp-gw\"} |= \"LogCounterMetric\""
export LOKI_START_DATE="2025-11-15T22:00:00Z"
export LOKI_END_DATE="2025-11-15T22:01:00Z"
# export LOKI_LIMIT="10"

# Clickhouse config
export CLICKHOUSE_HOST="localhost:8123"
export CLICKHOUSE_DATABASE="loki_log"
export CLICKHOUSE_USER="default"
export CLICKHOUSE_PASSWORD="MyStrongPassword123!"

go run cmd/*.go
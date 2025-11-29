#!/bin/bash

export LOKI_URL="http://127.0.0.1:3100/loki/api/v1/query_range"
export LOKI_QUERY="{container=\"splp-gw\"} |= \"LogCounterMetric\""
export LOKI_LIMIT="100"
export LOKI_LAST_MONTH="-2"

# Clickhouse config
export CLICKHOUSE_HOST="localhost:8123"
export CLICKHOUSE_DATABASE="loki_log"
export CLICKHOUSE_TABLE="log_entry_dev"
export CLICKHOUSE_USER="default"
export CLICKHOUSE_PASSWORD="MyStrongPassword123!"

go run cmd/n-to-now/*.go
#!/bin/bash

export LOKI_URL="http://127.0.0.1:3100/loki/api/v1/query_range"
export LOKI_QUERY="{namespace=\"splp-2025-apps\"} |= \"LogCounterMetric\""
export LOKI_LIMIT="100"
export LOKI_LAST_MONTH="-2"

# Clickhouse config
export CLICKHOUSE_HOST="localhost:8123"
export CLICKHOUSE_DATABASE="loki_log"
export CLICKHOUSE_TABLE="log_entry_uat"
export CLICKHOUSE_USER="appuser"
export CLICKHOUSE_PASSWORD="bintang123#"

go run cmd/n-to-now/*.go

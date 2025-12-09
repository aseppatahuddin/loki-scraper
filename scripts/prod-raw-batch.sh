#!/bin/bash

export LOKI_URL="http://10.31.67.116:3100/loki/api/v1/query_range"
export LOKI_QUERY="{container=\"splp-gw\"} |= \"LogCounterMetric\""
export LOKI_START_DATE="2025-11-30T08:00:00.000Z"
export LOKI_END_DATE="2025-11-30T08:05:00.000Z"
export LOKI_LIMIT="100"
export LOKI_DEBUG="true"

# Clickhouse config
export CLICKHOUSE_HOST="localhost:8123"
export CLICKHOUSE_DATABASE="loki_log"
export CLICKHOUSE_TABLE="log_entry_dev"
export CLICKHOUSE_USER="appuser"
export CLICKHOUSE_PASSWORD="bintang123#"

./loki-scraper
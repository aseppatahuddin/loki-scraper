#!/bin/bash

# http://127.0.0.1:3100/loki/api/v1/query_range
export LOKI_URL="http://127.0.0.1:3100/loki/api/v1/query_range"
export LOKI_QUERY="{container=\"splp-gw\"}"
export LOKI_START_DATE="2025-11-01 00:00:00"
export LOKI_END_DATE="2025-11-22 00:00:00"
export LOKI_LIMIT="100"

# Clickhouse config
export CLICKHOUSE_HOST="localhost:8123"
export CLICKHOUSE_DATABASE="loki_log"
export CLICKHOUSE_USER="default"
export CLICKHOUSE_PASSWORD="default"

./loki-scraper
#!/bin/bash

# Google IAM Token Location
export GOOGLE_APPLICATION_CREDENTIALS="gdrive-credentials/gdrive-secret.json"

# Clickhouse config
export CLICKHOUSE_HOST="localhost:8123"
export CLICKHOUSE_DATABASE="loki_log"
export CLICKHOUSE_TABLE="log_entry_dev"
export CLICKHOUSE_USER="appuser"
export CLICKHOUSE_PASSWORD="bintang123#"

go run cmd/read-files/*.go
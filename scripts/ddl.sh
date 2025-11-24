#!/bin/bash

CREATE TABLE IF NOT EXISTS loki_log.log_entry (
    timestamp String,
    log       String,
    stream    String,
    time      DateTime64(9, 'UTC')
)
ENGINE = MergeTree
PARTITION BY toDate(time)
ORDER BY (time, stream);

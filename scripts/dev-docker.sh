#!/bin/bash

docker run -e LOKI_URL="https://pawon-beta.terpusat.com/api/v1/query_range" -e LOKI_START_DATE="2025-01-01 00:00:00" -e LOKI_END_DATE="2025-01-30 00:00:00" -e LOKI_LIMIT="100" prakasa1904/loki-scraper:7452d01
WITH 
    window_size AS (
        SELECT greatest(
            toUnixTimestamp($__toTime) - toUnixTimestamp($__fromTime),
            60  -- Minimum 60 seconds
        ) AS window_seconds
    )
SELECT
    time_bucket AS time,
    count_in_window AS value
FROM (
    SELECT
        time_bucket,
        countIf(
            log_time > time_bucket - (SELECT window_seconds FROM window_size)
            AND log_time <= time_bucket
        ) AS count_in_window
    FROM (
        SELECT DISTINCT toStartOfSecond(time) AS time_bucket
        FROM loki_log.log_entry_dev
        WHERE
            container = 'splp-gw'
            AND time >= $__fromTime
            AND time < $__toTime
    ) AS buckets
    CROSS JOIN (
        SELECT toStartOfSecond(time) AS log_time
        FROM loki_log.log_entry_dev
        WHERE
            container = 'splp-gw'
            AND time >= $__fromTime - (SELECT window_seconds FROM window_size)
            AND time < $__toTime
    ) AS logs
    GROUP BY time_bucket
)
ORDER BY time_bucket;
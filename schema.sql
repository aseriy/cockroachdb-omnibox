CREATE TABLE IF NOT EXISTS datapoints (
    device STRING NOT NULL,
    at TIMESTAMP NOT NULL,
    param0 INT8 NULL,
    param1 INT8 NULL,
    param2 FLOAT8 NULL,
    param3 FLOAT8 NULL,
    param4 STRING NULL,
    param5 JSONB NULL,
    CONSTRAINT datapoints_pkey PRIMARY KEY (device ASC, at ASC) USING HASH WITH (bucket_count=16),
    INDEX datapoints_at_idx (at ASC)
);


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
)  WITH (
    ttl = 'on', ttl_expiration_expression = e'((at::TIMESTAMPTZ) + INTERVAL ''1 week'')', 
    ttl_job_cron = '@daily', schema_locked = true
);

CREATE INDEX ON omnibox.public.datapoints (device) STORING (param0);
CREATE INDEX ON omnibox.public.datapoints (device) STORING (param1);
CREATE INDEX ON omnibox.public.datapoints (device) STORING (param2);
CREATE INDEX ON omnibox.public.datapoints (device) STORING (param3);
CREATE INDEX ON omnibox.public.datapoints (device) STORING (param4);

CREATE TABLE IF NOT EXISTS message_board (
    device_from STRING NOT NULL,
    device_to STRING NOT NULL,
    received_at TIMESTAMP NOT NULL DEFAULT now(),
    message STRING NOT NULL,
    read_at TIMESTAMP,
    CONSTRAINT message_board_pkey PRIMARY KEY (device_from ASC, device_to ASC, received_at ASC) USING HASH WITH (bucket_count=16)
) WITH (
    ttl_expiration_expression = '((read_at::TIMESTAMPTZ) + INTERVAL ''24 hours'')',
    ttl_job_cron = '@hourly', schema_locked = true
);

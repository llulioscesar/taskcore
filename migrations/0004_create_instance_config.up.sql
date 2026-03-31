CREATE TABLE instance_config (
    key        TEXT        PRIMARY KEY,
    value      TEXT        NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO instance_config (key, value) VALUES ('initialized', 'false');

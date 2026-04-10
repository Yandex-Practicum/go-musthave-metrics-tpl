CREATE TABLE IF NOT EXISTS metrics (
    id TEXT NOT NULL,
    type TEXT NOT NULL,
    delta BIGINT,
    value DOUBLE PRECISION,
    PRIMARY KEY (id, type)
);
CREATE TABLE IF NOT EXISTS user_tokens(
    userid BIGINT NOT NULL,
    token BYTEA NOT NULL UNIQUE,
    expiry TIMESTAMPTZ NOT NULL
);
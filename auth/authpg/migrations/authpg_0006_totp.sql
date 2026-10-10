-- +goose Up
CREATE TABLE auth_totp (
    owner_id UUID PRIMARY KEY,
    secret BYTEA NOT NULL,
    last_step BIGINT NOT NULL DEFAULT 0,
    confirmed_at TIMESTAMPTZ,
    recovery_codes BYTEA[] NOT NULL DEFAULT '{}'
);

-- +goose Down
DROP TABLE auth_totp;

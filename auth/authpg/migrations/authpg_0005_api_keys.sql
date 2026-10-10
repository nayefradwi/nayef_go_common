-- +goose Up
CREATE TABLE auth_api_keys (
    id UUID PRIMARY KEY,
    owner_id UUID NOT NULL,
    name TEXT NOT NULL,
    hash BYTEA NOT NULL UNIQUE,
    claims JSONB,
    expires_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_auth_api_keys_owner_id ON auth_api_keys(owner_id);
CREATE INDEX idx_auth_api_keys_expires_at ON auth_api_keys(expires_at);

-- +goose Down
DROP TABLE auth_api_keys;

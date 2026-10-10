-- +goose Up
CREATE TABLE auth_sessions (
    refresh_hash BYTEA PRIMARY KEY,
    family_id UUID NOT NULL,
    owner_id UUID NOT NULL,
    access_hash BYTEA UNIQUE,
    access_expires_at TIMESTAMPTZ,
    claims JSONB,
    expires_at TIMESTAMPTZ NOT NULL,
    rotated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_auth_sessions_family_id ON auth_sessions(family_id);
CREATE INDEX idx_auth_sessions_owner_id ON auth_sessions(owner_id);
CREATE INDEX idx_auth_sessions_expires_at ON auth_sessions(expires_at);

-- +goose Down
DROP TABLE auth_sessions;

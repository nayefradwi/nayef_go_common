-- +goose Up
CREATE TABLE auth_otps (
    key TEXT PRIMARY KEY,
    hash BYTEA NOT NULL,
    attempts INT NOT NULL,
    sends INT NOT NULL,
    sent_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_auth_otps_expires_at ON auth_otps(expires_at);

-- +goose Down
DROP TABLE auth_otps;

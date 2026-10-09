-- +goose Up
CREATE TABLE auth_attempts (
    key TEXT PRIMARY KEY,
    count INT NOT NULL,
    reset_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_auth_attempts_reset_at ON auth_attempts(reset_at);

-- +goose Down
DROP TABLE auth_attempts;

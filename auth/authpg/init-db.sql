CREATE TABLE tokens (
    id UUID PRIMARY KEY,
    value TEXT NOT NULL,
    owner_id UUID NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    issued_at TIMESTAMPTZ NOT NULL,
    claims JSONB,
    type INT NOT NULL
);
CREATE INDEX idx_tokens_owner_id ON tokens(owner_id);
CREATE INDEX idx_tokens_reference_type ON tokens(id, type);

package auth

import (
	"time"

	"github.com/google/uuid"
)

const (
	ownerClaimKey     = "owner"
	expiryClaimKey    = "exp"
	issuerClaimKey    = "iss"
	issuedAtClaimKey  = "iat"
	tokenTypeClaimKey = "token_type"
	audienceClaimKey  = "aud"
)

type TokenType string

const (
	AccessTokenType  TokenType = "access"
	RefreshTokenType TokenType = "refresh"
)

type Token struct {
	Value     string
	OwnerId   uuid.UUID
	ExpiresAt time.Time
	IssuedAt  time.Time
	Claims    map[string]any
	Type      TokenType
}

func (t Token) IsExpired() bool {
	return time.Now().UTC().After(t.ExpiresAt)
}

func (t Token) IsOwner(owner uuid.UUID) bool {
	return t.OwnerId == owner
}

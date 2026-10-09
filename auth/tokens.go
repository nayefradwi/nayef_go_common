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

const (
	AccessTokenType  = 1
	RefreshTokenType = 2
)

type Token struct {
	Id        uuid.UUID
	Value     string
	OwnerId   uuid.UUID
	ExpiresAt time.Time
	IssuedAt  time.Time
	Claims    map[string]any
	Type      int
}

func (t Token) IsExpired() bool {
	return time.Now().UTC().After(t.ExpiresAt)
}

func (t Token) IsOwner(owner uuid.UUID) bool {
	return t.OwnerId == owner
}

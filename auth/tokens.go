package auth

import (
	"context"
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

type TokenKey struct{}

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

func (t Token) WithToken(ctx context.Context) context.Context {
	return context.WithValue(ctx, TokenKey{}, t)
}

func GetToken(ctx context.Context) Token {
	t, ok := ctx.Value(TokenKey{}).(Token)
	if !ok {
		return Token{
			Claims: make(map[string]any),
		}
	}

	return t
}

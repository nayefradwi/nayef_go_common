package auth

import (
	"context"
	"time"
)

const (
	ownerClaimKey    = "owner"
	expiryClaimKey   = "exp"
	issuerClaimKey   = "iss"
	issuedAtClaimKey = "iat"
)

const (
	AccessTokenType  = 1
	RefreshTokenType = 2
)

type TokenKey struct{}

type Token struct {
	Id        string
	Value     string
	OwnerId   string
	ExpiresAt time.Time
	IssuedAt  time.Time
	Claims    map[string]interface{}
	Type      int
}

func (t Token) IsExpired() bool {
	return time.Now().UTC().After(t.ExpiresAt)
}

func (t Token) IsOwner(owner string) bool {
	return t.OwnerId == owner
}

func (t Token) WithToken(ctx context.Context) context.Context {
	return context.WithValue(ctx, TokenKey{}, t)
}

func GetToken(ctx context.Context) Token {
	t, ok := ctx.Value(TokenKey{}).(Token)
	if !ok {
		return Token{
			Claims: make(map[string]interface{}),
		}
	}

	return t
}

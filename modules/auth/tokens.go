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

type TokenKey struct{}

type Token struct {
	Id        string
	Value     string
	OwnerId   string
	ExpiresAt time.Time
	issuedAt  time.Time
	Claims    map[string]interface{}
	Type      int
}

type ITokenProvider interface {
	GetClaims(token string) (Token, error)
	SignClaims(owner string, claims map[string]interface{}) (string, error)
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

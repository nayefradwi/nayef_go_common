package auth

import (
	"time"
)

const (
	ownerClaimKey    = "owner"
	expiryClaimKey   = "exp"
	issuerClaimKey   = "iss"
	issuedAtClaimKey = "iat"
)

type Token[T string | int] struct {
	Id        T
	Value     string
	OwnerId   T
	ExpiresAt time.Time
	issuedAt  time.Time
	Claims    map[string]interface{}
	Type      int
}

type ITokenProvider[T string | int] interface {
	GetClaims(token string) (Token[T], error)
	SignClaims(owner T, claims map[string]interface{}) (string, error)
}

func (t Token[T]) IsExpired() bool {
	return time.Now().UTC().After(t.ExpiresAt)
}

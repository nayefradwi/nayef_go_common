package auth

import (
	"time"
)

type Token[T string | int] struct {
	Value     string
	OwnerId   T
	ExpiresAt time.Time
	Claims    map[string]Claim[T]
}

type ITokenProvider[T string | int] interface {
	GetClaims(token string) (Token[T], error)
	SignClaims(owner T, claims Claim[T]) (string, error)
}

func (t Token[T]) IsExpired() bool {
	return time.Now().UTC().After(t.ExpiresAt)
}

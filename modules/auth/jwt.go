package auth

import (
	"time"

	"github.com/nayefradwi/nayef_go_common/core"
)

type JwtTokenProviderConfig struct {
	SecretKey string
	ExpiresIn time.Duration
}

var defaultSecretKey = "SuperSecretKeyShouldBeOverriden"
var defaultExpiresIn = time.Hour * 24

func ReplaceDefaultJwtSecretKey(secretKey string) {
	DefaultJwtTokenProviderConfig.SecretKey = secretKey
}

func ReplaceDefaultJwtExpiresIn(expiresIn time.Duration) {
	DefaultJwtTokenProviderConfig.ExpiresIn = expiresIn
}

var DefaultJwtTokenProviderConfig = NewJwtTokenProviderConfig(defaultSecretKey, defaultExpiresIn)

func NewJwtTokenProviderConfig(secretKey string, expiresIn time.Duration) JwtTokenProviderConfig {
	return JwtTokenProviderConfig{
		SecretKey: secretKey,
		ExpiresIn: expiresIn,
	}
}

type JwtTokenProvider[T string | int] struct {
	Config JwtTokenProviderConfig
}

func NewJwtTokenProvider[T string | int](config JwtTokenProviderConfig) JwtTokenProvider[T] {
	return JwtTokenProvider[T]{
		Config: config,
	}
}

func NewDefaultJwtTokenProvider[T string | int]() JwtTokenProvider[T] {
	return NewJwtTokenProvider[T](DefaultJwtTokenProviderConfig)
}

func (t JwtTokenProvider[T]) GetClaims(token string) (Token[T], *core.ResultError) {
	// Implement this
	return Token[T]{}, nil
}

func (t JwtTokenProvider[T]) SignClaims(owner T, claims Claim[T]) (string, *core.ResultError) {
	// Implement this
	return "", nil
}

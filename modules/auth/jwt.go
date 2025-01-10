package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nayefradwi/nayef_go_common/core"
)

type JwtTokenProviderConfig struct {
	SecretKey     string
	ExpiresIn     time.Duration
	SigningMethod jwt.SigningMethod
}

var defaultSecretKey = "SuperSecretKeyShouldBeOverriden"
var defaultExpiresIn = time.Hour * 24

func ReplaceDefaultJwtSecretKey(secretKey string) JwtTokenProviderConfig {
	DefaultJwtTokenProviderConfig.SecretKey = secretKey
	return DefaultJwtTokenProviderConfig
}

func ReplaceDefaultJwtExpiresIn(expiresIn time.Duration) JwtTokenProviderConfig {
	DefaultJwtTokenProviderConfig.ExpiresIn = expiresIn
	return DefaultJwtTokenProviderConfig
}

func ReplaceDefaultJwtSigningMethod(signingMethod jwt.SigningMethod) JwtTokenProviderConfig {
	DefaultJwtTokenProviderConfig.SigningMethod = signingMethod
	return DefaultJwtTokenProviderConfig
}

var DefaultJwtTokenProviderConfig = NewJwtTokenProviderConfig(defaultSecretKey, defaultExpiresIn)

func NewJwtTokenProviderConfig(secretKey string, expiresIn time.Duration) JwtTokenProviderConfig {
	return JwtTokenProviderConfig{
		SecretKey:     secretKey,
		ExpiresIn:     expiresIn,
		SigningMethod: jwt.SigningMethodHS256,
	}
}

func (c JwtTokenProviderConfig) SetSecretKey(secretKey string) JwtTokenProviderConfig {
	c.SecretKey = secretKey
	return c
}

func (c JwtTokenProviderConfig) SetExpiresIn(expiresIn time.Duration) JwtTokenProviderConfig {
	c.ExpiresIn = expiresIn
	return c
}

func (c JwtTokenProviderConfig) SetSigningMethod(signingMethod jwt.SigningMethod) JwtTokenProviderConfig {
	c.SigningMethod = signingMethod
	return c
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

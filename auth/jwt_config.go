package auth

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	. "github.com/nayefradwi/nayef_go_common/errors"
)

type JwtTokenProviderConfig struct {
	SecretKey     string
	ExpiresIn     time.Duration
	signingMethod jwt.SigningMethod
	privateKey    any
	publicKey     any
	Issuer        string
	TokenType     TokenType
	Audience      string
	parser        func(token *jwt.Token) (any, error)
	signer        func(token *jwt.Token) (string, error)
}

// HS256 needs at least 256 bits of key to resist offline brute force.
const minSecretKeyBytes = 32

func NewJwtTokenProviderConfig(secretKey string, expiresIn time.Duration, tokenType TokenType) (JwtTokenProviderConfig, error) {
	if len(secretKey) < minSecretKeyBytes {
		return JwtTokenProviderConfig{}, BadRequestError("secret key must be at least 32 bytes")
	}
	if tokenType != AccessTokenType && tokenType != RefreshTokenType {
		return JwtTokenProviderConfig{}, BadRequestError("token type must be access or refresh")
	}
	return JwtTokenProviderConfig{
		SecretKey:     secretKey,
		ExpiresIn:     expiresIn,
		signingMethod: jwt.SigningMethodHS256,
		Issuer:        "AuthModule",
		TokenType:     tokenType,
		parser:        defaultHMACParser(secretKey),
		signer:        defaultHMACSigner(secretKey),
	}, nil
}

func (c JwtTokenProviderConfig) SetSecretKey(secretKey string) (JwtTokenProviderConfig, error) {
	if len(secretKey) < minSecretKeyBytes {
		return c, BadRequestError("secret key must be at least 32 bytes")
	}
	c.SecretKey = secretKey
	c.parser = defaultHMACParser(secretKey)
	c.signer = defaultHMACSigner(secretKey)
	return c, nil
}

func (c JwtTokenProviderConfig) SetExpiresIn(expiresIn time.Duration) JwtTokenProviderConfig {
	c.ExpiresIn = expiresIn
	return c
}

func (c JwtTokenProviderConfig) SetIssuer(issuer string) JwtTokenProviderConfig {
	c.Issuer = issuer
	return c
}

func (c JwtTokenProviderConfig) SetAudience(audience string) JwtTokenProviderConfig {
	c.Audience = audience
	return c
}

func (c JwtTokenProviderConfig) parserOptions() []jwt.ParserOption {
	opts := []jwt.ParserOption{
		jwt.WithIssuer(c.Issuer),
		jwt.WithExpirationRequired(),
		jwt.WithValidMethods([]string{c.signingMethod.Alg()}),
	}
	if c.Audience != "" {
		opts = append(opts, jwt.WithAudience(c.Audience))
	}
	return opts
}

// === HMAC (HS256, HS384, HS512) ===
func (c JwtTokenProviderConfig) SetHS256SigningMethod() JwtTokenProviderConfig {
	c.signingMethod = jwt.SigningMethodHS256
	c.publicKey = nil
	c.privateKey = nil
	c.parser = defaultHMACParser(c.SecretKey)
	c.signer = defaultHMACSigner(c.SecretKey)
	return c
}

func (c JwtTokenProviderConfig) SetHS384SigningMethod() JwtTokenProviderConfig {
	c.signingMethod = jwt.SigningMethodHS384
	c.publicKey = nil
	c.privateKey = nil
	c.parser = defaultHMACParser(c.SecretKey)
	c.signer = defaultHMACSigner(c.SecretKey)
	return c
}

func (c JwtTokenProviderConfig) SetHS512SigningMethod() JwtTokenProviderConfig {
	c.signingMethod = jwt.SigningMethodHS512
	c.publicKey = nil
	c.privateKey = nil
	c.parser = defaultHMACParser(c.SecretKey)
	c.signer = defaultHMACSigner(c.SecretKey)
	return c
}

// === RSA (RS256, RS384, RS512) ===
func (c JwtTokenProviderConfig) SetRSASigningMethod(signingMethod jwt.SigningMethod, privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) (JwtTokenProviderConfig, error) {
	alg := signingMethod.Alg()
	if alg != "RS256" && alg != "RS384" && alg != "RS512" {
		return c, BadRequestError(fmt.Sprintf("invalid RSA signing method: %s", alg))
	}
	if privateKey == nil {
		return c, BadRequestError("RSA private key must not be nil")
	}
	if publicKey == nil {
		return c, BadRequestError("RSA public key must not be nil")
	}

	c.signingMethod = signingMethod
	c.privateKey = privateKey
	c.publicKey = publicKey
	c.parser = defaultRSAParser(publicKey)
	c.signer = defaultRSASigner(privateKey)
	return c, nil
}

// === ECDSA (ES256, ES384, ES512) ===
func (c JwtTokenProviderConfig) SetECDSASigningMethod(
	signingMethod jwt.SigningMethod,
	privateKey *ecdsa.PrivateKey,
	publicKey *ecdsa.PublicKey,
) (JwtTokenProviderConfig, error) {
	alg := signingMethod.Alg()
	if alg != "ES256" && alg != "ES384" && alg != "ES512" {
		return c, BadRequestError(fmt.Sprintf("invalid ECDSA signing method: %s", alg))
	}
	if privateKey == nil {
		return c, BadRequestError("ECDSA private key must not be nil")
	}
	if publicKey == nil {
		return c, BadRequestError("ECDSA public key must not be nil")
	}

	c.signingMethod = signingMethod
	c.privateKey = privateKey
	c.publicKey = publicKey
	c.parser = defaultECDSAParser(publicKey)
	c.signer = defaultECDSASigner(privateKey)
	return c, nil
}

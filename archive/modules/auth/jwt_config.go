package auth

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type JwtTokenProviderConfig struct {
	SecretKey     string
	ExpiresIn     time.Duration
	signingMethod jwt.SigningMethod
	privateKey    interface{}
	publicKey     interface{}
	Issuer        string
	parser        func(token *jwt.Token) (interface{}, error)
	signer        func(token *jwt.Token) (string, error)
}

var defaultSecretKey = "SuperSecretKeyShouldBeOverriden"
var defaultExpiresIn = time.Hour * 24

func ReplaceDefaultJwtSecretKey(secretKey string) JwtTokenProviderConfig {
	return DefaultJwtTokenProviderConfig.SetSecretKey(secretKey)
}

func ReplaceDefaultJwtExpiresIn(expiresIn time.Duration) JwtTokenProviderConfig {
	return DefaultJwtTokenProviderConfig.SetExpiresIn(expiresIn)
}

func ReplaceDefaultJwtIssuer(issuer string) JwtTokenProviderConfig {
	return DefaultJwtTokenProviderConfig.SetIssuer(issuer)
}

var DefaultJwtTokenProviderConfig = NewJwtTokenProviderConfig(defaultSecretKey, defaultExpiresIn)

func NewJwtTokenProviderConfig(secretKey string, expiresIn time.Duration) JwtTokenProviderConfig {
	return JwtTokenProviderConfig{
		SecretKey:     secretKey,
		ExpiresIn:     expiresIn,
		signingMethod: jwt.SigningMethodHS256,
		Issuer:        "AuthModule",
		parser:        defaultHMACParser(secretKey),
		signer:        defaultHMACSigner(secretKey),
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

func (c JwtTokenProviderConfig) SetIssuer(issuer string) JwtTokenProviderConfig {
	c.Issuer = issuer
	return c
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
func (c JwtTokenProviderConfig) SetRSASigningMethod(signingMethod jwt.SigningMethod, privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) JwtTokenProviderConfig {

	if signingMethod.Alg() != "RS256" && signingMethod.Alg() != "RS384" && signingMethod.Alg() != "RS512" {
		zap.L().Fatal("Invalid RSA signing method", zap.String("method", signingMethod.Alg()))
	}

	if privateKey == nil {
		zap.L().Fatal("Invalid RSA private key")
	}

	if publicKey == nil {
		zap.L().Fatal("Invalid RSA public key")
	}

	c.signingMethod = signingMethod
	c.privateKey = privateKey
	c.publicKey = publicKey
	c.parser = defaultRSAParser(publicKey)
	c.signer = defaultRSASigner(privateKey)
	return c
}

// === ECDSA (ES256, ES384, ES512) ===
func (c JwtTokenProviderConfig) SetECDSASigningMethod(
	signingMethod jwt.SigningMethod,
	privateKey *ecdsa.PrivateKey,
	publicKey *ecdsa.PublicKey,
) JwtTokenProviderConfig {

	if signingMethod.Alg() != "ES256" && signingMethod.Alg() != "ES384" && signingMethod.Alg() != "ES512" {
		zap.L().Fatal("Invalid ECDSA signing method", zap.String("method", signingMethod.Alg()))
	}

	if privateKey == nil {
		zap.L().Fatal("Invalid ECDSA private key")
	}

	if publicKey == nil {
		zap.L().Fatal("Invalid ECDSA public key")
	}

	c.signingMethod = signingMethod
	c.privateKey = privateKey
	c.publicKey = publicKey
	c.parser = defaultECDSAParser(publicKey)
	c.signer = defaultECDSASigner(privateKey)
	return c
}

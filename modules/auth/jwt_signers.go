package auth

import (
	"crypto/ecdsa"
	"crypto/rsa"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nayefradwi/nayef_go_common/core"
	"go.uber.org/zap"
)

var (
	nilTokenError = core.InternalError("trying to sign nil token using")
)

type tokenSigner func(token *jwt.Token) (string, error)

func defaultHMACSigner(secretKey string) tokenSigner {
	return func(token *jwt.Token) (string, error) {
		if token == nil {
			return "", nilTokenError
		}
		return token.SignedString([]byte(secretKey))
	}
}

func defaultRSASigner(privateKey interface{}) tokenSigner {
	privateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		zap.L().Fatal("Invalid RSA private key")
	}

	return func(token *jwt.Token) (string, error) {
		if token == nil {
			return "", nilTokenError
		}

		return token.SignedString(privateKey)
	}
}

func defaultECDSASigner(privateKey interface{}) tokenSigner {
	privateKey, ok := privateKey.(*ecdsa.PrivateKey)
	if !ok {
		zap.L().Fatal("Invalid ECDSA private key")
	}

	return func(token *jwt.Token) (string, error) {
		if token == nil {
			return "", nilTokenError
		}

		return token.SignedString(privateKey)
	}
}

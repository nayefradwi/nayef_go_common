package auth

import (
	"crypto/ecdsa"
	"crypto/rsa"

	"github.com/golang-jwt/jwt/v5"
	. "github.com/nayefradwi/nayef_go_common/errors"
)

var (
	nilTokenError = InternalError("trying to sign nil token using")
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

func defaultRSASigner(key *rsa.PrivateKey) tokenSigner {
	return func(token *jwt.Token) (string, error) {
		if token == nil {
			return "", nilTokenError
		}
		return token.SignedString(key)
	}
}

func defaultECDSASigner(key *ecdsa.PrivateKey) tokenSigner {
	return func(token *jwt.Token) (string, error) {
		if token == nil {
			return "", nilTokenError
		}
		return token.SignedString(key)
	}
}

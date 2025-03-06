package auth

import (
	"crypto/ecdsa"
	"crypto/rsa"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nayefradwi/nayef_go_common/core"
)

func defaultHMACParser(secretKey string) func(token *jwt.Token) (interface{}, error) {
	return func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, core.UnauthorizedError("Invalid token")
		}
		return []byte(secretKey), nil
	}
}

func defaultRSAParser(publicKey *rsa.PublicKey) func(token *jwt.Token) (interface{}, error) {
	return func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, core.UnauthorizedError("Invalid token")
		}
		return publicKey, nil
	}
}

func defaultECDSAParser(publicKey *ecdsa.PublicKey) func(token *jwt.Token) (interface{}, error) {
	return func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, core.UnauthorizedError("Invalid token")
		}
		return publicKey, nil
	}
}

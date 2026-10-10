package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"

	. "github.com/nayefradwi/nayef_go_common/errors"
)

const opaqueTokenBytes = 32

func NewOpaqueToken() (string, []byte, error) {
	b := make([]byte, opaqueTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", nil, InternalError("failed to generate token: " + err.Error())
	}

	token := base64.RawURLEncoding.EncodeToString(b)
	return token, HashOpaqueToken(token), nil
}

func HashOpaqueToken(token string) []byte {
	h := sha256.Sum256([]byte(token))
	return h[:]
}

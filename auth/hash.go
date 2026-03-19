package auth

import (
	. "github.com/nayefradwi/nayef_go_common/errors"
	"golang.org/x/crypto/bcrypt"
)

type HashingConfig struct {
	Salt int
}

const defaultSalt = 10

var DefaultHashingConfig = NewHashingConfig(defaultSalt)

func NewHashingConfig(salt int) HashingConfig {
	return HashingConfig{Salt: salt}
}

const maxBcryptPasswordBytes = 72

func (hc HashingConfig) Hash(password string) (string, error) {
	if len([]byte(password)) > maxBcryptPasswordBytes {
		return "", BadRequestError("password exceeds maximum length of 72 bytes")
	}
	salt := hc.getSalt()
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), salt)
	return string(bytes), err
}

func (hc HashingConfig) getSalt() int {
	if hc.Salt == 0 {
		return defaultSalt
	}

	return hc.Salt
}

func CompareHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

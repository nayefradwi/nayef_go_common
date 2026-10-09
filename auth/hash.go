package auth

import (
	"fmt"
	"unicode/utf8"

	. "github.com/nayefradwi/nayef_go_common/errors"
	"golang.org/x/crypto/bcrypt"
)

type HashingConfig struct {
	Cost int
}

const defaultCost = 10

var DefaultHashingConfig = NewHashingConfig(defaultCost)

func NewHashingConfig(cost int) HashingConfig {
	return HashingConfig{Cost: cost}
}

const maxBcryptPasswordBytes = 72

func (hc HashingConfig) Hash(password string) (string, error) {
	if len([]byte(password)) > maxBcryptPasswordBytes {
		return "", BadRequestError("password exceeds maximum length of 72 bytes")
	}
	cost := hc.getCost()
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(bytes), err
}

func (hc HashingConfig) getCost() int {
	if hc.Cost == 0 {
		return defaultCost
	}

	return hc.Cost
}

func CompareHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

type PasswordPolicy struct {
	MinLength int
}

var DefaultPasswordPolicy = PasswordPolicy{MinLength: 8}

func (p PasswordPolicy) Validate(password string) error {
	if utf8.RuneCountInString(password) < p.MinLength {
		return NewValidationError(Field("password", "too_short", fmt.Sprintf("password must be at least %d characters", p.MinLength)))
	}
	if len(password) > maxBcryptPasswordBytes {
		return NewValidationError(Field("password", "too_long", "password exceeds maximum length of 72 bytes"))
	}
	return nil
}

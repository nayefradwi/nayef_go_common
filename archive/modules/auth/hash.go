package auth

import "golang.org/x/crypto/bcrypt"

type HashingConfig struct {
	Salt int
}

const defaultSalt = 10

var DefaultHashingConfig = NewHashingConfig(defaultSalt)

func NewHashingConfig(salt int) HashingConfig {
	return HashingConfig{Salt: salt}
}

func (hc HashingConfig) Hash(password string) (string, error) {
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

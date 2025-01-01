package common

import (
	"golang.org/x/crypto/bcrypt"
)

func Hash(password string) (string, error) {
	return HashWithSalt(password, 10)
}

func HashWithSalt(password string, salt int) (string, error) {
	salt = getSalt(salt)
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), salt)
	return string(bytes), err
}

func getSalt(salt int) int {
	if salt == 0 {
		salt = 10
	}
	return salt
}

func CompareHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

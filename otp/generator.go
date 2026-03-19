package otp

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
)

type CodeGenerator struct {
	Length   int
	HasAlpha bool
}

func NewCodeGenerator(length int, hasAlpha bool) CodeGenerator {
	return CodeGenerator{
		Length:   length,
		HasAlpha: hasAlpha,
	}
}

func (g CodeGenerator) GenerateOtp() string {
	if g.HasAlpha {
		return g.generateAlphaNumeric()
	}

	return g.generateNumeric()
}

func (g CodeGenerator) generateNumeric() string {
	const digits = "0123456789"
	return g.generateFromCharset(digits)
}

func (g CodeGenerator) generateAlphaNumeric() string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	return g.generateFromCharset(alphanum)
}

func (g CodeGenerator) generateFromCharset(charset string) string {
	otp := make([]byte, g.Length)
	for i := range otp {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		otp[i] = charset[n.Int64()]
	}

	return string(otp)
}

func HashCode(code string) string {
	h := sha256.Sum256([]byte(code))
	return hex.EncodeToString(h[:])
}

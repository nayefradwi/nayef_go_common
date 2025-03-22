package otp

import (
	"math/rand"
	"time"
)

type CodeGenerator struct {
	Length   int
	HasAlpha bool
	rng      *rand.Rand
}

func NewCodeGenerator(length int, hasAlpha bool) CodeGenerator {
	src := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(src)

	return CodeGenerator{
		Length:   length,
		HasAlpha: hasAlpha,
		rng:      rng,
	}
}

func (g CodeGenerator) GenerateOtp() string {
	if g.HasAlpha {
		return g.generateAlphaNumeric()
	}

	return g.generateNumeric()
}

func (g CodeGenerator) generateAlphaNumeric() string {
	const digits = "0123456789"
	return g.generateFromCharset(digits)
}

func (g CodeGenerator) generateNumeric() string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	return g.generateFromCharset(alphanum)
}

func (g CodeGenerator) generateFromCharset(charset string) string {
	otp := make([]byte, g.Length)
	for i := range otp {
		otp[i] = charset[g.rng.Intn(len(charset))]
	}

	return string(otp)
}

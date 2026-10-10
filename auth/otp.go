package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"math/big"
	"time"

	. "github.com/nayefradwi/nayef_go_common/errors"
)

var (
	ErrCodeNotFound      = errors.New("code not found")
	ErrCodeResendBlocked = errors.New("code resend blocked")
)

type CodeLimits struct {
	TTL         time.Duration
	ResendAfter time.Duration
	MaxSends    int
	MaxAttempts int
}

type OtpConfig struct {
	Secret []byte
	Length int
	CodeLimits
}

type OtpManager struct {
	store  CodeStore
	config OtpConfig
}

const (
	minOtpLength       = 6
	minOtpSecretLength = 32
)

func NewOtpManager(store CodeStore, config OtpConfig) (OtpManager, error) {
	if store == nil {
		return OtpManager{}, BadRequestError("otp store must not be nil")
	}

	if len(config.Secret) < minOtpSecretLength {
		return OtpManager{}, BadRequestError("otp secret must be at least 32 bytes")
	}

	if config.Length < minOtpLength {
		return OtpManager{}, BadRequestError("otp length must be at least 6")
	}

	if err := validLimits(config.CodeLimits); err != nil {
		return OtpManager{}, err
	}

	return OtpManager{store: store, config: config}, nil
}

func (m OtpManager) Issue(ctx context.Context, key string) (string, error) {
	code, err := newCode(m.config.Length)
	if err != nil {
		return "", err
	}

	if err := m.store.Save(ctx, key, hmacSum(m.config.Secret, code), m.config.CodeLimits); err != nil {
		return "", err
	}

	return code, nil
}

func (m OtpManager) Verify(ctx context.Context, key, code string) error {
	stored, err := checkCode(ctx, m.store, key, m.config.MaxAttempts, hmacSum(m.config.Secret, code))
	if err != nil {
		return err
	}

	return consumeCode(ctx, m.store, key, stored)
}

func (m OtpManager) DeleteExpired(ctx context.Context) error {
	return m.store.DeleteExpired(ctx)
}

func hmacSum(secret []byte, s string) []byte {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(s))
	return h.Sum(nil)
}

var otpDigits = big.NewInt(10)

func newCode(length int) (string, error) {
	code := make([]byte, length)
	for i := range code {
		n, err := rand.Int(rand.Reader, otpDigits)
		if err != nil {
			return "", InternalError("failed to generate otp: " + err.Error())
		}
		code[i] = byte('0' + n.Int64())
	}

	return string(code), nil
}

func validLimits(l CodeLimits) error {
	if l.TTL <= 0 || l.ResendAfter < 0 || l.MaxSends < 1 || l.MaxAttempts < 1 {
		return BadRequestError("code needs a positive ttl, a non-negative resend gap and limits of at least 1")
	}
	return nil
}

func checkCode(ctx context.Context, store CodeStore, key string, maxAttempts int, hash []byte) ([]byte, error) {
	stored, err := store.Attempt(ctx, key, maxAttempts)
	if errors.Is(err, ErrCodeNotFound) {
		return nil, UnauthorizedError("Invalid code")
	}

	if err != nil {
		return nil, err
	}

	if !hmac.Equal(stored, hash) {
		return nil, UnauthorizedError("Invalid code")
	}

	return stored, nil
}

func consumeCode(ctx context.Context, store CodeStore, key string, hash []byte) error {
	err := store.Consume(ctx, key, hash)
	if errors.Is(err, ErrCodeNotFound) {
		return UnauthorizedError("Invalid code")
	}

	return err
}

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
	ErrOtpNotFound      = errors.New("otp not found")
	ErrOtpResendBlocked = errors.New("otp resend blocked")
)

type OtpLimits struct {
	TTL         time.Duration
	ResendAfter time.Duration
	MaxSends    int
	MaxAttempts int
}

type OtpConfig struct {
	Secret []byte
	Length int
	OtpLimits
}

type OtpManager struct {
	store  OtpStore
	config OtpConfig
}

const (
	minOtpLength       = 6
	minOtpSecretLength = 32
)

func NewOtpManager(store OtpStore, config OtpConfig) (OtpManager, error) {
	if store == nil {
		return OtpManager{}, BadRequestError("otp store must not be nil")
	}

	if len(config.Secret) < minOtpSecretLength {
		return OtpManager{}, BadRequestError("otp secret must be at least 32 bytes")
	}

	if config.Length < minOtpLength {
		return OtpManager{}, BadRequestError("otp length must be at least 6")
	}

	l := config.OtpLimits
	if l.TTL <= 0 || l.ResendAfter < 0 || l.MaxSends < 1 || l.MaxAttempts < 1 {
		return OtpManager{}, BadRequestError("otp needs a positive ttl, a non-negative resend gap and limits of at least 1")
	}

	return OtpManager{store: store, config: config}, nil
}

func (m OtpManager) Issue(ctx context.Context, key string) (string, error) {
	code, err := newCode(m.config.Length)
	if err != nil {
		return "", err
	}

	if err := m.store.Save(ctx, key, m.hash(code), m.config.OtpLimits); err != nil {
		return "", err
	}

	return code, nil
}

func (m OtpManager) Verify(ctx context.Context, key, code string) error {
	stored, err := m.store.Attempt(ctx, key, m.config.MaxAttempts)
	if errors.Is(err, ErrOtpNotFound) {
		return UnauthorizedError("Invalid code")
	}

	if err != nil {
		return err
	}

	if !hmac.Equal(stored, m.hash(code)) {
		return UnauthorizedError("Invalid code")
	}

	err = m.store.Consume(ctx, key, stored)
	if errors.Is(err, ErrOtpNotFound) {
		return UnauthorizedError("Invalid code")
	}

	return err
}

func (m OtpManager) DeleteExpired(ctx context.Context) error {
	return m.store.DeleteExpired(ctx)
}

func (m OtpManager) hash(code string) []byte {
	h := hmac.New(sha256.New, m.config.Secret)
	h.Write([]byte(code))
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

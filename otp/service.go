package otp

import (
	"context"
	"crypto/subtle"
	"log/slog"
	"time"

	"github.com/nayefradwi/nayef_go_common/errors"
	"github.com/redis/go-redis/v9"
)

type OtpConfig struct {
	ExpiresIn   time.Duration
	MaxTries    int
	ResendAfter time.Duration
}

type OtpService struct {
	otpRepository IOtpRepository
	codeGenerator ICodeGenerator
	config        OtpConfig
}

func NewOtpService(otpRepository IOtpRepository, codeGenerator ICodeGenerator, config OtpConfig) IOtpService {
	return &OtpService{
		otpRepository: otpRepository,
		codeGenerator: codeGenerator,
		config:        config,
	}
}

func (s *OtpService) GenerateOtp(ctx context.Context, ownerId string) (*OTP, error) {
	o, err := s.getOtp(ctx, ownerId)
	if err != nil {
		return nil, err
	}

	if o != nil && !s.shouldGenerateNewOtp(o) {
		return o, nil
	}

	o = s.generateNewOtp(ownerId)
	if err := s.otpRepository.UpsertOtp(ctx, o); err != nil {
		slog.ErrorContext(ctx, "failed to upsert otp", "error", err)
		return nil, errors.InternalError("failed to create otp")
	}

	return o, nil
}

func (s *OtpService) VerifyOtp(ctx context.Context, ownerId, code string) error {
	o, err := s.getOtp(ctx, ownerId)
	if err != nil {
		return err
	}

	if o == nil {
		return ErrOtpNotFound
	}

	if o.IsExpired() {
		return ErrExpiredOTP
	}

	if o.IsMaxRetriesReached() {
		return ErrMaxTriesExceeded
	}

	hashedInput := HashCode(code)
	match := subtle.ConstantTimeCompare([]byte(o.HashedCode), []byte(hashedInput)) == 1
	if match {
		return nil
	}

	o.IncrementRetryCount()
	if err := s.otpRepository.UpsertOtp(ctx, o); err != nil {
		slog.ErrorContext(ctx, "failed to persist retry count", "error", err)
	}

	return ErrIncorrectOTP
}

func (s *OtpService) getOtp(ctx context.Context, ownerId string) (*OTP, error) {
	o, err := s.otpRepository.GetOtp(ctx, ownerId)
	if err == nil {
		return o, nil
	}

	if err == redis.Nil {
		return nil, nil
	}

	slog.ErrorContext(ctx, "failed to get otp", "error", err)
	return nil, errors.InternalError("failed to get otp")
}

func (s *OtpService) generateNewOtp(ownerId string) *OTP {
	code := s.codeGenerator.GenerateOtp()
	hashedCode := HashCode(code)
	now := time.Now().UTC()
	expiresAt := now.Add(s.config.ExpiresIn)
	return NewOtp(ownerId, code, hashedCode, 0, s.config.MaxTries, expiresAt, now)
}

func (s *OtpService) shouldGenerateNewOtp(o *OTP) bool {
	if o.IsExpired() {
		return true
	}
	return time.Now().UTC().After(o.UpdatedAt.Add(s.config.ResendAfter))
}

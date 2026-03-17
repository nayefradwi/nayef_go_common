package otp

import (
	"context"
	"time"

	"github.com/nayefradwi/nayef_go_common/modules/otp"
	"github.com/nayefradwi/nayef_go_common/result"
	"go.uber.org/zap"
)

type OtpConfig struct {
	ExpiresInSeconds int
	MaxTries         int
	ResendAfter      int
}

type OtpService struct {
	otpRepository otp.IOtpRepository
	codeGenerator otp.ICodeGenerator
	config        OtpConfig
}

func NewOtpService(otpRepository otp.IOtpRepository, codeGenerator otp.ICodeGenerator, config OtpConfig) otp.IOtpService {
	return &OtpService{
		otpRepository: otpRepository,
		codeGenerator: codeGenerator,
		config:        config,
	}
}

func (s *OtpService) GenerateOtp(ctx context.Context, ownerId string) (*otp.OTP, error) {
	o, err := s.getOtp(ctx, ownerId)
	if err != nil {
		return nil, err
	}

	if s.shouldResendOtp(o) {
		return o, nil
	}

	o = s.generateNewOtp(ownerId)
	if err := s.otpRepository.UpsertOtp(ctx, o); err != nil {
		zap.L().Error("failed to upsert otp", zap.Error(err))
		return nil, result.InternalError("failed to create otp")
	}

	return o, nil
}

func (s *OtpService) VerifyOtp(ctx context.Context, ownerId, code string) error {
	o, err := s.getOtp(ctx, ownerId)
	if err != nil {
		return err
	}

	if o.IsExpired() {
		return otp.ErrExpiredOTP
	}

	if o.IsMaxRetriesReached() {
		return otp.ErrMaxTriesExceeded
	}

	o.IncrementRetryCount()
	s.otpRepository.UpsertOtp(ctx, o)

	if o.Code != code {
		return otp.ErrIncorrectOTP
	}

	return nil
}

func (s *OtpService) getOtp(ctx context.Context, ownerId string) (*otp.OTP, error) {
	otp, err := s.otpRepository.GetOtp(ctx, ownerId)
	if err != nil {
		zap.L().Error("failed to get otp", zap.Error(err))
		return nil, result.InternalError("failed to get otp")
	}

	return otp, nil
}

func (s *OtpService) generateNewOtp(ownerId string) *otp.OTP {
	code := s.codeGenerator.GenerateOtp()
	expiresAt := time.Now().UTC().Add(time.Second * time.Duration(s.config.ExpiresInSeconds))
	return otp.NewOtp(ownerId, code, 0, s.config.MaxTries, expiresAt, time.Now().UTC())
}

func (s *OtpService) shouldResendOtp(o *otp.OTP) bool {
	return o.IsExpired() || time.Now().UTC().After(o.UpdateAt.Add(time.Second*time.Duration(s.config.ResendAfter)))
}

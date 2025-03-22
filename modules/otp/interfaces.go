package otp

import "context"

type ICodeGenerator interface {
	GenerateOtp() string
}

type IOtpRepository interface {
	GetOtp(ctx context.Context, ownerId string) (*OTP, error)
	UpsertOtp(ctx context.Context, otp *OTP) error
}

type IOtpService interface {
	GenerateOtp(ctx context.Context, ownerId string) (*OTP, error)
	VerifyOtp(ctx context.Context, ownerId, code string) error
}

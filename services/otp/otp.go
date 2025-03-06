package otp

import "context"

type IOtpGenerator interface {
	GenerateOtp() string
}

type IOtpService interface {
	SendOtp(ctx context.Context, ownerId string) (string, error)
	VerifyOtp(ctx context.Context, ownerId, otp string) error
}

type OtpGenerator struct {
	Max      int
	HasAlpha bool
}

func NewOtpGenerator(max int, hasAlpha bool) OtpGenerator {
	return OtpGenerator{
		Max:      max,
		HasAlpha: hasAlpha,
	}
}

func (g OtpGenerator) GenerateOtp() string {
	// Implementation here
	return ""
}

package otp

import "github.com/nayefradwi/nayef_go_common/result"

const (
	IncorrectOtpErrorCode = "INCORRECT_OTP"
	MaxTriesExceededCode  = "MAX_TRIES_EXCEEDED"
	ExpiredOtpCode        = "EXPIRED_OTP"
)

var (
	ErrIncorrectOTP     = result.NewResultError("incorrect otp", IncorrectOtpErrorCode)
	ErrMaxTriesExceeded = result.NewResultError("max tries exceeded", MaxTriesExceededCode)
	ErrExpiredOTP       = result.NewResultError("expired otp", ExpiredOtpCode)
)

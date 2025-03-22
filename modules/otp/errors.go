package otp

import "github.com/nayefradwi/nayef_go_common/core"

const (
	IncorrectOtpErrorCode = "INCORRECT_OTP"
	MaxTriesExceededCode  = "MAX_TRIES_EXCEEDED"
	ExpiredOtpCode        = "EXPIRED_OTP"
)

var (
	ErrIncorrectOTP     = core.NewResultError("incorrect otp", IncorrectOtpErrorCode)
	ErrMaxTriesExceeded = core.NewResultError("max tries exceeded", MaxTriesExceededCode)
	ErrExpiredOTP       = core.NewResultError("expired otp", ExpiredOtpCode)
)

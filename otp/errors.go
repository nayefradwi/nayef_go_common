package otp

import "github.com/nayefradwi/nayef_go_common/errors"

const (
	IncorrectOtpErrorCode = "INCORRECT_OTP"
	MaxTriesExceededCode  = "MAX_TRIES_EXCEEDED"
	ExpiredOtpCode        = "EXPIRED_OTP"
	OtpNotFoundCode       = "OTP_NOT_FOUND"
)

var (
	ErrIncorrectOTP     = errors.NewResultError("incorrect otp", IncorrectOtpErrorCode)
	ErrMaxTriesExceeded = errors.NewResultError("max tries exceeded", MaxTriesExceededCode)
	ErrExpiredOTP       = errors.NewResultError("expired otp", ExpiredOtpCode)
	ErrOtpNotFound      = errors.NewResultError("otp not found, request a new one", OtpNotFoundCode)
)

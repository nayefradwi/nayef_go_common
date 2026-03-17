package otp

import "time"

type OTP struct {
	OwnerId    string    `json:"owner_id"`
	Code       string    `json:"code"`
	RetryCount int       `json:"retry_count"`
	MaxRetries int       `json:"max_retries"`
	ExpiresAt  time.Time `json:"expires_at"`
	UpdateAt   time.Time `json:"updated_at"`
}

func NewOtp(ownerId, code string, retryCount, maxRetries int, expiresAt, updatedAt time.Time) *OTP {
	return &OTP{
		OwnerId:    ownerId,
		Code:       code,
		RetryCount: retryCount,
		MaxRetries: maxRetries,
		ExpiresAt:  expiresAt,
		UpdateAt:   updatedAt,
	}
}

func (o *OTP) IsExpired() bool {
	if o == nil {
		return true
	}

	if o.ExpiresAt.IsZero() {
		return true
	}

	return time.Now().UTC().After(o.ExpiresAt)
}

func (o *OTP) IncrementRetryCount() {
	if o == nil {
		return
	}

	o.RetryCount++
	o.UpdateAt = time.Now().UTC()
}

func (o *OTP) IsMaxRetriesReached() bool {
	if o == nil {
		return true
	}

	return o.RetryCount >= o.MaxRetries
}

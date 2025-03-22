package otp

import (
	"context"
	"time"

	"github.com/nayefradwi/nayef_go_common/modules/dates"
	"github.com/nayefradwi/nayef_go_common/modules/otp"
	"github.com/redis/go-redis/v9"
)

type RedisOtpRepository struct {
	client *redis.Client
}

func NewRedisOtpRepository(client *redis.Client) otp.IOtpRepository {
	return &RedisOtpRepository{client: client}
}

func (r *RedisOtpRepository) UpsertOtp(ctx context.Context, o *otp.OTP) error {
	if o == nil {
		return nil
	}
	remaining := dates.RemainingSeconds(o.ExpiresAt)
	if remaining <= 0 {
		return otp.ErrExpiredOTP
	}

	return r.client.SetEx(ctx, o.OwnerId, o, time.Duration(remaining)).Err()
}

func (r *RedisOtpRepository) GetOtp(ctx context.Context, ownerId string) (*otp.OTP, error) {
	otp := &otp.OTP{}
	if err := r.client.Get(ctx, ownerId).Scan(otp); err != nil {
		return nil, err
	}

	return otp, nil
}

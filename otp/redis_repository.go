package otp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const redisKeyPrefix = "otp:"

type RedisOtpRepository struct {
	client *redis.Client
}

func NewRedisOtpRepository(client *redis.Client) IOtpRepository {
	return &RedisOtpRepository{client: client}
}

func (r *RedisOtpRepository) UpsertOtp(ctx context.Context, o *OTP) error {
	if o == nil {
		return fmt.Errorf("otp: cannot upsert nil OTP")
	}

	remaining := time.Until(o.ExpiresAt)
	if remaining <= 0 {
		return ErrExpiredOTP
	}

	data, err := json.Marshal(o)
	if err != nil {
		return fmt.Errorf("otp: failed to marshal OTP: %w", err)
	}

	return r.client.Set(ctx, redisKeyPrefix+o.OwnerId, data, remaining).Err()
}

func (r *RedisOtpRepository) GetOtp(ctx context.Context, ownerId string) (*OTP, error) {
	data, err := r.client.Get(ctx, redisKeyPrefix+ownerId).Bytes()
	if err != nil {
		return nil, err
	}

	o := &OTP{}
	if err := json.Unmarshal(data, o); err != nil {
		return nil, fmt.Errorf("otp: failed to unmarshal OTP: %w", err)
	}

	return o, nil
}

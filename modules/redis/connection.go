package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type ConnectionConfig struct {
	opts *redis.Options
	err  error
}

func NewConnectionConfig(url string) ConnectionConfig {
	opts, err := redis.ParseURL(url)
	return ConnectionConfig{
		opts,
		err,
	}
}

func (cc ConnectionConfig) Connect(ctx context.Context) *redis.Client {
	zap.L().Debug("connecting to redis...")
	if cc.err != nil {
		zap.L().Fatal("failed to parse redis connection url", zap.Error(cc.err))
	}

	redisClient := redis.NewClient(cc.opts)
	_, connectionErr := redisClient.Ping(ctx).Result()
	if connectionErr != nil {
		zap.L().Fatal("failed to set up redis connection", zap.Error(connectionErr))
	}
	zap.L().Info("connected to redis successfully")
	return redisClient
}

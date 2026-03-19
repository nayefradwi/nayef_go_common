package redisutil

import (
	"context"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

func ConnectToRedis(ctx context.Context, url string) *redis.Client {
	opts, err := redis.ParseURL(url)
	if err != nil {
		slog.Error("failed to connect to redis", "error", err.Error())
		panic(err)
	}

	client := redis.NewClient(opts)
	if _, err := client.Ping(ctx).Result(); err != nil {
		slog.Error("failed to ping redis connection", "error", err.Error())
		panic(err)
	}

	slog.Info("connected to redis successfully", "address", opts.Addr)
	return client
}

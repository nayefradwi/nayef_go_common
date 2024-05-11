package common

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// later I can have a nice way of passing the connection url
func ConnectDatabasePool(ctx context.Context, connectionUrl string) *pgxpool.Pool {
	dbPool, err := pgxpool.Connect(ctx, connectionUrl)
	if err != nil {
		GetLogger().Panic("failed to set up db connection", zap.Error(err))
	}
	GetLogger().Info("connected to db successfully")
	return dbPool
}

func ConnectRedis(ctx context.Context, connectionUrl string) *redis.Client {
	opt, parsingErr := redis.ParseURL(connectionUrl)
	if parsingErr != nil {
		GetLogger().Panic("failed to parse redis connection url", zap.Error(parsingErr))
	}
	opt.MaxRetries = 5
	redisClient := redis.NewClient(opt)
	_, connectionErr := redisClient.Ping(ctx).Result()
	if connectionErr != nil {
		GetLogger().Panic("failed to set up redis connection", zap.Error(connectionErr))
	}
	GetLogger().Info("connected to redis successfully")
	return redisClient
}

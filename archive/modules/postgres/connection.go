package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type ConnectionConfig struct {
	*pgxpool.Config
	err error
}

func NewConnectionConfig(url string) ConnectionConfig {
	config, err := pgxpool.ParseConfig(url)
	return ConnectionConfig{
		config,
		err,
	}
}

func (cc ConnectionConfig) Connect(ctx context.Context) *pgxpool.Pool {
	zap.L().Info("connecting to postgres...")
	if cc.err != nil {
		zap.L().Fatal("error parsing config", zap.Error(cc.err))
	}

	pool, err := pgxpool.NewWithConfig(ctx, cc.Config)
	if err != nil {
		zap.L().Fatal("error connecting to postgres", zap.Error(err))
	}

	testConn, err := pool.Acquire(ctx)
	if err != nil {
		zap.L().Fatal("error acquiring test connection", zap.Error(err))
	}
	testConn.Release()

	zap.L().Info("connected to postgres")
	return pool
}

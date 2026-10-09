package authpg

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nayefradwi/nayef_go_common/auth"
	. "github.com/nayefradwi/nayef_go_common/errors"
)

type AttemptStore struct {
	pool *pgxpool.Pool
}

func NewAttemptStore(pool *pgxpool.Pool) AttemptStore {
	return AttemptStore{pool: pool}
}

func (s AttemptStore) Hit(ctx context.Context, key string, window time.Duration) (auth.Attempt, error) {
	var a auth.Attempt
	err := s.pool.QueryRow(ctx, `
		INSERT INTO auth_attempts (key, count, reset_at) VALUES ($1, 1, now() + $2)
		ON CONFLICT (key) DO UPDATE SET
			count    = CASE WHEN auth_attempts.reset_at <= now() THEN 1 ELSE auth_attempts.count + 1 END,
			reset_at = CASE WHEN auth_attempts.reset_at <= now() THEN now() + $2 ELSE auth_attempts.reset_at END
		RETURNING count, reset_at`, key, window,
	).Scan(&a.Count, &a.ResetAt)
	if err != nil {
		return auth.Attempt{}, InternalError("failed to record attempt: " + err.Error())
	}
	return a, nil
}

func (s AttemptStore) Reset(ctx context.Context, key string) error {
	if _, err := s.pool.Exec(ctx, `DELETE FROM auth_attempts WHERE key = $1`, key); err != nil {
		return InternalError("failed to reset attempts: " + err.Error())
	}
	return nil
}

func (s AttemptStore) DeleteExpired(ctx context.Context) error {
	if _, err := s.pool.Exec(ctx, `DELETE FROM auth_attempts WHERE reset_at <= now()`); err != nil {
		return InternalError("failed to delete expired attempts: " + err.Error())
	}
	return nil
}

var _ auth.AttemptStore = AttemptStore{}

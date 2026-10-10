package authredis

import (
	"context"
	"time"

	"github.com/nayefradwi/nayef_go_common/auth"
	. "github.com/nayefradwi/nayef_go_common/errors"
	"github.com/redis/go-redis/v9"
)

const attemptPrefix = "auth:attempts:"

// NX sets the window on the first hit only
var hitScript = redis.NewScript(`
local n = redis.call('INCR', KEYS[1])
redis.call('PEXPIRE', KEYS[1], ARGV[1], 'NX')
return {n, redis.call('PTTL', KEYS[1])}
`)

type AttemptStore struct {
	client *redis.Client
}

func NewAttemptStore(client *redis.Client) AttemptStore {
	return AttemptStore{client: client}
}

func (s AttemptStore) Hit(ctx context.Context, key string, window time.Duration) (auth.Attempt, error) {
	res, err := hitScript.Run(ctx, s.client, []string{attemptPrefix + key}, window.Milliseconds()).Int64Slice()
	if err != nil {
		return auth.Attempt{}, InternalError("failed to record attempt: " + err.Error())
	}

	return auth.Attempt{Count: int(res[0]), ResetAt: time.Now().Add(time.Duration(res[1]) * time.Millisecond)}, nil
}

func (s AttemptStore) Reset(ctx context.Context, key string) error {
	if err := s.client.Del(ctx, attemptPrefix+key).Err(); err != nil {
		return InternalError("failed to reset attempts: " + err.Error())
	}
	return nil
}

// keys expire on their own
func (s AttemptStore) DeleteExpired(ctx context.Context) error {
	return nil
}

var _ auth.AttemptStore = AttemptStore{}

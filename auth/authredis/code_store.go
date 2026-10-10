package authredis

import (
	"context"
	"errors"

	"github.com/nayefradwi/nayef_go_common/auth"
	. "github.com/nayefradwi/nayef_go_common/errors"
	"github.com/redis/go-redis/v9"
)

// same rule as authpg: resend only after ResendAfter, under the send cap, and while not locked
var saveCodeScript = redis.NewScript(luaNow + `
local now = nowMs()
local v = redis.call('HMGET', KEYS[1], 'attempts', 'sends', 'sent_at')
if v[1] then
	if tonumber(v[3]) + tonumber(ARGV[3]) > now
		or tonumber(v[2]) >= tonumber(ARGV[4])
		or tonumber(v[1]) >= tonumber(ARGV[5]) then
		return 0
	end
	redis.call('HSET', KEYS[1], 'hash', ARGV[1], 'sent_at', string.format('%d', now))
	redis.call('HINCRBY', KEYS[1], 'sends', 1)
else
	redis.call('HSET', KEYS[1], 'hash', ARGV[1], 'attempts', 0, 'sends', 1, 'sent_at', string.format('%d', now))
end
redis.call('PEXPIRE', KEYS[1], ARGV[2])
return 1
`)

var attemptCodeScript = redis.NewScript(`
local v = redis.call('HMGET', KEYS[1], 'hash', 'attempts')
if not v[1] or tonumber(v[2]) >= tonumber(ARGV[1]) then
	return false
end
redis.call('HINCRBY', KEYS[1], 'attempts', 1)
return v[1]
`)

var consumeCodeScript = redis.NewScript(`
if redis.call('HGET', KEYS[1], 'hash') ~= ARGV[1] then
	return 0
end
return redis.call('DEL', KEYS[1])
`)

type CodeStore struct {
	client *redis.Client
	prefix string
}

type CodeStoreConfig struct {
	Prefix string
}

const defaultCodePrefix = "auth:otps:"

var DefaultCodeStoreConfig = CodeStoreConfig{
	Prefix: defaultCodePrefix,
}

func NewCodeStore(client *redis.Client, configs ...CodeStoreConfig) (CodeStore, error) {
	config := DefaultCodeStoreConfig
	if len(configs) > 0 {
		config = configs[0]
	}

	if err := validPrefix(config.Prefix); err != nil {
		return CodeStore{}, err
	}

	return CodeStore{client: client, prefix: config.Prefix}, nil
}

func (s CodeStore) Save(ctx context.Context, key string, hash []byte, l auth.CodeLimits) error {
	saved, err := saveCodeScript.Run(ctx, s.client, []string{s.prefix + key},
		hash, l.TTL.Milliseconds(), l.ResendAfter.Milliseconds(), l.MaxSends, l.MaxAttempts,
	).Int()
	if err != nil {
		return InternalError("failed to save code: " + err.Error())
	}

	if saved == 0 {
		return auth.ErrCodeResendBlocked
	}

	return nil
}

func (s CodeStore) Attempt(ctx context.Context, key string, maxAttempts int) ([]byte, error) {
	hash, err := attemptCodeScript.Run(ctx, s.client, []string{s.prefix + key}, maxAttempts).Text()
	if errors.Is(err, redis.Nil) {
		return nil, auth.ErrCodeNotFound
	}

	if err != nil {
		return nil, InternalError("failed to attempt code: " + err.Error())
	}

	return []byte(hash), nil
}

func (s CodeStore) Consume(ctx context.Context, key string, hash []byte) error {
	deleted, err := consumeCodeScript.Run(ctx, s.client, []string{s.prefix + key}, hash).Int()
	if err != nil {
		return InternalError("failed to consume code: " + err.Error())
	}

	if deleted == 0 {
		return auth.ErrCodeNotFound
	}

	return nil
}

// keys expire on their own
func (s CodeStore) DeleteExpired(ctx context.Context) error {
	return nil
}

var _ auth.CodeStore = CodeStore{}

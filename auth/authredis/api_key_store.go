package authredis

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/nayefradwi/nayef_go_common/auth"
	. "github.com/nayefradwi/nayef_go_common/errors"
	"github.com/redis/go-redis/v9"
)

// keys: k:<id> api key, h:<hash> → id, o:<owner> ids by creation time, expiring with its longest-lived key
// KEYS: key, hash index, owner set; ARGV: id, owner, name, hash, claims, expiry, last used, created, score, prefix
var createApiKeyScript = redis.NewScript(`
redis.call('HSET', KEYS[1],
	'id', ARGV[1], 'owner_id', ARGV[2], 'name', ARGV[3], 'hash', ARGV[4], 'claims', ARGV[5],
	'expires_at', ARGV[6], 'last_used_at', ARGV[7], 'created_at', ARGV[8])
redis.call('SET', KEYS[2], ARGV[1])
if ARGV[6] ~= '' then
	redis.call('PEXPIREAT', KEYS[1], ARGV[6])
	redis.call('PEXPIREAT', KEYS[2], ARGV[6])
end
local forever, latest = ARGV[6] == '', tonumber(ARGV[6]) or 0
for _, id in ipairs(redis.call('ZRANGE', KEYS[3], 0, -1)) do
	local at = redis.call('PEXPIRETIME', ARGV[10] .. 'k:' .. id)
	if at == -2 then
		redis.call('ZREM', KEYS[3], id)
	elseif at == -1 then
		forever = true
	elseif at > latest then
		latest = at
	end
end
redis.call('ZADD', KEYS[3], ARGV[9], ARGV[1])
if forever then
	redis.call('PERSIST', KEYS[3])
else
	redis.call('PEXPIREAT', KEYS[3], latest)
end
return 1
`)

var getApiKeyScript = redis.NewScript(`
local id = redis.call('GET', KEYS[1])
if not id then
	return false
end
return redis.call('HGETALL', ARGV[1] .. 'k:' .. id)
`)

// a plain HSET would recreate an expired key without a ttl
var touchApiKeyScript = redis.NewScript(luaNow + `
if redis.call('EXISTS', KEYS[1]) == 0 then
	return 0
end
return redis.call('HSET', KEYS[1], 'last_used_at', string.format('%d', nowMs()))
`)

// KEYS: key, owner set; ARGV: owner, id, prefix
var deleteApiKeyScript = redis.NewScript(`
local v = redis.call('HMGET', KEYS[1], 'owner_id', 'hash')
if v[1] ~= ARGV[1] then
	return 0
end
redis.call('DEL', KEYS[1], ARGV[3] .. 'h:' .. v[2])
redis.call('ZREM', KEYS[2], ARGV[2])
return 1
`)

type ApiKeyStore struct {
	client *redis.Client
	prefix string
}

type ApiKeyStoreConfig struct {
	Prefix string
}

const defaultApiKeyPrefix = "auth:api_keys:"

var DefaultApiKeyStoreConfig = ApiKeyStoreConfig{
	Prefix: defaultApiKeyPrefix,
}

func NewApiKeyStore(client *redis.Client, configs ...ApiKeyStoreConfig) (ApiKeyStore, error) {
	config := DefaultApiKeyStoreConfig
	if len(configs) > 0 {
		config = configs[0]
	}

	if err := validPrefix(config.Prefix); err != nil {
		return ApiKeyStore{}, err
	}

	return ApiKeyStore{client: client, prefix: config.Prefix}, nil
}

func (s ApiKeyStore) Create(ctx context.Context, k auth.ApiKey) error {
	claims, err := encodeClaims(k.Claims)
	if err != nil {
		return InternalError("failed to create api key: " + err.Error())
	}

	keys := []string{s.key(k.Id), s.prefix + "h:" + string(k.Hash), s.ownerKey(k.OwnerId)}
	// µs score keeps keys created in the same millisecond in order
	err = createApiKeyScript.Run(ctx, s.client, keys,
		k.Id.String(), k.OwnerId.String(), k.Name, k.Hash, claims,
		formatMs(k.ExpiresAt), formatMs(k.LastUsedAt), formatMs(k.CreatedAt), k.CreatedAt.UnixMicro(), s.prefix,
	).Err()
	if err != nil {
		return InternalError("failed to create api key: " + err.Error())
	}

	return nil
}

func (s ApiKeyStore) GetByHash(ctx context.Context, hash []byte) (auth.ApiKey, error) {
	pairs, err := getApiKeyScript.RunRO(ctx, s.client, []string{s.prefix + "h:" + string(hash)}, s.prefix).StringSlice()
	if errors.Is(err, redis.Nil) || (err == nil && len(pairs) == 0) {
		return auth.ApiKey{}, auth.ErrApiKeyNotFound
	}

	if err != nil {
		return auth.ApiKey{}, InternalError("failed to read api key: " + err.Error())
	}

	fields := make(map[string]string, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		fields[pairs[i]] = pairs[i+1]
	}

	k, err := parseApiKey(fields)
	if err != nil {
		return auth.ApiKey{}, InternalError("failed to read api key: " + err.Error())
	}

	return k, nil
}

func (s ApiKeyStore) ListByOwner(ctx context.Context, ownerId uuid.UUID) ([]auth.ApiKey, error) {
	ids, err := s.client.ZRevRange(ctx, s.ownerKey(ownerId), 0, -1).Result()
	if err != nil {
		return nil, InternalError("failed to list api keys: " + err.Error())
	}

	cmds := make([]*redis.MapStringStringCmd, len(ids))
	_, err = s.client.Pipelined(ctx, func(p redis.Pipeliner) error {
		for i, id := range ids {
			cmds[i] = p.HGetAll(ctx, s.prefix+"k:"+id)
		}
		return nil
	})
	if err != nil {
		return nil, InternalError("failed to list api keys: " + err.Error())
	}

	keys := make([]auth.ApiKey, 0, len(cmds))
	for _, cmd := range cmds {
		fields := cmd.Val()
		if len(fields) == 0 {
			continue
		}

		k, err := parseApiKey(fields)
		if err != nil {
			return nil, InternalError("failed to list api keys: " + err.Error())
		}
		keys = append(keys, k)
	}

	return keys, nil
}

func (s ApiKeyStore) Touch(ctx context.Context, id uuid.UUID) error {
	if err := touchApiKeyScript.Run(ctx, s.client, []string{s.key(id)}).Err(); err != nil {
		return InternalError("failed to touch api key: " + err.Error())
	}
	return nil
}

func (s ApiKeyStore) Delete(ctx context.Context, ownerId, id uuid.UUID) error {
	deleted, err := deleteApiKeyScript.Run(ctx, s.client, []string{s.key(id), s.ownerKey(ownerId)},
		ownerId.String(), id.String(), s.prefix,
	).Int()
	if err != nil {
		return InternalError("failed to delete api key: " + err.Error())
	}

	if deleted == 0 {
		return auth.ErrApiKeyNotFound
	}

	return nil
}

// keys expire on their own
func (s ApiKeyStore) DeleteExpired(ctx context.Context) error {
	return nil
}

func (s ApiKeyStore) key(id uuid.UUID) string {
	return s.prefix + "k:" + id.String()
}

func (s ApiKeyStore) ownerKey(ownerId uuid.UUID) string {
	return s.prefix + "o:" + ownerId.String()
}

func parseApiKey(f map[string]string) (auth.ApiKey, error) {
	k := auth.ApiKey{Name: f["name"], Hash: []byte(f["hash"])}
	var err error
	if k.Id, err = uuid.Parse(f["id"]); err != nil {
		return auth.ApiKey{}, err
	}

	if k.OwnerId, err = uuid.Parse(f["owner_id"]); err != nil {
		return auth.ApiKey{}, err
	}

	if k.Claims, err = decodeClaims(f["claims"]); err != nil {
		return auth.ApiKey{}, err
	}

	if k.ExpiresAt, err = parseMs(f["expires_at"]); err != nil {
		return auth.ApiKey{}, err
	}

	if k.LastUsedAt, err = parseMs(f["last_used_at"]); err != nil {
		return auth.ApiKey{}, err
	}

	if k.CreatedAt, err = parseMs(f["created_at"]); err != nil {
		return auth.ApiKey{}, err
	}

	return k, nil
}

var _ auth.ApiKeyStore = ApiKeyStore{}

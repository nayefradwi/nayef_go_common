package authredis

import (
	"context"
	"strconv"

	"github.com/google/uuid"
	"github.com/nayefradwi/nayef_go_common/auth"
	. "github.com/nayefradwi/nayef_go_common/errors"
	"github.com/redis/go-redis/v9"
)

var saveTotpScript = redis.NewScript(`
if redis.call('HEXISTS', KEYS[1], 'confirmed_at') == 1 then
	return 0
end
redis.call('HSET', KEYS[1], 'secret', ARGV[1], 'last_step', 0)
return 1
`)

var confirmTotpScript = redis.NewScript(luaNow + `
local v = redis.call('HMGET', KEYS[1], 'last_step', 'confirmed_at')
if not v[1] or v[2] or tonumber(v[1]) >= tonumber(ARGV[1]) then
	return 0
end
redis.call('HSET', KEYS[1], 'confirmed_at', string.format('%d', nowMs()), 'last_step', ARGV[1])
redis.call('DEL', KEYS[2])
if #ARGV > 1 then
	redis.call('SADD', KEYS[2], unpack(ARGV, 2))
end
return 1
`)

var useStepScript = redis.NewScript(`
local v = redis.call('HMGET', KEYS[1], 'last_step', 'confirmed_at')
if not v[2] or tonumber(v[1]) >= tonumber(ARGV[1]) then
	return 0
end
redis.call('HSET', KEYS[1], 'last_step', ARGV[1])
return 1
`)

var setRecoveryCodesScript = redis.NewScript(`
if redis.call('HEXISTS', KEYS[1], 'confirmed_at') == 0 then
	return 0
end
redis.call('DEL', KEYS[2])
if #ARGV > 0 then
	redis.call('SADD', KEYS[2], unpack(ARGV))
end
return 1
`)

// one hash plus a recovery code set per owner; Delete drops both in one DEL
type TotpStore struct {
	client *redis.Client
	prefix string
}

type TotpStoreConfig struct {
	Prefix string
}

const defaultTotpPrefix = "auth:totp:"

var DefaultTotpStoreConfig = TotpStoreConfig{
	Prefix: defaultTotpPrefix,
}

func NewTotpStore(client *redis.Client, configs ...TotpStoreConfig) (TotpStore, error) {
	config := DefaultTotpStoreConfig
	if len(configs) > 0 {
		config = configs[0]
	}

	if err := validPrefix(config.Prefix); err != nil {
		return TotpStore{}, err
	}

	return TotpStore{client: client, prefix: config.Prefix}, nil
}

func (s TotpStore) Save(ctx context.Context, ownerId uuid.UUID, secret []byte) error {
	saved, err := saveTotpScript.Run(ctx, s.client, s.keys(ownerId)[:1], secret).Int()
	if err != nil {
		return InternalError("failed to save totp: " + err.Error())
	}

	if saved == 0 {
		return auth.ErrTotpEnabled
	}

	return nil
}

func (s TotpStore) Get(ctx context.Context, ownerId uuid.UUID) (auth.Totp, error) {
	keys := s.keys(ownerId)
	var fields *redis.MapStringStringCmd
	var codes *redis.StringSliceCmd
	_, err := s.client.TxPipelined(ctx, func(p redis.Pipeliner) error {
		fields = p.HGetAll(ctx, keys[0])
		codes = p.SMembers(ctx, keys[1])
		return nil
	})
	if err != nil {
		return auth.Totp{}, InternalError("failed to get totp: " + err.Error())
	}

	f := fields.Val()
	if len(f) == 0 {
		return auth.Totp{}, auth.ErrTotpNotFound
	}

	lastStep, err := strconv.ParseInt(f["last_step"], 10, 64)
	if err != nil {
		return auth.Totp{}, InternalError("failed to get totp: " + err.Error())
	}

	confirmedAt, err := parseMs(f["confirmed_at"])
	if err != nil {
		return auth.Totp{}, InternalError("failed to get totp: " + err.Error())
	}

	t := auth.Totp{OwnerId: ownerId, Secret: []byte(f["secret"]), LastStep: lastStep, ConfirmedAt: confirmedAt}
	for _, c := range codes.Val() {
		t.RecoveryCodes = append(t.RecoveryCodes, []byte(c))
	}

	return t, nil
}

func (s TotpStore) Confirm(ctx context.Context, ownerId uuid.UUID, step int64, recoveryCodes [][]byte) error {
	return s.update(ctx, "confirm totp", confirmTotpScript, ownerId, append([]any{step}, byteArgs(recoveryCodes)...)...)
}

func (s TotpStore) UseStep(ctx context.Context, ownerId uuid.UUID, step int64) error {
	return s.update(ctx, "use totp step", useStepScript, ownerId, step)
}

func (s TotpStore) SetRecoveryCodes(ctx context.Context, ownerId uuid.UUID, hashes [][]byte) error {
	return s.update(ctx, "set recovery codes", setRecoveryCodesScript, ownerId, byteArgs(hashes)...)
}

// only Confirm and SetRecoveryCodes fill the set, so a code here means the totp is confirmed
func (s TotpStore) UseRecoveryCode(ctx context.Context, ownerId uuid.UUID, hash []byte) error {
	removed, err := s.client.SRem(ctx, s.keys(ownerId)[1], hash).Result()
	if err != nil {
		return InternalError("failed to use recovery code: " + err.Error())
	}

	if removed == 0 {
		return auth.ErrTotpNotFound
	}

	return nil
}

func (s TotpStore) Delete(ctx context.Context, ownerId uuid.UUID) error {
	if err := s.client.Del(ctx, s.keys(ownerId)...).Err(); err != nil {
		return InternalError("failed to delete totp: " + err.Error())
	}
	return nil
}

func (s TotpStore) update(ctx context.Context, action string, script *redis.Script, ownerId uuid.UUID, args ...any) error {
	updated, err := script.Run(ctx, s.client, s.keys(ownerId), args...).Int()
	if err != nil {
		return InternalError("failed to " + action + ": " + err.Error())
	}

	if updated == 0 {
		return auth.ErrTotpNotFound
	}

	return nil
}

func (s TotpStore) keys(ownerId uuid.UUID) []string {
	key := s.prefix + ownerId.String()
	return []string{key, key + ":recovery"}
}

var _ auth.TotpStore = TotpStore{}

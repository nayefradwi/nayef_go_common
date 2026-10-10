package authredis

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/nayefradwi/nayef_go_common/auth"
	. "github.com/nayefradwi/nayef_go_common/errors"
	"github.com/redis/go-redis/v9"
)

// keys: r:<refresh> session, a:<access> → refresh key, o:<owner> family ids, f:<family> marker;
// a session counts only while its family marker exists, so revoking is one DEL
const luaWriteSession = luaExtend + `
local function writeSession(key, accessKey, refreshHash, family, owner, accessHash, accessExp, claims, exp)
	redis.call('HSET', key,
		'refresh_hash', refreshHash, 'family_id', family, 'owner_id', owner,
		'access_hash', accessHash, 'access_expires_at', accessExp,
		'claims', claims, 'expires_at', exp, 'rotated_at', '')
	redis.call('PEXPIREAT', key, exp)
	if accessKey then
		redis.call('SET', accessKey, key, 'PXAT', accessExp)
	end
end
`

const luaReadSession = `
local function readSession(key, prefix)
	local v = redis.call('HMGET', key,
		'refresh_hash', 'family_id', 'owner_id', 'access_hash', 'access_expires_at',
		'claims', 'expires_at', 'rotated_at')
	if not v[1] or redis.call('EXISTS', prefix .. 'f:' .. v[2]) == 0 then
		return false
	end
	return v
end
`

// KEYS: refresh, marker, owner set, [access]
// ARGV: refresh hash, family, owner, access hash, access expiry, claims, expiry, prefix
var createSessionScript = redis.NewScript(luaWriteSession + `
writeSession(KEYS[1], KEYS[4], ARGV[1], ARGV[2], ARGV[3], ARGV[4], ARGV[5], ARGV[6], ARGV[7])
redis.call('SET', KEYS[2], '1', 'PXAT', ARGV[7])
for _, f in ipairs(redis.call('SMEMBERS', KEYS[3])) do
	if redis.call('EXISTS', ARGV[8] .. 'f:' .. f) == 0 then
		redis.call('SREM', KEYS[3], f)
	end
end
redis.call('SADD', KEYS[3], ARGV[2])
extend(KEYS[3], ARGV[7])
return 1
`)

// KEYS: old refresh, new refresh, [new access]
// ARGV: new refresh hash, new access hash, new access expiry, new expiry, prefix
var rotateSessionScript = redis.NewScript(luaNow + luaWriteSession + `
local v = redis.call('HMGET', KEYS[1], 'family_id', 'owner_id', 'claims', 'rotated_at')
if not v[1] then
	return false
end
local marker = ARGV[5] .. 'f:' .. v[1]
if v[4] ~= '' or redis.call('EXISTS', marker) == 0 then
	return false
end
redis.call('HSET', KEYS[1], 'rotated_at', string.format('%d', nowMs()))
writeSession(KEYS[2], KEYS[3], ARGV[1], v[1], v[2], ARGV[2], ARGV[3], v[3], ARGV[4])
extend(marker, ARGV[4])
extend(ARGV[5] .. 'o:' .. v[2], ARGV[4])
return {v[1], v[2], v[3]}
`)

var getByRefreshScript = redis.NewScript(luaReadSession + `
return readSession(KEYS[1], ARGV[1])
`)

var getByAccessScript = redis.NewScript(luaReadSession + `
local key = redis.call('GET', KEYS[1])
if not key then
	return false
end
return readSession(key, ARGV[1])
`)

var deleteOwnerScript = redis.NewScript(`
for _, f in ipairs(redis.call('SMEMBERS', KEYS[1])) do
	redis.call('DEL', ARGV[1] .. 'f:' .. f)
end
return redis.call('DEL', KEYS[1])
`)

type SessionStore struct {
	client *redis.Client
	prefix string
}

type SessionStoreConfig struct {
	Prefix string
}

const defaultSessionPrefix = "auth:sessions:"

var DefaultSessionStoreConfig = SessionStoreConfig{
	Prefix: defaultSessionPrefix,
}

func NewSessionStore(client *redis.Client, configs ...SessionStoreConfig) (SessionStore, error) {
	config := DefaultSessionStoreConfig
	if len(configs) > 0 {
		config = configs[0]
	}

	if err := validPrefix(config.Prefix); err != nil {
		return SessionStore{}, err
	}

	return SessionStore{client: client, prefix: config.Prefix}, nil
}

func (s SessionStore) Create(ctx context.Context, session auth.Session) error {
	claims, err := encodeClaims(session.Claims)
	if err != nil {
		return InternalError("failed to create session: " + err.Error())
	}

	keys := []string{
		s.refreshKey(session.RefreshHash),
		s.prefix + "f:" + session.FamilyId.String(),
		s.prefix + "o:" + session.OwnerId.String(),
	}
	keys = s.withAccessKey(keys, session.AccessHash)

	err = createSessionScript.Run(ctx, s.client, keys,
		session.RefreshHash, session.FamilyId.String(), session.OwnerId.String(),
		session.AccessHash, formatMs(session.AccessExpiresAt), claims, formatMs(session.ExpiresAt), s.prefix,
	).Err()
	if err != nil {
		return InternalError("failed to create session: " + err.Error())
	}

	return nil
}

func (s SessionStore) Rotate(ctx context.Context, refreshHash []byte, next auth.Session) (auth.Session, error) {
	keys := s.withAccessKey([]string{s.refreshKey(refreshHash), s.refreshKey(next.RefreshHash)}, next.AccessHash)
	v, err := rotateSessionScript.Run(ctx, s.client, keys,
		next.RefreshHash, next.AccessHash, formatMs(next.AccessExpiresAt), formatMs(next.ExpiresAt), s.prefix,
	).StringSlice()
	if errors.Is(err, redis.Nil) {
		return auth.Session{}, auth.ErrSessionNotFound
	}

	if err != nil {
		return auth.Session{}, InternalError("failed to rotate session: " + err.Error())
	}

	if next.FamilyId, err = uuid.Parse(v[0]); err != nil {
		return auth.Session{}, InternalError("failed to rotate session: " + err.Error())
	}

	if next.OwnerId, err = uuid.Parse(v[1]); err != nil {
		return auth.Session{}, InternalError("failed to rotate session: " + err.Error())
	}

	if next.Claims, err = decodeClaims(v[2]); err != nil {
		return auth.Session{}, InternalError("failed to rotate session: " + err.Error())
	}

	return next, nil
}

func (s SessionStore) GetByRefresh(ctx context.Context, refreshHash []byte) (auth.Session, error) {
	return s.read(getByRefreshScript.RunRO(ctx, s.client, []string{s.refreshKey(refreshHash)}, s.prefix))
}

func (s SessionStore) GetByAccess(ctx context.Context, accessHash []byte) (auth.Session, error) {
	return s.read(getByAccessScript.RunRO(ctx, s.client, []string{s.prefix + "a:" + string(accessHash)}, s.prefix))
}

// orphaned refresh and access keys stop resolving once the marker is gone and expire on their own
func (s SessionStore) DeleteFamily(ctx context.Context, familyId uuid.UUID) error {
	if err := s.client.Del(ctx, s.prefix+"f:"+familyId.String()).Err(); err != nil {
		return InternalError("failed to delete sessions: " + err.Error())
	}
	return nil
}

func (s SessionStore) DeleteOwner(ctx context.Context, ownerId uuid.UUID) error {
	err := deleteOwnerScript.Run(ctx, s.client, []string{s.prefix + "o:" + ownerId.String()}, s.prefix).Err()
	if err != nil {
		return InternalError("failed to delete sessions: " + err.Error())
	}
	return nil
}

// keys expire on their own
func (s SessionStore) DeleteExpired(ctx context.Context) error {
	return nil
}

func (s SessionStore) read(cmd *redis.Cmd) (auth.Session, error) {
	v, err := cmd.StringSlice()
	if errors.Is(err, redis.Nil) {
		return auth.Session{}, auth.ErrSessionNotFound
	}

	if err != nil {
		return auth.Session{}, InternalError("failed to read session: " + err.Error())
	}

	session, err := parseSession(v)
	if err != nil {
		return auth.Session{}, InternalError("failed to read session: " + err.Error())
	}

	return session, nil
}

func (s SessionStore) refreshKey(refreshHash []byte) string {
	return s.prefix + "r:" + string(refreshHash)
}

// jwt sessions have no access hash
func (s SessionStore) withAccessKey(keys []string, accessHash []byte) []string {
	if len(accessHash) == 0 {
		return keys
	}
	return append(keys, s.prefix+"a:"+string(accessHash))
}

func parseSession(v []string) (auth.Session, error) {
	session := auth.Session{RefreshHash: []byte(v[0])}
	var err error
	if session.FamilyId, err = uuid.Parse(v[1]); err != nil {
		return auth.Session{}, err
	}

	if session.OwnerId, err = uuid.Parse(v[2]); err != nil {
		return auth.Session{}, err
	}

	if v[3] != "" {
		session.AccessHash = []byte(v[3])
	}

	if session.AccessExpiresAt, err = parseMs(v[4]); err != nil {
		return auth.Session{}, err
	}

	if session.Claims, err = decodeClaims(v[5]); err != nil {
		return auth.Session{}, err
	}

	if session.ExpiresAt, err = parseMs(v[6]); err != nil {
		return auth.Session{}, err
	}

	if session.RotatedAt, err = parseMs(v[7]); err != nil {
		return auth.Session{}, err
	}

	return session, nil
}

func encodeClaims(claims map[string]any) (string, error) {
	if claims == nil {
		return "", nil
	}

	b, err := json.Marshal(claims)
	return string(b), err
}

func decodeClaims(s string) (map[string]any, error) {
	if s == "" {
		return nil, nil
	}

	var claims map[string]any
	err := json.Unmarshal([]byte(s), &claims)
	return claims, err
}

var _ auth.SessionStore = SessionStore{}

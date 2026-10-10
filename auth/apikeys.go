package auth

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	. "github.com/nayefradwi/nayef_go_common/errors"
)

var ErrApiKeyNotFound = errors.New("api key not found")

type ApiKey struct {
	Id         uuid.UUID
	OwnerId    uuid.UUID
	Name       string
	Hash       []byte
	Claims     map[string]any
	ExpiresAt  time.Time
	LastUsedAt time.Time
	CreatedAt  time.Time
}

func (k ApiKey) IsExpired() bool {
	return !k.ExpiresAt.IsZero() && !time.Now().Before(k.ExpiresAt)
}

type ApiKeyManager struct {
	store  ApiKeyStore
	prefix string
}

var apiKeyPrefix = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

const (
	maxApiKeyPrefixLen = 16
	apiKeyTouchEvery   = time.Minute
)

func NewApiKeyManager(store ApiKeyStore, prefix string) (ApiKeyManager, error) {
	if store == nil {
		return ApiKeyManager{}, BadRequestError("api key store must not be nil")
	}

	if len(prefix) > maxApiKeyPrefixLen || !apiKeyPrefix.MatchString(prefix) {
		return ApiKeyManager{}, BadRequestError("api key prefix must be lowercase letters, digits and underscores, at most 16 bytes")
	}

	return ApiKeyManager{store: store, prefix: prefix + "_"}, nil
}

func (m ApiKeyManager) Issue(ctx context.Context, ownerId uuid.UUID, name string, claims map[string]any, expiresAt time.Time) (string, ApiKey, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", ApiKey{}, InternalError("failed to generate api key id: " + err.Error())
	}

	token, _, err := NewOpaqueToken()
	if err != nil {
		return "", ApiKey{}, err
	}

	key := m.prefix + token
	k := ApiKey{
		Id:        id,
		OwnerId:   ownerId,
		Name:      name,
		Hash:      HashOpaqueToken(key),
		Claims:    claims,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now().UTC(),
	}

	if err := m.store.Create(ctx, k); err != nil {
		return "", ApiKey{}, err
	}

	return key, k, nil
}

func (m ApiKeyManager) VerifyKey(ctx context.Context, key string) (Identity, error) {
	if !strings.HasPrefix(key, m.prefix) {
		return Identity{}, UnauthorizedError("Invalid key")
	}

	k, err := m.store.GetByHash(ctx, HashOpaqueToken(key))
	if errors.Is(err, ErrApiKeyNotFound) {
		return Identity{}, UnauthorizedError("Invalid key")
	}

	if err != nil {
		return Identity{}, err
	}

	if k.IsExpired() {
		return Identity{}, UnauthorizedError("Invalid key")
	}

	if time.Since(k.LastUsedAt) > apiKeyTouchEvery {
		if err := m.store.Touch(ctx, k.Id); err != nil {
			return Identity{}, err
		}
	}

	return Identity{OwnerId: k.OwnerId, Kind: KindKey, Claims: k.Claims}, nil
}

func (m ApiKeyManager) List(ctx context.Context, ownerId uuid.UUID) ([]ApiKey, error) {
	return m.store.ListByOwner(ctx, ownerId)
}

func (m ApiKeyManager) Revoke(ctx context.Context, ownerId, id uuid.UUID) error {
	err := m.store.Delete(ctx, ownerId, id)
	if errors.Is(err, ErrApiKeyNotFound) {
		return NotFoundError("api key not found")
	}

	return err
}

func (m ApiKeyManager) DeleteExpired(ctx context.Context) error {
	return m.store.DeleteExpired(ctx)
}

var _ KeyVerifier = ApiKeyManager{}

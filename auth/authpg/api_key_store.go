package authpg

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nayefradwi/nayef_go_common/auth"
	. "github.com/nayefradwi/nayef_go_common/errors"
)

const apiKeyColumns = `id, owner_id, name, hash, claims, expires_at, last_used_at, created_at`

type ApiKeyStore struct {
	db    dbExecutor
	table string
}

type ApiKeyStoreConfig struct {
	Table string
}

const defaultApiKeyTable = "auth_api_keys"

var DefaultApiKeyStoreConfig = ApiKeyStoreConfig{
	Table: defaultApiKeyTable,
}

func NewApiKeyStore(pool *pgxpool.Pool, configs ...ApiKeyStoreConfig) (ApiKeyStore, error) {
	config := DefaultApiKeyStoreConfig
	if len(configs) > 0 {
		config = configs[0]
	}

	if err := validTable(config.Table); err != nil {
		return ApiKeyStore{}, err
	}

	return ApiKeyStore{db: pool, table: config.Table}, nil
}

func (s ApiKeyStore) WithTx(tx pgx.Tx) ApiKeyStore {
	s.db = tx
	return s
}

func (s ApiKeyStore) Create(ctx context.Context, k auth.ApiKey) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO `+s.table+` (`+apiKeyColumns+`) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		k.Id, k.OwnerId, k.Name, k.Hash, k.Claims, nullTime(k.ExpiresAt), nullTime(k.LastUsedAt), k.CreatedAt,
	)
	if err != nil {
		return InternalError("failed to create api key: " + err.Error())
	}
	return nil
}

func (s ApiKeyStore) GetByHash(ctx context.Context, hash []byte) (auth.ApiKey, error) {
	row := s.db.QueryRow(ctx, `SELECT `+apiKeyColumns+` FROM `+s.table+` WHERE hash = $1`, hash)
	k, err := scanApiKey(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.ApiKey{}, auth.ErrApiKeyNotFound
	}

	if err != nil {
		return auth.ApiKey{}, InternalError("failed to read api key: " + err.Error())
	}

	return k, nil
}

func (s ApiKeyStore) ListByOwner(ctx context.Context, ownerId uuid.UUID) ([]auth.ApiKey, error) {
	rows, err := s.db.Query(ctx, `SELECT `+apiKeyColumns+` FROM `+s.table+` WHERE owner_id = $1 ORDER BY created_at DESC`, ownerId)
	if err != nil {
		return nil, InternalError("failed to list api keys: " + err.Error())
	}

	keys, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (auth.ApiKey, error) {
		return scanApiKey(row)
	})
	if err != nil {
		return nil, InternalError("failed to list api keys: " + err.Error())
	}

	return keys, nil
}

func (s ApiKeyStore) Touch(ctx context.Context, id uuid.UUID) error {
	if _, err := s.db.Exec(ctx, `UPDATE `+s.table+` SET last_used_at = now() WHERE id = $1`, id); err != nil {
		return InternalError("failed to touch api key: " + err.Error())
	}
	return nil
}

func (s ApiKeyStore) Delete(ctx context.Context, ownerId, id uuid.UUID) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM `+s.table+` WHERE owner_id = $1 AND id = $2`, ownerId, id)
	if err != nil {
		return InternalError("failed to delete api key: " + err.Error())
	}

	if tag.RowsAffected() == 0 {
		return auth.ErrApiKeyNotFound
	}

	return nil
}

func (s ApiKeyStore) DeleteExpired(ctx context.Context) error {
	if _, err := s.db.Exec(ctx, `DELETE FROM `+s.table+` WHERE expires_at <= now()`); err != nil {
		return InternalError("failed to delete expired api keys: " + err.Error())
	}
	return nil
}

func scanApiKey(row pgx.Row) (auth.ApiKey, error) {
	var k auth.ApiKey
	var expiresAt, lastUsedAt *time.Time
	if err := row.Scan(&k.Id, &k.OwnerId, &k.Name, &k.Hash, &k.Claims, &expiresAt, &lastUsedAt, &k.CreatedAt); err != nil {
		return auth.ApiKey{}, err
	}

	if expiresAt != nil {
		k.ExpiresAt = *expiresAt
	}

	if lastUsedAt != nil {
		k.LastUsedAt = *lastUsedAt
	}

	return k, nil
}

var _ auth.ApiKeyStore = ApiKeyStore{}

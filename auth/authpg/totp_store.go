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

const totpColumns = `owner_id, secret, last_step, confirmed_at, recovery_codes`

// pool only, no WithTx: a caller's rollback must not undo a used step
type TotpStore struct {
	pool  *pgxpool.Pool
	table string
}

type TotpStoreConfig struct {
	Table string
}

const defaultTotpTable = "auth_totp"

var DefaultTotpStoreConfig = TotpStoreConfig{
	Table: defaultTotpTable,
}

func NewTotpStore(pool *pgxpool.Pool, configs ...TotpStoreConfig) (TotpStore, error) {
	config := DefaultTotpStoreConfig
	if len(configs) > 0 {
		config = configs[0]
	}

	if err := validTable(config.Table); err != nil {
		return TotpStore{}, err
	}

	return TotpStore{pool: pool, table: config.Table}, nil
}

func (s TotpStore) Save(ctx context.Context, ownerId uuid.UUID, secret []byte) error {
	var saved uuid.UUID
	err := s.pool.QueryRow(ctx, `
		INSERT INTO `+s.table+` AS t (`+totpColumns+`) VALUES ($1, $2, 0, NULL, '{}')
		ON CONFLICT (owner_id) DO UPDATE SET secret = EXCLUDED.secret, last_step = 0
		WHERE t.confirmed_at IS NULL
		RETURNING owner_id`, ownerId, secret,
	).Scan(&saved)

	if errors.Is(err, pgx.ErrNoRows) {
		return auth.ErrTotpEnabled
	}

	if err != nil {
		return InternalError("failed to save totp: " + err.Error())
	}

	return nil
}

func (s TotpStore) Get(ctx context.Context, ownerId uuid.UUID) (auth.Totp, error) {
	var t auth.Totp
	var confirmedAt *time.Time
	err := s.pool.QueryRow(ctx, `SELECT `+totpColumns+` FROM `+s.table+` WHERE owner_id = $1`, ownerId).
		Scan(&t.OwnerId, &t.Secret, &t.LastStep, &confirmedAt, &t.RecoveryCodes)

	if errors.Is(err, pgx.ErrNoRows) {
		return auth.Totp{}, auth.ErrTotpNotFound
	}

	if err != nil {
		return auth.Totp{}, InternalError("failed to get totp: " + err.Error())
	}

	if confirmedAt != nil {
		t.ConfirmedAt = *confirmedAt
	}

	return t, nil
}

func (s TotpStore) Confirm(ctx context.Context, ownerId uuid.UUID, step int64, recoveryCodes [][]byte) error {
	return s.update(ctx, "confirm totp", `
		UPDATE `+s.table+` SET confirmed_at = now(), last_step = $2, recovery_codes = COALESCE($3::bytea[], '{}')
		WHERE owner_id = $1 AND confirmed_at IS NULL AND last_step < $2`, ownerId, step, recoveryCodes)
}

func (s TotpStore) UseStep(ctx context.Context, ownerId uuid.UUID, step int64) error {
	return s.update(ctx, "use totp step", `
		UPDATE `+s.table+` SET last_step = $2
		WHERE owner_id = $1 AND confirmed_at IS NOT NULL AND last_step < $2`, ownerId, step)
}

func (s TotpStore) SetRecoveryCodes(ctx context.Context, ownerId uuid.UUID, hashes [][]byte) error {
	return s.update(ctx, "set recovery codes", `
		UPDATE `+s.table+` SET recovery_codes = COALESCE($2::bytea[], '{}')
		WHERE owner_id = $1 AND confirmed_at IS NOT NULL`, ownerId, hashes)
}

func (s TotpStore) UseRecoveryCode(ctx context.Context, ownerId uuid.UUID, hash []byte) error {
	return s.update(ctx, "use recovery code", `
		UPDATE `+s.table+` SET recovery_codes = array_remove(recovery_codes, $2)
		WHERE owner_id = $1 AND confirmed_at IS NOT NULL AND $2 = ANY(recovery_codes)`, ownerId, hash)
}

func (s TotpStore) Delete(ctx context.Context, ownerId uuid.UUID) error {
	if _, err := s.pool.Exec(ctx, `DELETE FROM `+s.table+` WHERE owner_id = $1`, ownerId); err != nil {
		return InternalError("failed to delete totp: " + err.Error())
	}
	return nil
}

func (s TotpStore) update(ctx context.Context, action, sql string, args ...any) error {
	tag, err := s.pool.Exec(ctx, sql, args...)
	if err != nil {
		return InternalError("failed to " + action + ": " + err.Error())
	}

	if tag.RowsAffected() == 0 {
		return auth.ErrTotpNotFound
	}

	return nil
}

var _ auth.TotpStore = TotpStore{}

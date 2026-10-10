package authpg

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nayefradwi/nayef_go_common/auth"
	. "github.com/nayefradwi/nayef_go_common/errors"
)

const otpColumns = `key, hash, attempts, sends, sent_at, expires_at`

// pool only, no WithTx: a caller's rollback must not undo a failed attempt
type OtpStore struct {
	pool  *pgxpool.Pool
	table string
}

type OtpStoreConfig struct {
	Table string
}

const defaultOtpTable = "auth_otps"

var DefaultOtpStoreConfig = OtpStoreConfig{
	Table: defaultOtpTable,
}

func NewOtpStore(pool *pgxpool.Pool, configs ...OtpStoreConfig) (OtpStore, error) {
	config := DefaultOtpStoreConfig
	if len(configs) > 0 {
		config = configs[0]
	}

	if err := validTable(config.Table); err != nil {
		return OtpStore{}, err
	}

	return OtpStore{pool: pool, table: config.Table}, nil
}

func (s OtpStore) Save(ctx context.Context, key string, hash []byte, l auth.OtpLimits) error {
	var saved string
	err := s.pool.QueryRow(ctx, `
		INSERT INTO `+s.table+` AS o (`+otpColumns+`) VALUES ($1, $2, 0, 1, now(), now() + $3)
		ON CONFLICT (key) DO UPDATE SET
			hash       = EXCLUDED.hash,
			sent_at    = now(),
			expires_at = EXCLUDED.expires_at,
			attempts   = CASE WHEN o.expires_at <= now() THEN 0 ELSE o.attempts END,
			sends      = CASE WHEN o.expires_at <= now() THEN 1 ELSE o.sends + 1 END
		WHERE o.expires_at <= now()
		   OR (o.sent_at + $4 <= now() AND o.sends < $5 AND o.attempts < $6)
		RETURNING key`, key, hash, l.TTL, l.ResendAfter, l.MaxSends, l.MaxAttempts,
	).Scan(&saved)

	if errors.Is(err, pgx.ErrNoRows) {
		return auth.ErrOtpResendBlocked
	}

	if err != nil {
		return InternalError("failed to save otp: " + err.Error())
	}

	return nil
}

func (s OtpStore) Attempt(ctx context.Context, key string, maxAttempts int) ([]byte, error) {
	var hash []byte
	err := s.pool.QueryRow(ctx, `
		UPDATE `+s.table+` SET attempts = attempts + 1
		WHERE key = $1 AND attempts < $2 AND expires_at > now()
		RETURNING hash`, key, maxAttempts,
	).Scan(&hash)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, auth.ErrOtpNotFound
	}

	if err != nil {
		return nil, InternalError("failed to attempt otp: " + err.Error())
	}

	return hash, nil
}

func (s OtpStore) Consume(ctx context.Context, key string, hash []byte) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM `+s.table+` WHERE key = $1 AND hash = $2`, key, hash)
	if err != nil {
		return InternalError("failed to consume otp: " + err.Error())
	}

	if tag.RowsAffected() == 0 {
		return auth.ErrOtpNotFound
	}

	return nil
}

func (s OtpStore) DeleteExpired(ctx context.Context) error {
	if _, err := s.pool.Exec(ctx, `DELETE FROM `+s.table+` WHERE expires_at <= now()`); err != nil {
		return InternalError("failed to delete expired otps: " + err.Error())
	}
	return nil
}

var _ auth.OtpStore = OtpStore{}

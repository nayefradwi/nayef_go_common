package authpg

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nayefradwi/nayef_go_common/auth"
	. "github.com/nayefradwi/nayef_go_common/errors"
)

type dbExecutor interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

const sessionColumns = `family_id, owner_id, refresh_hash, access_hash, access_expires_at, claims, expires_at, rotated_at`

type SessionStore struct {
	db    dbExecutor
	table string
}

type SessionStoreConfig struct {
	Table string
}

const defaultSessionTable = "auth_sessions"

var DefaultSessionStoreConfig = SessionStoreConfig{
	Table: defaultSessionTable,
}

func NewSessionStore(pool *pgxpool.Pool, configs ...SessionStoreConfig) (SessionStore, error) {
	config := DefaultSessionStoreConfig
	if len(configs) > 0 {
		config = configs[0]
	}

	if err := validTable(config.Table); err != nil {
		return SessionStore{}, err
	}

	return SessionStore{db: pool, table: config.Table}, nil
}

func (s SessionStore) WithTx(tx pgx.Tx) SessionStore {
	s.db = tx
	return s
}

func (s SessionStore) Create(ctx context.Context, session auth.Session) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO `+s.table+` (`+sessionColumns+`) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		session.FamilyId, session.OwnerId, session.RefreshHash, session.AccessHash,
		nullTime(session.AccessExpiresAt), session.Claims, session.ExpiresAt, nullTime(session.RotatedAt),
	)
	if err != nil {
		return InternalError("failed to create session: " + err.Error())
	}
	return nil
}

func (s SessionStore) Rotate(ctx context.Context, refreshHash []byte, next auth.Session) (auth.Session, error) {
	row := s.db.QueryRow(ctx, `
		WITH used AS (
			UPDATE `+s.table+` SET rotated_at = now()
			WHERE refresh_hash = $1 AND rotated_at IS NULL AND expires_at > now()
			RETURNING family_id, owner_id, claims
		)
		INSERT INTO `+s.table+` (`+sessionColumns+`)
		SELECT family_id, owner_id, $2, $3, $4, claims, $5, NULL FROM used
		RETURNING `+sessionColumns,
		refreshHash, next.RefreshHash, next.AccessHash, nullTime(next.AccessExpiresAt), next.ExpiresAt,
	)
	return scanSession(row, "rotate")
}

func (s SessionStore) GetByRefresh(ctx context.Context, refreshHash []byte) (auth.Session, error) {
	row := s.db.QueryRow(ctx, `SELECT `+sessionColumns+` FROM `+s.table+` WHERE refresh_hash = $1`, refreshHash)
	return scanSession(row, "read")
}

func (s SessionStore) GetByAccess(ctx context.Context, accessHash []byte) (auth.Session, error) {
	row := s.db.QueryRow(ctx, `SELECT `+sessionColumns+` FROM `+s.table+` WHERE access_hash = $1`, accessHash)
	return scanSession(row, "read")
}

func (s SessionStore) DeleteFamily(ctx context.Context, familyId uuid.UUID) error {
	return s.delete(ctx, `DELETE FROM `+s.table+` WHERE family_id = $1`, familyId)
}

func (s SessionStore) DeleteOwner(ctx context.Context, ownerId uuid.UUID) error {
	return s.delete(ctx, `DELETE FROM `+s.table+` WHERE owner_id = $1`, ownerId)
}

func (s SessionStore) DeleteExpired(ctx context.Context) error {
	return s.delete(ctx, `DELETE FROM `+s.table+` WHERE expires_at <= now()`)
}

func (s SessionStore) delete(ctx context.Context, sql string, args ...any) error {
	if _, err := s.db.Exec(ctx, sql, args...); err != nil {
		return InternalError("failed to delete sessions: " + err.Error())
	}
	return nil
}

func scanSession(row pgx.Row, op string) (auth.Session, error) {
	var session auth.Session
	var accessExpiresAt, rotatedAt *time.Time
	err := row.Scan(
		&session.FamilyId, &session.OwnerId, &session.RefreshHash, &session.AccessHash,
		&accessExpiresAt, &session.Claims, &session.ExpiresAt, &rotatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return auth.Session{}, auth.ErrSessionNotFound
	}

	if err != nil {
		return auth.Session{}, InternalError("failed to " + op + " session: " + err.Error())
	}

	if accessExpiresAt != nil {
		session.AccessExpiresAt = *accessExpiresAt
	}

	if rotatedAt != nil {
		session.RotatedAt = *rotatedAt
	}

	return session, nil
}

var tableName = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

const maxTableLen = 48

func validTable(name string) error {
	if len(name) > maxTableLen || !tableName.MatchString(name) {
		return BadRequestError("table name must be lowercase letters, digits and underscores, at most 48 bytes")
	}

	return nil
}

func nullTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

var _ auth.SessionStore = SessionStore{}

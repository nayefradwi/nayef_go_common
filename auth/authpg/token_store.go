package authpg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nayefradwi/nayef_go_common/auth"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	. "github.com/nayefradwi/nayef_go_common/errors"
)

type dbExecutor interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type TokenStore struct {
	db     dbExecutor
	config TokenStoreConfig
}

type TokenStoreConfig struct {
	TableName string
}

var DefaultTokenStoreConfig = TokenStoreConfig{
	TableName: "tokens",
}

func NewTokenStore(pool *pgxpool.Pool, configs ...TokenStoreConfig) TokenStore {
	config := DefaultTokenStoreConfig
	if len(configs) > 0 {
		config = configs[0]
	}

	return TokenStore{db: pool, config: config}
}

func (s TokenStore) WithTx(tx pgx.Tx) TokenStore {
	s.db = tx
	return s
}

func (s TokenStore) StoreToken(ctx context.Context, token auth.Token) error {
	return s.StoreTokens(ctx, token)
}

func (s TokenStore) StoreTokens(ctx context.Context, tokens ...auth.Token) error {
	if len(tokens) == 0 {
		return nil
	}

	const cols = 7
	rows := make([]string, len(tokens))
	args := make([]any, 0, len(tokens)*cols)
	for i, token := range tokens {
		claimsJSON, err := json.Marshal(token.Claims)
		if err != nil {
			return InternalError("failed to marshal claims: " + err.Error())
		}
		n := i * cols
		// let postgres handle the uuid parsing of the text args
		rows[i] = fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d)", n+1, n+2, n+3, n+4, n+5, n+6, n+7)
		args = append(args, token.Id.String(), token.Value, token.OwnerId.String(), token.ExpiresAt, token.IssuedAt, claimsJSON, token.Type)
	}

	sql := `INSERT INTO ` + s.config.TableName + ` (id, value, owner_id, expires_at, issued_at, claims, type) VALUES ` + strings.Join(rows, ", ")
	if _, err := s.db.Exec(ctx, sql, args...); err != nil {
		return InternalError("failed to store tokens: " + err.Error())
	}

	return nil
}

func (s TokenStore) GetTokenByReference(ctx context.Context, reference uuid.UUID, tokenType int) (auth.Token, error) {
	row := s.db.QueryRow(ctx,
		`SELECT id::text, value, owner_id::text, expires_at, issued_at, claims, type
		 FROM `+s.config.TableName+` WHERE id = $1 AND type = $2`, reference.String(), tokenType,
	)
	return scanPgxToken(row)
}

func (s TokenStore) GetTokenByOwner(ctx context.Context, ownerId uuid.UUID, tokenType int) (auth.Token, error) {
	row := s.db.QueryRow(ctx,
		`SELECT id::text, value, owner_id::text, expires_at, issued_at, claims, type
		 FROM `+s.config.TableName+` WHERE owner_id = $1 AND type = $2`, ownerId.String(), tokenType,
	)
	return scanPgxToken(row)
}

func (s TokenStore) DeleteToken(ctx context.Context, reference uuid.UUID) error {
	_, err := s.db.Exec(ctx, `DELETE FROM `+s.config.TableName+` WHERE id = $1`, reference.String())
	if err != nil {
		return InternalError("failed to delete token: " + err.Error())
	}
	return nil
}

func (s TokenStore) DeleteAllTokensByOwner(ctx context.Context, ownerId uuid.UUID) error {
	_, err := s.db.Exec(ctx, `DELETE FROM `+s.config.TableName+` WHERE owner_id = $1`, ownerId.String())
	if err != nil {
		return InternalError("failed to delete tokens by owner: " + err.Error())
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanPgxToken(row rowScanner) (auth.Token, error) {
	var t auth.Token
	var idStr, ownerStr string
	var claimsJSON []byte
	err := row.Scan(&idStr, &t.Value, &ownerStr, &t.ExpiresAt, &t.IssuedAt, &claimsJSON, &t.Type)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.Token{}, UnauthorizedError("auth.Token not found")
	}
	if err != nil {
		slog.Error("failed to read token", "err", err)
		return auth.Token{}, InternalError("failed to read token")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return auth.Token{}, InternalError("invalid id in store: " + err.Error())
	}
	t.Id = id
	owner, err := uuid.Parse(ownerStr)
	if err != nil {
		return auth.Token{}, InternalError("invalid owner_id in store: " + err.Error())
	}
	t.OwnerId = owner
	if claimsJSON != nil {
		if err := json.Unmarshal(claimsJSON, &t.Claims); err != nil {
			return auth.Token{}, InternalError("failed to unmarshal claims: " + err.Error())
		}
	}
	return t, nil
}

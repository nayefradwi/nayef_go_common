package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	. "github.com/nayefradwi/nayef_go_common/errors"
)

type PostgresTokenStore struct {
	pool   *pgxpool.Pool
	config PostgresTokenStoreConfig
}

type PostgresTokenStoreConfig struct {
	TableName string
}

var DefaultPostgresTokenStoreConfig = PostgresTokenStoreConfig{
	TableName: "tokens",
}

func NewPostgresTokenStore(pool *pgxpool.Pool, configs ...PostgresTokenStoreConfig) ITokenStore {
	config := DefaultPostgresTokenStoreConfig
	if len(configs) > 0 {
		config = configs[0]
	}

	return PostgresTokenStore{pool: pool, config: config}
}

func (s PostgresTokenStore) StoreToken(token Token) error {
	return s.StoreTokens(token)
}

func (s PostgresTokenStore) StoreTokens(tokens ...Token) error {
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
	if _, err := s.pool.Exec(context.Background(), sql, args...); err != nil {
		return InternalError("failed to store tokens: " + err.Error())
	}

	return nil
}

func (s PostgresTokenStore) GetTokenByReference(reference uuid.UUID, tokenType int) (Token, error) {
	row := s.pool.QueryRow(context.Background(),
		`SELECT id::text, value, owner_id::text, expires_at, issued_at, claims, type
		 FROM `+s.config.TableName+` WHERE id = $1 AND type = $2`, reference.String(), tokenType,
	)
	return scanPgxToken(row)
}

func (s PostgresTokenStore) GetTokenByOwner(ownerId uuid.UUID, tokenType int) (Token, error) {
	row := s.pool.QueryRow(context.Background(),
		`SELECT id::text, value, owner_id::text, expires_at, issued_at, claims, type
		 FROM `+s.config.TableName+` WHERE owner_id = $1 AND type = $2`, ownerId.String(), tokenType,
	)
	return scanPgxToken(row)
}

func (s PostgresTokenStore) DeleteToken(reference uuid.UUID) error {
	_, err := s.pool.Exec(context.Background(), `DELETE FROM `+s.config.TableName+` WHERE id = $1`, reference.String())
	if err != nil {
		return InternalError("failed to delete token: " + err.Error())
	}
	return nil
}

func (s PostgresTokenStore) DeleteAllTokensByOwner(ownerId uuid.UUID) error {
	_, err := s.pool.Exec(context.Background(), `DELETE FROM `+s.config.TableName+` WHERE owner_id = $1`, ownerId.String())
	if err != nil {
		return InternalError("failed to delete tokens by owner: " + err.Error())
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanPgxToken(row rowScanner) (Token, error) {
	var t Token
	var idStr, ownerStr string
	var claimsJSON []byte
	if err := row.Scan(&idStr, &t.Value, &ownerStr, &t.ExpiresAt, &t.IssuedAt, &claimsJSON, &t.Type); err != nil {
		return Token{}, UnauthorizedError("Token not found")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return Token{}, InternalError("invalid id in store: " + err.Error())
	}
	t.Id = id
	owner, err := uuid.Parse(ownerStr)
	if err != nil {
		return Token{}, InternalError("invalid owner_id in store: " + err.Error())
	}
	t.OwnerId = owner
	if claimsJSON != nil {
		if err := json.Unmarshal(claimsJSON, &t.Claims); err != nil {
			return Token{}, InternalError("failed to unmarshal claims: " + err.Error())
		}
	}
	return t, nil
}

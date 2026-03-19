package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

type pgxTokenStore struct {
	conn *pgx.Conn
	ctx  context.Context
}

func (s *pgxTokenStore) StoreToken(token Token) error {
	claimsJSON, err := json.Marshal(token.Claims)
	if err != nil {
		return fmt.Errorf("marshal claims: %w", err)
	}
	_, err = s.conn.Exec(s.ctx,
		`INSERT INTO tokens (id, value, owner_id, expires_at, issued_at, claims, type)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		token.Id, token.Value, token.OwnerId, token.ExpiresAt, token.IssuedAt, claimsJSON, token.Type,
	)
	return err
}

func (s *pgxTokenStore) StoreTokens(tokens ...Token) error {
	for _, token := range tokens {
		if err := s.StoreToken(token); err != nil {
			return err
		}
	}
	return nil
}

func (s *pgxTokenStore) GetTokenByReference(reference string, tokenType int) (Token, error) {
	row := s.conn.QueryRow(s.ctx,
		`SELECT id, value, owner_id, expires_at, issued_at, claims, type
		 FROM tokens WHERE id = $1 AND type = $2`, reference, tokenType,
	)
	return scanToken(row)
}

func (s *pgxTokenStore) GetTokenByOwner(ownerId string, tokenType int) (Token, error) {
	row := s.conn.QueryRow(s.ctx,
		`SELECT id, value, owner_id, expires_at, issued_at, claims, type
		 FROM tokens WHERE owner_id = $1 AND type = $2`, ownerId, tokenType,
	)
	return scanToken(row)
}

func (s *pgxTokenStore) DeleteToken(reference string) error {
	_, err := s.conn.Exec(s.ctx, `DELETE FROM tokens WHERE id = $1`, reference)
	return err
}

func (s *pgxTokenStore) DeleteAllTokensByOwner(ownerId string) error {
	_, err := s.conn.Exec(s.ctx, `DELETE FROM tokens WHERE owner_id = $1`, ownerId)
	return err
}

func scanToken(row pgx.Row) (Token, error) {
	var t Token
	var claimsJSON []byte
	err := row.Scan(&t.Id, &t.Value, &t.OwnerId, &t.ExpiresAt, &t.IssuedAt, &claimsJSON, &t.Type)
	if err != nil {
		return Token{}, err
	}
	if claimsJSON != nil {
		if err := json.Unmarshal(claimsJSON, &t.Claims); err != nil {
			return Token{}, fmt.Errorf("unmarshal claims: %w", err)
		}
	}
	return t, nil
}

type testEnv struct {
	provider IReferenceTokenProvider
	store    ITokenStore
}

func setupTestEnv(t *testing.T) testEnv {
	t.Helper()

	conn := mustCreatePostgresConn(t)
	store := &pgxTokenStore{conn: conn, ctx: context.Background()}

	accessCfg, err := NewJwtTokenProviderConfig("test-access-secret-key", time.Hour)
	require.NoError(t, err)
	accessCfg = accessCfg.SetTokenType(AccessTokenType)

	refreshCfg, err := NewJwtTokenProviderConfig("test-refresh-secret-key", 24*time.Hour)
	require.NoError(t, err)
	refreshCfg = refreshCfg.SetTokenType(RefreshTokenType)

	accessProvider := NewJwtTokenProvider(accessCfg)
	refreshProvider := NewJwtTokenProvider(refreshCfg)
	refreshTokenProvider := NewJwtRefreshTokenProvider(refreshProvider, accessProvider)
	referenceProvider := NewJwtReferenceTokenProvider(refreshTokenProvider, store)

	return testEnv{
		provider: referenceProvider,
		store:    store,
	}
}

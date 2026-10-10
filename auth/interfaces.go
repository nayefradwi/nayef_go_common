package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ITokenProvider interface {
	GetClaims(token string) (Token, error)
	SignClaims(owner uuid.UUID, claims map[string]any) (string, error)
}

type IRefreshTokenProvider interface {
	GenerateToken(ctx context.Context, ownerId uuid.UUID, claims map[string]any) (TokenDTO, error)
	GetAccessToken(accessToken string) (Token, error)
	GetRefreshToken(refreshToken string) (Token, error)
	GetAccessTokenProvider() ITokenProvider
}

type AccessVerifier interface {
	VerifyAccess(ctx context.Context, token string) (Token, error)
}

type KeyVerifier interface {
	VerifyKey(ctx context.Context, key string) (Identity, error)
}

type SessionStore interface {
	Create(ctx context.Context, s Session) error
	Rotate(ctx context.Context, refreshHash []byte, next Session) (Session, error)
	GetByRefresh(ctx context.Context, refreshHash []byte) (Session, error)
	GetByAccess(ctx context.Context, accessHash []byte) (Session, error)
	DeleteFamily(ctx context.Context, familyId uuid.UUID) error
	DeleteOwner(ctx context.Context, ownerId uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

type AttemptStore interface {
	Hit(ctx context.Context, key string, window time.Duration) (Attempt, error)
	Reset(ctx context.Context, key string) error
	DeleteExpired(ctx context.Context) error
}

type CodeStore interface {
	Save(ctx context.Context, key string, hash []byte, limits CodeLimits) error
	Attempt(ctx context.Context, key string, maxAttempts int) ([]byte, error)
	Consume(ctx context.Context, key string, hash []byte) error // must reject expired rows
	DeleteExpired(ctx context.Context) error
}

type ApiKeyStore interface {
	Create(ctx context.Context, k ApiKey) error
	GetByHash(ctx context.Context, hash []byte) (ApiKey, error)
	ListByOwner(ctx context.Context, ownerId uuid.UUID) ([]ApiKey, error)
	Touch(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, ownerId, id uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

type TotpStore interface {
	Save(ctx context.Context, ownerId uuid.UUID, secret []byte) error
	Get(ctx context.Context, ownerId uuid.UUID) (Totp, error)
	Confirm(ctx context.Context, ownerId uuid.UUID, step int64, recoveryCodes [][]byte) error
	UseStep(ctx context.Context, ownerId uuid.UUID, step int64) error
	SetRecoveryCodes(ctx context.Context, ownerId uuid.UUID, hashes [][]byte) error
	UseRecoveryCode(ctx context.Context, ownerId uuid.UUID, hash []byte) error
	Delete(ctx context.Context, ownerId uuid.UUID) error
}

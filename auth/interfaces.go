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

type OtpStore interface {
	Save(ctx context.Context, key string, hash []byte, limits OtpLimits) error
	Attempt(ctx context.Context, key string, maxAttempts int) ([]byte, error)
	Consume(ctx context.Context, key string, hash []byte) error
	DeleteExpired(ctx context.Context) error
}

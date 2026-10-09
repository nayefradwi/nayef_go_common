package auth

import (
	"context"

	"github.com/google/uuid"
)

type ITokenProvider interface {
	GetClaims(token string) (Token, error)
	SignClaims(owner uuid.UUID, claims map[string]any) (string, error)
}

type ITokenStore interface {
	StoreToken(ctx context.Context, token Token) error
	StoreTokens(ctx context.Context, tokens ...Token) error
	GetTokenByReference(ctx context.Context, reference uuid.UUID, tokenType int) (Token, error)
	GetTokenByOwner(ctx context.Context, ownerId uuid.UUID, tokenType int) (Token, error)
	DeleteToken(ctx context.Context, reference uuid.UUID) error
	DeleteAllTokensByOwner(ctx context.Context, ownerId uuid.UUID) error
}

type IRefreshTokenProvider interface {
	GenerateToken(ctx context.Context, ownerId uuid.UUID, claims map[string]any) (TokenDTO, error)
	GetAccessToken(accessToken string) (Token, error)
	GetRefreshToken(refreshToken string) (Token, error)
	GetAccessTokenProvider() ITokenProvider
}

type IRefreshTokenProviderWithRevoke interface {
	IRefreshTokenProvider
	GenerateId() (uuid.UUID, error)
	RevokeToken(ctx context.Context, reference uuid.UUID) error
	RevokeOwner(ctx context.Context, ownerId uuid.UUID) error
}

type IReferenceTokenProvider interface {
	GenerateId() (uuid.UUID, error)
	GenerateToken(ctx context.Context, ownerId uuid.UUID, claims map[string]any) (TokenDTO, error)
	GetAccessToken(ctx context.Context, id uuid.UUID) (Token, error)
	GetRefreshToken(ctx context.Context, id uuid.UUID) (Token, error)
	RevokeToken(ctx context.Context, id uuid.UUID) error
	RevokeOwner(ctx context.Context, ownerId uuid.UUID) error
	GetAccessTokenProvider() ITokenProvider
}

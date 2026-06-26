package auth

import "github.com/google/uuid"

type ITokenProvider interface {
	GetClaims(token string) (Token, error)
	SignClaims(owner uuid.UUID, claims map[string]any) (string, error)
}

type ITokenStore interface {
	StoreToken(token Token) error
	StoreTokens(tokens ...Token) error
	GetTokenByReference(reference uuid.UUID, tokenType int) (Token, error)
	GetTokenByOwner(ownerId uuid.UUID, tokenType int) (Token, error)
	DeleteToken(reference uuid.UUID) error
	DeleteAllTokensByOwner(ownerId uuid.UUID) error
}

type IRefreshTokenProvider interface {
	GenerateToken(ownerId uuid.UUID, claims map[string]any) (TokenDTO, error)
	GetAccessToken(accessToken string) (Token, error)
	GetRefreshToken(refreshToken string) (Token, error)
	GetAccessTokenProvider() ITokenProvider
}

type IRefreshTokenProviderWithRevoke interface {
	IRefreshTokenProvider
	GenerateId() (uuid.UUID, error)
	RevokeToken(reference uuid.UUID) error
	RevokeOwner(ownerId uuid.UUID) error
}

type IReferenceTokenProvider interface {
	GenerateId() (uuid.UUID, error)
	GenerateToken(ownerId uuid.UUID, claims map[string]any) (TokenDTO, error)
	GetAccessToken(id uuid.UUID) (Token, error)
	GetRefreshToken(id uuid.UUID) (Token, error)
	RevokeToken(id uuid.UUID) error
	RevokeOwner(ownerId uuid.UUID) error
	GetAccessTokenProvider() ITokenProvider
}

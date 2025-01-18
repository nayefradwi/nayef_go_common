package auth

import (
	"github.com/nayefradwi/nayef_go_common/modules/auth"
)

type IRefreshTokenProvider interface {
	GenerateToken(ownerId string, claims map[string]interface{}) (*RefreshToken, error)
	GetAccessToken(accessToken string) (auth.Token, error)
	GetRefreshToken(refreshToken string) (auth.Token, error)
	GetAccessTokenProvider() auth.ITokenProvider
}

type IReferenceTokenRepository interface {
	StoreToken(accessToken auth.Token, refreshToken auth.Token) (ReferenceToken, error)
	GetTokenByReference(id string, tokenType int) (auth.Token, error)
	GetTokenByOwner(ownerId string, tokenType int) (auth.Token, error)
	DeleteToken(id string) error
	DeleteAllTokensByOwner(ownerId string) error
}

type IReferenceTokenProvider interface {
	GenerateId() string
	GenerateToken(ownerId string, claims map[string]interface{}) (ReferenceToken, error)
	GetAccessToken(id string) (auth.Token, error)
	GetRefreshToken(id string) (auth.Token, error)
	RevokeToken(id string) error
	RevokeOwner(ownerId string) error
	GetAccessTokenProvider() auth.ITokenProvider
}

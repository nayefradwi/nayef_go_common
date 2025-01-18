package auth

import (
	"github.com/nayefradwi/nayef_go_common/modules/auth"
)

type IRefreshTokenProvider[T string | int] interface {
	GenerateToken(ownerId T, claims map[string]interface{}) (*RefreshToken, error)
	GetAccessToken(accessToken string) (auth.Token[T], error)
	GetRefreshToken(refreshToken string) (auth.Token[T], error)
}

type IReferenceTokenRepository[T string | int] interface {
	StoreToken(accessToken auth.Token[T], refreshToken auth.Token[T]) (ReferenceToken, error)
	GetTokenByReference(id T, tokenType int) (auth.Token[T], error)
	GetTokenByOwner(ownerId T, tokenType int) (auth.Token[T], error)
	DeleteToken(id T) error
	DeleteAllTokensByOwner(ownerId T) error
}

type IReferenceTokenProvider[T string | int] interface {
	GenerateId() T
	GenerateToken(ownerId T, claims map[string]interface{}) (ReferenceToken, error)
	GetAccessToken(id string) (auth.Token[T], error)
	GetRefreshToken(id string) (auth.Token[T], error)
	RevokeToken(id T) error
	RevokeOwner(ownerId T) error
}

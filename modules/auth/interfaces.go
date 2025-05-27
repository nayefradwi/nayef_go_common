package auth

type ITokenProvider interface {
	GetClaims(token string) (Token, error)
	SignClaims(owner string, claims map[string]interface{}) (string, error)
}

type ITokenStore interface {
	StoreToken(token Token) error
	StoreTokens(tokens ...Token) error
	GetTokenByReference(reference string, tokenType int) (Token, error)
	GetTokenByOwner(ownerId string, tokenType int) (Token, error)
	DeleteToken(reference string) error
	DeleteAllTokensByOwner(ownerId string) error
}

type IRefreshTokenProvider interface {
	GenerateToken(ownerId string, claims map[string]interface{}) (TokenDTO, error)
	GetAccessToken(accessToken string) (Token, error)
	GetRefreshToken(refreshToken string) (Token, error)
	GetAccessTokenProvider() ITokenProvider
}

type IRefreshTokenProviderWithRevoke interface {
	IRefreshTokenProvider
	RevokeToken(reference string) error
	RevokeOwner(ownerId string) error
}

type IReferenceTokenProvider interface {
	GenerateId() string
	GenerateToken(ownerId string, claims map[string]interface{}) (TokenDTO, error)
	GetAccessToken(id string) (Token, error)
	GetRefreshToken(id string) (Token, error)
	RevokeToken(id string) error
	RevokeOwner(ownerId string) error
	GetAccessTokenProvider() ITokenProvider
}

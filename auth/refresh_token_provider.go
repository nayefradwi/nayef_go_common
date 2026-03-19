package auth

type JwtRefreshTokenProvider struct {
	RefreshTokenProvider JwtTokenProvider
	AccessTokenProvider  JwtTokenProvider
}

func NewJwtRefreshTokenProvider(refreshTokenProvider JwtTokenProvider, accessTokenProvider JwtTokenProvider) IRefreshTokenProvider {
	return JwtRefreshTokenProvider{
		RefreshTokenProvider: refreshTokenProvider,
		AccessTokenProvider:  accessTokenProvider,
	}
}

func (t JwtRefreshTokenProvider) GenerateToken(ownerId string, claims map[string]any) (TokenDTO, error) {
	accessToken, err := t.AccessTokenProvider.SignClaims(ownerId, claims)
	if err != nil {
		return EmptyTokenDTO(), err
	}

	refreshToken, err := t.RefreshTokenProvider.SignClaims(ownerId, claims)
	if err != nil {
		return EmptyTokenDTO(), err
	}

	return NewTokenDTOWithRefresh(accessToken, refreshToken), nil
}

func (t JwtRefreshTokenProvider) GetAccessToken(accessToken string) (Token, error) {
	return t.AccessTokenProvider.GetClaims(accessToken)
}

func (t JwtRefreshTokenProvider) GetRefreshToken(refreshToken string) (Token, error) {
	return t.RefreshTokenProvider.GetClaims(refreshToken)
}

func (t JwtRefreshTokenProvider) GetAccessTokenProvider() ITokenProvider {
	return t.AccessTokenProvider
}

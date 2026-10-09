package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	. "github.com/nayefradwi/nayef_go_common/errors"
)

type JwtReferenceTokenProvider struct {
	tokenProvider IRefreshTokenProvider
	tokenStore    ITokenStore
}

func NewJwtReferenceTokenProvider(tokenProvider IRefreshTokenProvider, tokenStore ITokenStore) IReferenceTokenProvider {
	return JwtReferenceTokenProvider{
		tokenProvider: tokenProvider,
		tokenStore:    tokenStore,
	}
}

func (t JwtReferenceTokenProvider) GenerateId() (uuid.UUID, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, InternalError("failed to generate token id: " + err.Error())
	}
	return id, nil
}

func (t JwtReferenceTokenProvider) GenerateToken(ctx context.Context, ownerId uuid.UUID, claims map[string]any) (TokenDTO, error) {
	tokenPair, err := t.tokenProvider.GenerateToken(ctx, ownerId, claims)
	if err != nil {
		return EmptyTokenDTO(), err
	}

	accessToken, err := t.tokenProvider.GetAccessToken(tokenPair.AccessToken)
	if err != nil {
		return EmptyTokenDTO(), err
	}

	refreshToken, err := t.tokenProvider.GetRefreshToken(tokenPair.RefreshToken)
	if err != nil {
		return EmptyTokenDTO(), err
	}

	accessTokenId, err := t.GenerateId()
	if err != nil {
		return EmptyTokenDTO(), err
	}

	refreshTokenId, err := t.GenerateId()
	if err != nil {
		return EmptyTokenDTO(), err
	}

	accessToken.Id, refreshToken.Id = accessTokenId, refreshTokenId
	if err := t.tokenStore.StoreTokens(ctx, accessToken, refreshToken); err != nil {
		return EmptyTokenDTO(), err
	}

	return NewTokenDTOWithRefresh(accessTokenId.String(), refreshTokenId.String()), nil
}

func (t JwtReferenceTokenProvider) GetAccessToken(ctx context.Context, id uuid.UUID) (Token, error) {
	return t.getToken(ctx, id, AccessTokenType)
}

func (t JwtReferenceTokenProvider) GetRefreshToken(ctx context.Context, id uuid.UUID) (Token, error) {
	return t.getToken(ctx, id, RefreshTokenType)
}

func (t JwtReferenceTokenProvider) getToken(ctx context.Context, id uuid.UUID, tokenType int) (Token, error) {
	token, err := t.tokenStore.GetTokenByReference(ctx, id, tokenType)
	if err != nil {
		return Token{}, err
	}
	if token.IsExpired() {
		return Token{}, UnauthorizedError("Token expired")
	}
	return token, nil
}

func (t JwtReferenceTokenProvider) RevokeToken(ctx context.Context, id uuid.UUID) error {
	return t.tokenStore.DeleteToken(ctx, id)
}

func (t JwtReferenceTokenProvider) RevokeOwner(ctx context.Context, ownerId uuid.UUID) error {
	return t.tokenStore.DeleteAllTokensByOwner(ctx, ownerId)
}

func (t JwtReferenceTokenProvider) GetAccessTokenProvider() ITokenProvider {
	return t.tokenProvider.GetAccessTokenProvider()
}

func (t JwtReferenceTokenProvider) WithTx(tx pgx.Tx) IReferenceTokenProvider {
	t.tokenStore = t.tokenStore.WithTx(tx)
	return t
}

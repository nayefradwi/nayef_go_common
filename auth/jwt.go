package auth

import (
	"maps"
	"time"

	"github.com/golang-jwt/jwt/v5"
	. "github.com/nayefradwi/nayef_go_common/errors"
)

type JwtTokenProvider struct {
	Config JwtTokenProviderConfig
}

func NewJwtTokenProvider(config JwtTokenProviderConfig) JwtTokenProvider {
	return JwtTokenProvider{
		Config: config,
	}
}

func (t JwtTokenProvider) GetClaims(token string) (Token, error) {
	jwtToken, err := jwt.Parse(token, t.Config.parser, t.Config.parserOpts...)
	if err != nil {
		return Token{}, UnauthorizedError(err.Error())
	}

	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return Token{}, UnauthorizedError("Invalid token")
	}

	owner, _ := claims[ownerClaimKey].(string)
	issuedAt, _ := claims[issuedAtClaimKey].(float64)
	expiresAt, _ := claims[expiryClaimKey].(float64)
	tokenType, _ := claims[tokenTypeClaimKey].(float64)

	if t.Config.TokenType != 0 && int(tokenType) != t.Config.TokenType {
		return Token{}, UnauthorizedError("invalid token type")
	}

	return Token{
		Value:     token,
		OwnerId:   owner,
		ExpiresAt: time.Unix(int64(expiresAt), 0),
		Claims:    claims,
		IssuedAt:  time.Unix(int64(issuedAt), 0),
		Type:      int(tokenType),
	}, nil
}

func (t JwtTokenProvider) SignClaims(owner string, claims map[string]any) (string, error) {
	issuer, issuedAt := t.Config.Issuer, time.Now().UTC()
	expiresAt := issuedAt.Add(t.Config.ExpiresIn)
	newClaims := maps.Clone(claims)
	newClaims[issuerClaimKey] = issuer
	newClaims[issuedAtClaimKey] = issuedAt.Unix()
	newClaims[expiryClaimKey] = expiresAt.Unix()
	newClaims[ownerClaimKey] = owner
	if t.Config.TokenType != 0 {
		newClaims[tokenTypeClaimKey] = t.Config.TokenType
	}
	if t.Config.Audience != "" {
		newClaims[audienceClaimKey] = t.Config.Audience
	}
	token := jwt.NewWithClaims(t.Config.signingMethod, jwt.MapClaims(newClaims))
	return t.Config.signer(token)
}

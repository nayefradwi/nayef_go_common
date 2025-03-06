package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nayefradwi/nayef_go_common/core"
)

type JwtTokenProvider struct {
	Config JwtTokenProviderConfig
}

func NewJwtTokenProvider(config JwtTokenProviderConfig) JwtTokenProvider {
	return JwtTokenProvider{
		Config: config,
	}
}

func NewDefaultJwtTokenProvider() JwtTokenProvider {
	return NewJwtTokenProvider(DefaultJwtTokenProviderConfig)
}

func (t JwtTokenProvider) GetClaims(token string) (Token, error) {

	jwtToken, err := jwt.Parse(token, t.Config.parser)
	if err != nil {
		return Token{}, core.UnauthorizedError(err.Error())
	}

	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return Token{}, core.UnauthorizedError("Invalid token")
	}

	owner, _ := claims[ownerClaimKey].(string)
	issuedAt, _ := claims[issuedAtClaimKey].(float64)
	expiresAt, _ := claims[expiryClaimKey].(float64)
	return Token{
		Value:     token,
		OwnerId:   owner,
		ExpiresAt: time.Unix(int64(expiresAt), 0),
		Claims:    claims,
		IssuedAt:  time.Unix(int64(issuedAt), 0),
	}, nil

}

func (t JwtTokenProvider) SignClaims(owner string, claims map[string]interface{}) (string, error) {
	issuer, issuedAt := t.Config.Issuer, time.Now().UTC()
	expiresAt := issuedAt.Add(t.Config.ExpiresIn)
	claims[issuerClaimKey] = issuer
	claims[issuedAtClaimKey] = issuedAt.Unix()
	claims[expiryClaimKey] = expiresAt.Unix()
	claims[ownerClaimKey] = owner
	token := jwt.NewWithClaims(t.Config.signingMethod, jwt.MapClaims(claims))
	return t.Config.signer(token)
}

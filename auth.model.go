package common

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Token struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken,omitempty"`
}

type TokenOptions struct {
	AccessTokenExpiry  time.Time
	RefreshTokenExpiry time.Time
}

func NewTokenOptions(accessTokenExpiry time.Time, refreshTokenExpiry time.Time) *TokenOptions {
	return &TokenOptions{
		AccessTokenExpiry:  accessTokenExpiry,
		RefreshTokenExpiry: refreshTokenExpiry,
	}
}
func DefaultTokenOptions() *TokenOptions {
	return &TokenOptions{
		AccessTokenExpiry:  time.Now().Add(AccessTokenDuration),
		RefreshTokenExpiry: time.Now().Add(RefreshTokenDuration),
	}
}

func generateSignedTokenString(claims map[string]interface{}, secret string, options *TokenOptions) (string, error) {
	setIssuedAtClaim(claims)
	setExpiryDate(claims, options.AccessTokenExpiry)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claims))
	return token.SignedString([]byte(secret))
}

func generateRefreshTokenString(secret string, options *TokenOptions) (string, error) {
	claims := make(map[string]interface{})
	setIssuedAtClaim(claims)
	setExpiryDate(claims, options.RefreshTokenExpiry)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claims))
	return token.SignedString([]byte(secret))
}

func setIssuedAtClaim(claims map[string]interface{}) {
	createdAt := &jwt.NumericDate{Time: time.Now()}
	claims["iat"] = createdAt
}

func setExpiryDate(claims map[string]interface{}, expiry time.Time) {
	claims["exp"] = &jwt.NumericDate{Time: expiry}
}

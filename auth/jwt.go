package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/nayefradwi/nayef_go_common/baseError"
)

func GenerateAccessToken(claims map[string]interface{}, secret string) (Token, error) {
	options := DefaultTokenOptions()
	tokenString, err := generateSignedTokenString(claims, secret, options)
	if err != nil {
		return Token{}, err
	}
	refreshToken, err := generateRefreshTokenString(secret, options)
	return Token{AccessToken: tokenString, RefreshToken: refreshToken}, err
}

func GenerateAccessTokenWithOptions(claims map[string]interface{}, secret string, options *TokenOptions) (Token, error) {
	if options == nil {
		options = DefaultTokenOptions()
	}
	tokenString, err := generateSignedTokenString(claims, secret, options)
	if err != nil {
		return Token{}, err
	}
	refreshToken, err := generateRefreshTokenString(secret, options)
	return Token{AccessToken: tokenString, RefreshToken: refreshToken}, err
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

func DecodeAccessToken(tokenString string, secret string) (map[string]interface{}, error) {
	if isParsed, token := verifyToken(tokenString, secret); isParsed {
		claims := parseToken(token)
		isValid := claims.Valid()
		if isValid != nil {
			return nil, baseError.NewUnAuthorizedError()
		}
		return claims, nil
	}
	return nil, baseError.NewUnAuthorizedError()
}

func verifyToken(tokenString string, secret string) (bool, *jwt.Token) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, baseError.NewUnAuthorizedError()
		}
		return []byte(secret), nil
	})
	return err == nil && token.Valid, token
}

func parseToken(token *jwt.Token) jwt.MapClaims {
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok {
		return claims
	}
	return jwt.MapClaims{}
}

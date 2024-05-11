package common

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type ClaimsKey struct{}

type UserKey struct{}

var secret string = ""
var AccessTokenDuration = time.Hour * 24 * 7
var RefreshTokenDuration = time.Hour * 24 * 30

func SetSecret(envSecret string) {
	secret = envSecret
}

func AuthenticationHeaderMiddleware(f http.Handler) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		token := getIfTokenExists(authHeader)
		if len(token) <= 0 {
			WriteResponseFromError(w, NewUnAuthorizedError("Invalid token"))
			return
		}
		claims, err := DecodeAccessToken(token)
		if err != nil {
			WriteResponseFromError(w, NewUnAuthorizedError("Invalid token"))
			return
		}
		ctx := context.WithValue(r.Context(), ClaimsKey{}, claims)
		f.ServeHTTP(w, r.WithContext(ctx))
	})
	return handler
}

func getIfTokenExists(authHeader string) string {
	tokenSplit := strings.Split(authHeader, " ")
	if len(tokenSplit) != 2 {
		return ""
	}
	return tokenSplit[1]
}

func GenerateAccessToken(claims map[string]interface{}) (Token, error) {
	options := DefaultTokenOptions()
	tokenString, err := generateSignedTokenString(claims, secret, options)
	if err != nil {
		return Token{}, err
	}
	refreshToken, err := generateRefreshTokenString(secret, options)
	return Token{AccessToken: tokenString, RefreshToken: refreshToken}, err
}

func GenerateAccessTokenWithOptions(claims map[string]interface{}, options *TokenOptions) (Token, error) {
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

func DecodeAccessToken(tokenString string) (map[string]interface{}, error) {
	if isVerified, token := verifyToken(tokenString); isVerified {
		claims := parseToken(token)
		isValid := claims.Valid()
		if isValid != nil {
			return nil, NewUnAuthorizedError("Invalid token")
		}
		return claims, nil
	}
	return nil, NewUnAuthorizedError("Invalid token")
}

func verifyToken(tokenString string) (bool, *jwt.Token) {
	token, err := jwt.Parse(tokenString, validateTokenMethod)
	return err == nil && token.Valid, token
}

func validateTokenMethod(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, NewUnAuthorizedError("Invalid token")
	}
	return []byte(secret), nil
}

func parseToken(token *jwt.Token) jwt.MapClaims {
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok {
		return claims
	}
	return jwt.MapClaims{}
}

func GetClaimsFromContext(ctx context.Context) map[string]interface{} {
	claims := ctx.Value(ClaimsKey{})
	if claims != nil {
		return claims.(map[string]interface{})
	}
	return nil
}

type UserIdExtractor func(ctx context.Context) int

var defaultUserIdExtractor UserIdExtractor = func(ctx context.Context) int {
	return 0
}

func SetUserIdExtractor(extractor UserIdExtractor) {
	defaultUserIdExtractor = extractor
}

func GetUserIdFromContext(ctx context.Context) int {
	return defaultUserIdExtractor(ctx)
}

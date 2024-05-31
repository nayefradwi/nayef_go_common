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

type GetIdCallback[T string | int] func(ctx context.Context) T

type AuthenticationService[T string | int] struct {
	Options       TokenOptions
	GetIdCallback GetIdCallback[T]
}

func (s AuthenticationService[T]) AuthenticationHeaderMiddleware(f http.Handler) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		token := getIfTokenExists(authHeader)
		if len(token) <= 0 {
			WriteResponseFromError(w, NewUnAuthorizedError("Invalid token"))
			return
		}
		claims, err := s.DecodeAccessToken(token)
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
func (s AuthenticationService[T]) DecodeAccessToken(tokenString string) (map[string]interface{}, error) {
	if isVerified, token := s.verifyToken(tokenString); isVerified {
		claims := parseToken(token)
		isValid := claims.Valid()
		if isValid != nil {
			return nil, NewUnAuthorizedError("Invalid token")
		}
		return claims, nil
	}
	return nil, NewUnAuthorizedError("Invalid token")
}

func (s AuthenticationService[T]) verifyToken(tokenString string) (bool, *jwt.Token) {
	token, err := jwt.Parse(tokenString, s.validateTokenMethod)
	return err == nil && token.Valid, token
}

func (s AuthenticationService[T]) validateTokenMethod(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, NewUnAuthorizedError("Invalid token")
	}
	return []byte(s.Options.Secret), nil
}

func parseToken(token *jwt.Token) jwt.MapClaims {
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok {
		return claims
	}
	return jwt.MapClaims{}
}

func (s AuthenticationService[T]) GenerateAccessToken(claims map[string]interface{}) (Token, error) {
	tokenString, err := s.generateSignedTokenString(claims)
	if err != nil {
		return Token{}, err
	}
	refreshToken, err := s.generateRefreshTokenString()
	return Token{AccessToken: tokenString, RefreshToken: refreshToken}, err
}
func (s AuthenticationService[T]) generateSignedTokenString(claims map[string]interface{}) (string, error) {
	options := s.Options
	setIssuedAtClaim(claims)
	setExpiryDate(claims, options.AccessTokenExpiry)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claims))
	return token.SignedString([]byte(options.Secret))
}

func (s AuthenticationService[T]) generateRefreshTokenString() (string, error) {
	options := s.Options
	claims := make(map[string]interface{})
	setIssuedAtClaim(claims)
	setExpiryDate(claims, options.RefreshTokenExpiry)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claims))
	return token.SignedString([]byte(options.Secret))
}

func setIssuedAtClaim(claims map[string]interface{}) {
	createdAt := &jwt.NumericDate{Time: time.Now()}
	claims["iat"] = createdAt
}

func setExpiryDate(claims map[string]interface{}, expiry time.Time) {
	claims["exp"] = &jwt.NumericDate{Time: expiry}
}

func GetClaimsFromContext(ctx context.Context) map[string]interface{} {
	claims := ctx.Value(ClaimsKey{})
	if claims != nil {
		return claims.(map[string]interface{})
	}
	return nil
}

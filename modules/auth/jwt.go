package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nayefradwi/nayef_go_common/core"
)

type JwtTokenProviderConfig struct {
	SecretKey     string
	ExpiresIn     time.Duration
	SigningMethod jwt.SigningMethod
	Issuer        string
}

var defaultSecretKey = "SuperSecretKeyShouldBeOverriden"
var defaultExpiresIn = time.Hour * 24

func ReplaceDefaultJwtSecretKey(secretKey string) JwtTokenProviderConfig {
	DefaultJwtTokenProviderConfig.SecretKey = secretKey
	return DefaultJwtTokenProviderConfig
}

func ReplaceDefaultJwtExpiresIn(expiresIn time.Duration) JwtTokenProviderConfig {
	DefaultJwtTokenProviderConfig.ExpiresIn = expiresIn
	return DefaultJwtTokenProviderConfig
}

func ReplaceDefaultJwtSigningMethod(signingMethod jwt.SigningMethod) JwtTokenProviderConfig {
	DefaultJwtTokenProviderConfig.SigningMethod = signingMethod
	return DefaultJwtTokenProviderConfig
}

func ReplaceDefaultJwtIssuer(issuer string) JwtTokenProviderConfig {
	DefaultJwtTokenProviderConfig.Issuer = issuer
	return DefaultJwtTokenProviderConfig
}

var DefaultJwtTokenProviderConfig = NewJwtTokenProviderConfig(defaultSecretKey, defaultExpiresIn)

func NewJwtTokenProviderConfig(secretKey string, expiresIn time.Duration) JwtTokenProviderConfig {
	return JwtTokenProviderConfig{
		SecretKey:     secretKey,
		ExpiresIn:     expiresIn,
		SigningMethod: jwt.SigningMethodHS256,
		Issuer:        "AuthModule",
	}
}

func (c JwtTokenProviderConfig) SetSecretKey(secretKey string) JwtTokenProviderConfig {
	c.SecretKey = secretKey
	return c
}

func (c JwtTokenProviderConfig) SetExpiresIn(expiresIn time.Duration) JwtTokenProviderConfig {
	c.ExpiresIn = expiresIn
	return c
}

func (c JwtTokenProviderConfig) SetSigningMethod(signingMethod jwt.SigningMethod) JwtTokenProviderConfig {
	c.SigningMethod = signingMethod
	return c
}

func (c JwtTokenProviderConfig) SetIssuer(issuer string) JwtTokenProviderConfig {
	c.Issuer = issuer
	return c
}

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

func (t JwtTokenProvider) GetClaims(token string) (Token[string], error) {

	jwtToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, core.UnauthorizedError("Invalid token")
		}

		return []byte(t.Config.SecretKey), nil
	})

	if err != nil {
		return Token[string]{}, core.UnauthorizedError(err.Error())
	}

	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return Token[string]{}, core.UnauthorizedError("Invalid token")
	}

	owner, _ := claims[ownerClaimKey].(string)
	issuedAt, _ := claims[issuedAtClaimKey].(float64)
	expiresAt, _ := claims[expiryClaimKey].(float64)
	return Token[string]{
		Value:     token,
		OwnerId:   owner,
		ExpiresAt: time.Unix(int64(expiresAt), 0),
		Claims:    claims,
		issuedAt:  time.Unix(int64(issuedAt), 0),
	}, nil

}

func (t JwtTokenProvider) SignClaims(owner string, claims map[string]interface{}) (string, error) {
	issuer, issuedAt := t.Config.Issuer, time.Now().UTC()
	expiresAt := issuedAt.Add(t.Config.ExpiresIn)
	claims[issuerClaimKey] = issuer
	claims[issuedAtClaimKey] = issuedAt.Unix()
	claims[expiryClaimKey] = expiresAt.Unix()

	token := jwt.NewWithClaims(t.Config.SigningMethod, jwt.MapClaims(claims))
	tokenString, err := token.SignedString([]byte(t.Config.SecretKey))
	if err != nil {
		return "", core.InternalError(err.Error())
	}

	return tokenString, nil
}

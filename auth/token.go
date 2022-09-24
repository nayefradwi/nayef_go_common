package auth

import "time"

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
	// TODO: make expiry configurable
	return &TokenOptions{
		AccessTokenExpiry:  time.Now().Add(time.Hour * 24),
		RefreshTokenExpiry: time.Now().Add(time.Hour * 24 * 30),
	}
}

type Token struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

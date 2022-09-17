package auth

import "time"

type TokenOptions struct {
	Expiry time.Time
}

func DefaultTokenOptions() *TokenOptions {
	// TODO: make expiry configurable
	return &TokenOptions{
		Expiry: time.Now().Add(time.Hour * 24),
	}
}

type Token struct {
	AccessToken string `json:"accessToken"`
}

type RefreshToken struct {
	RefreshToken string `json:"refreshToken"`
	AccessToken  string `json:"accessToken"`
}

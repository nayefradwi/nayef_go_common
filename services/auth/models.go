package auth

const (
	AccessTokenType  = 1
	RefreshTokenType = 2
)

type RefreshToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type ReferenceToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

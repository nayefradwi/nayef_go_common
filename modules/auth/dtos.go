package auth

type TokenDTO struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty,omitzero"`
}

func EmptyTokenDTO() TokenDTO {
	return TokenDTO{}
}

func NewTokenDTO(accessToken string) TokenDTO {
	return TokenDTO{
		AccessToken: accessToken,
	}
}

func NewTokenDTOWithRefresh(accessToken, refreshToken string) TokenDTO {
	return TokenDTO{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
}

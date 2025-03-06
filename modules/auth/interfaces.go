package auth

type ITokenProvider interface {
	GetClaims(token string) (Token, error)
	SignClaims(owner string, claims map[string]interface{}) (string, error)
}

type ITokenStore interface {
	StoreToken(token Token) error
	StoreTokens(tokens ...Token) error
	GetTokenByReference(reference string, tokenType int) (Token, error)
	GetTokenByOwner(ownerId string, tokenType int) (Token, error)
	DeleteToken(reference string) error
	DeleteAllTokensByOwner(ownerId string) error
}

type IRefreshTokenProvider interface {
	GenerateToken(ownerId string, claims map[string]interface{}) (TokenDTO, error)
	GetAccessToken(accessToken string) (Token, error)
	GetRefreshToken(refreshToken string) (Token, error)
	GetAccessTokenProvider() ITokenProvider
}

type IRefreshTokenProviderWithRevoke interface {
	IRefreshTokenProvider
	RevokeToken(reference string) error
	RevokeOwner(ownerId string) error
}

type IReferenceTokenProvider interface {
	GenerateId() string
	GenerateToken(ownerId string, claims map[string]interface{}) (TokenDTO, error)
	GetAccessToken(id string) (Token, error)
	GetRefreshToken(id string) (Token, error)
	RevokeToken(id string) error
	RevokeOwner(ownerId string) error
	GetAccessTokenProvider() ITokenProvider
}

/// example code for how I want to use the interfaces
// package auth

// // REFURNTYPE options:
// // 1. {"access":JWT, "refresh":JWT}
// // 2. {"access":JWT}
// // 3. {"access":REFERENCE, "refresh":REFERENCE}
// // 4. {"access":REFERENCE}
// // 5. {"access":JWT, "refresh":REFERENCE}
// func (s *SomeService) Login(username, password string) (RETURNTYPE, error){
// 	user, err := // get user function call
// 	if err != nil {
// 		return nil, err // this can be mapped instead
// 	}

// 	// check password
// 	if !user.CheckPassword(password) {
// 		return nil, errors.New("invalid password") // different error this is just an example
// 	}

// 	// generate token which can be refresh + access, access only it doesnt matter
// 	// the code should just be like this:
// 	return s.tokenProvider.generateToken(user.ID, user.GetClaims)
// }

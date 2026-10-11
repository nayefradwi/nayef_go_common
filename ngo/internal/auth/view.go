package auth

import (
	"path"
	"strings"
)

type AuthView struct {
	GoModule        string
	ProjectName     string
	Identifier      string
	IdentifierField string
	IdentifierUpper string
	StorePkg        string
	StoreDiField    string
	Password        bool
	Otp             bool
	Mfa             bool
	Jwt             bool
	Redis           bool
	Grpc            bool
}

func newAuthView(req CreateAuthRequest) AuthView {
	identifier := string(req.Identifier)
	storePkg, storeDiField := "authpg", "Pool"
	if req.IsRedisStore() {
		storePkg, storeDiField = "authredis", "Redis"
	}

	return AuthView{
		GoModule:        req.GoModule,
		ProjectName:     path.Base(req.GoModule),
		Identifier:      identifier,
		IdentifierField: strings.ToUpper(identifier[:1]) + identifier[1:],
		IdentifierUpper: strings.ToUpper(identifier),
		StorePkg:        storePkg,
		StoreDiField:    storeDiField,
		Password:        req.IsPassword(),
		Otp:             !req.IsPassword(),
		Mfa:             req.Mfa,
		Jwt:             req.IsJwt(),
		Redis:           req.IsRedisStore(),
		Grpc:            req.IsGrpc(),
	}
}

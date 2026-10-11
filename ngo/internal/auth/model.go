package auth

import "github.com/nayefradwi/nayef_go_common/ngo/internal/common"

type CreateAuthRequest struct {
	common.Auth
	common.TakesInfraTypes
	ServiceType common.ServiceType
	GoModule    string
	RootDirPath string
	Project     common.Project
}

func (r CreateAuthRequest) IsGrpc() bool {
	return r.ServiceType == common.ServiceTypeGrpc
}

func (r CreateAuthRequest) IsPassword() bool {
	return r.Login == common.LoginPassword
}

func (r CreateAuthRequest) IsJwt() bool {
	return r.AccessToken == common.AccessTokenJwt
}

func (r CreateAuthRequest) IsRedisStore() bool {
	return r.Store == common.InfraTypeRedis
}

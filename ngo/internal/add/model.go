package add

import "github.com/nayefradwi/nayef_go_common/ngo/internal/common"

type CreateFeatureRequest struct {
	common.TakesFeatures
	common.TakesInfraTypes
	Name        string
	ServiceType common.ServiceType
	GoModule    string
	RootDirPath string
}

func (r CreateFeatureRequest) IsGrpc() bool {
	return r.ServiceType == common.ServiceTypeGrpc
}

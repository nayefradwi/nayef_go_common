package new

import (
	"github.com/nayefradwi/nayef_go_common/ngo/internal/common"
)

type CreateNewProjectRequest struct {
	common.TakesFeatures
	common.TakesInfraTypes
	Name           string
	ServiceType    common.ServiceType
	WithValidation bool
	RootDirPath    string
	GoModule       string
}

func (r CreateNewProjectRequest) IsGrpc() bool {
	return r.ServiceType == common.ServiceTypeGrpc
}

var featureToPackage = map[common.Feature]string{
	common.FeatureLocking:    LOCKING,
	common.FeaturePagination: PAGINATION,
}

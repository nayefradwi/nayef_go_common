package add

import "github.com/nayefradwi/nayef_go_common/ngo/internal/common"

type CreateFeatureRequest struct {
	common.TakesAuthType
	common.TakesFeatures
	common.TakesInfraTypes
	Name        string
	GoModule    string
	RootDirPath string
}

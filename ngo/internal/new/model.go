package new

import (
	"github.com/nayefradwi/nayef_go_common/ngo/internal/common"
)

type ServiceType string

const (
	ServiceTypeRest ServiceType = "REST"
	ServiceTypeGrpc ServiceType = "gRPC"
)

type ProviderType string

const (
	ProviderTypeAWS ProviderType = "AWS"
)

type DeploymentType string

const (
	DeploymentTypeManual       DeploymentType = "None"
	DeploymentTypeDokploy      DeploymentType = "Dokploy"
	DeploymentTypeManagedInfra DeploymentType = "Managed"
)

type CreateNewProjectRequest struct {
	common.TakesFeatures
	common.TakesInfraTypes
	common.TakesAuthType
	Name                     string
	ServiceType              ServiceType
	WithValidation           bool
	ProviderType             ProviderType
	StagingDeploymentType    DeploymentType
	ProductionDeploymentType DeploymentType
	RootDirPath              string
	GoModule                 string
}

func (r CreateNewProjectRequest) NeedsInfra(dt DeploymentType) bool {
	return dt == DeploymentTypeDokploy
}

var providerToDeploymentType = map[ProviderType][]DeploymentType{
	ProviderTypeAWS: {DeploymentTypeDokploy, DeploymentTypeManagedInfra},
}

var featureToPackage = map[common.Feature]string{
	common.FeatureLocking:    LOCKING,
	common.FeatureOtp:        OTP,
	common.FeaturePagination: PAGINATION,
}

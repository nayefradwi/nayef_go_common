package new

import "slices"

type ServiceType string

const (
	ServiceTypeRest ServiceType = "REST"
	ServiceTypeGrpc ServiceType = "gRPC"
)

type AuthType string

const (
	AuthTypeJWT     AuthType = "JWT"
	AuthTypeRefresh AuthType = "JWT + Revokable Refresh"
	AuthTypeOpaque  AuthType = "Opaque Tokens"
	AuthTypeNone    AuthType = "Something Else"
)

type InfraType string

const (
	InfraTypePostgres InfraType = "postgresql"
	InfraTypeRedis    InfraType = "redis"
)

type Feature string

const (
	FeatureLocking    Feature = "Locking"
	FeatureOtp        Feature = "OTP"
	FeaturePagination Feature = "Pagination"
)

type DBLibrary string

const (
	DBLibrarySqlc DBLibrary = "sqlc"
	DBLibraryNone DBLibrary = "Something Else"
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
	Name                     string
	ServiceType              ServiceType
	AuthType                 AuthType
	WithValidation           bool
	Features                 []Feature
	InfraTypes               []InfraType
	DBLibrary                DBLibrary
	ProviderType             ProviderType
	StagingDeploymentType    DeploymentType
	ProductionDeploymentType DeploymentType
	RootDirPath              string
	GoModule                 string
}

func (r CreateNewProjectRequest) HasPostgres() bool {
	return slices.Contains(r.InfraTypes, InfraTypePostgres)
}

func (r CreateNewProjectRequest) HasRedis() bool {
	return slices.Contains(r.InfraTypes, InfraTypeRedis)
}

func (r CreateNewProjectRequest) HasAuth() bool {
	return r.AuthType != AuthTypeNone
}

func (r CreateNewProjectRequest) HasFeature(f Feature) bool {
	return slices.Contains(r.Features, f)
}

func (r CreateNewProjectRequest) NeedsInfra(dt DeploymentType) bool {
	return dt == DeploymentTypeDokploy
}

var providerToDeploymentType = map[ProviderType][]DeploymentType{
	ProviderTypeAWS: {DeploymentTypeDokploy, DeploymentTypeManagedInfra},
}

var featureToPackage = map[Feature]string{
	FeatureLocking:    LOCKING,
	FeatureOtp:        OTP,
	FeaturePagination: PAGINATION,
}

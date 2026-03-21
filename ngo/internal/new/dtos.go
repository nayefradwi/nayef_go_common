package new

type ServiceType string

const (
	ServiceTypeRest ServiceType = "REST"
	ServiceTypeGrpc ServiceType = "gRPC"
	ServiceTypeBoth ServiceType = "Both"
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

var AllFeatures = []Feature{
	FeatureLocking,
	FeatureOtp,
}

var AllInfraTypes = []InfraType{
	InfraTypePostgres,
	InfraTypeRedis,
}

type DBLibrary string

const (
	DBLibrarySqlc DBLibrary = "sqlc"
	DBLibraryNone DBLibrary = "Something Else"
)

type CreateNewProjectRequest struct {
	Name           string
	ServiceType    ServiceType
	AuthType       AuthType
	WithValidation bool
	Features       []Feature
	InfraTypes     []InfraType
	DBLibrary      DBLibrary
	headDir        dir
	packages       []string
}

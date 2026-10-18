package deploy

import "github.com/nayefradwi/nayef_go_common/ngo/internal/common"

type CloudProvider string

var (
	Aws   CloudProvider = "aws"
	Gcp   CloudProvider = "gcp"
	Azure CloudProvider = "Azure"
	Vps   CloudProvider = "Vps"
)

type DeploymentSize string

var (
	Small  DeploymentSize = "small"
	Medium DeploymentSize = "medium"
	Large  DeploymentSize = "large"
)

type CreateDeploymentRequest struct {
	common.TakesInfraTypes
	Name        string
	ServiceType common.ServiceType
	Environment string
	Provider    CloudProvider
	Size        DeploymentSize
	RootDirPath string
}

func (r CreateDeploymentRequest) IsGrpc() bool {
	return r.ServiceType == common.ServiceTypeGrpc
}

func (r CreateDeploymentRequest) IsSmall() bool {
	return r.Size == Small
}

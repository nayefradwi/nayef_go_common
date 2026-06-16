package new

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/nayefradwi/nayef_go_common/errors"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/log"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/printer"
)

func runGoTidy(req CreateNewProjectRequest) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = req.RootDirPath
	return cmd.Run()
}

func runGoFmt(req CreateNewProjectRequest) error {
	cmd := exec.Command("go", "fmt", "./...")
	cmd.Dir = req.RootDirPath
	return cmd.Run()
}

func createGoMod(req CreateNewProjectRequest) error {
	stop := printer.Spin("Initializing Go module")

	log.Info("running go mod init at", "path", req.GoModule)
	cmd := exec.Command("go", "mod", "init", req.GoModule)
	cmd.Dir = req.RootDirPath

	err := cmd.Run()
	stop(err)
	if err != nil {
		return fmt.Errorf("failed to initialize go module at %s %w", req.RootDirPath, err)
	}

	return nil
}

func getPackagesFromRequest(req CreateNewProjectRequest) []string {
	packages := []string{GODOTENV, COMMON_ERRORS, TESTIFY}
	if req.ServiceType == ServiceTypeRest {
		packages = append(packages, CHI, HTTPUTIL)
	} else {
		packages = append(packages, GRPC, GRPCUTIL, ERRORSPB)
	}

	if req.AuthType != AuthTypeNone {
		packages = append(packages, AUTH)
	}

	if req.WithValidation {
		packages = append(packages, VALIDATION)
	}

	if req.HasPostgres() {
		packages = append(packages, PGX, PGUTIL)
	}

	if req.HasRedis() {
		packages = append(packages, REDIS, REDISUTIL)
	}

	for _, feature := range req.Features {
		packages = append(packages, featureToPackage[feature])
	}

	return packages
}

func installGoPackages(req CreateNewProjectRequest) error {
	stop := printer.Spin("Installing go packages")
	packages := getPackagesFromRequest(req)
	args := append([]string{"get"}, packages...)
	cmd := exec.Command("go", args...)
	cmd.Dir = req.RootDirPath

	err := cmd.Run()
	stop(err)
	if err != nil {
		return fmt.Errorf("failed to get go packages %w", err)
	}

	return nil
}
func generateRootFromRequest(req CreateNewProjectRequest) error {
	runner := errors.ResultRunnerWithParam[CreateNewProjectRequest]{}
	runner.Do(req, renderGitIgnore)
	runner.Do(req, renderEnv)
	runner.Do(req, renderAirToml)
	return runner.Error
}

func generateInternalFromRequest(req CreateNewProjectRequest) error {
	runner := errors.ResultRunnerWithParam[CreateNewProjectRequest]{}
	runner.Do(req, renderConfig)
	runner.Do(req, renderDi)
	runner.Do(req, renderHealth)
	return runner.Error
}

func generateCmdFromRequest(req CreateNewProjectRequest) error {
	runner := errors.ResultRunnerWithParam[CreateNewProjectRequest]{}
	runner.Do(req, renderMain)
	runner.Do(req, renderBootstrap)
	runner.Do(req, renderRouter)
	return runner.Error
}

func generateStagingDeploymentFromRequest(req CreateNewProjectRequest) error {
	if req.StagingDeploymentType == DeploymentTypeManual {
		return nil
	}

	if req.StagingDeploymentType == DeploymentTypeDokploy {
		runner := errors.ResultRunnerWithParam[CreateNewProjectRequest]{}
		runner.Do(req, renderStagingDockerCompose)
		runner.Do(req, func(r CreateNewProjectRequest) error {
			return renderTerraformEnv(r, STAGING)
		})
		return runner.Error
	}

	return nil
}

func generateProdDeploymentFromRequest(req CreateNewProjectRequest) error {
	if req.ProductionDeploymentType == DeploymentTypeManual {
		return nil
	}

	if req.ProductionDeploymentType == DeploymentTypeDokploy {
		runner := errors.ResultRunnerWithParam[CreateNewProjectRequest]{}
		runner.Do(req, renderProductionDockerCompose)
		runner.Do(req, func(r CreateNewProjectRequest) error {
			return renderTerraformEnv(r, PRODUCTION)
		})
		return runner.Error
	}

	return nil
}

func generateDeploymentsFromRequest(req CreateNewProjectRequest) error {
	runner := errors.ResultRunnerWithParam[CreateNewProjectRequest]{}
	runner.Do(req, renderDockerfile)
	runner.Do(req, renderLocalDockerCompose)
	runner.Do(req, renderLocalEnv)
	runner.Do(req, renderTerraformModules)
	runner.Do(req, generateStagingDeploymentFromRequest)
	runner.Do(req, generateProdDeploymentFromRequest)
	return runner.Error
}

func generateGithubFromRequest(req CreateNewProjectRequest) error {
	runner := errors.ResultRunnerWithParam[CreateNewProjectRequest]{}
	runner.Do(req, renderGithubCI)
	return runner.Error
}

func generateConfigFromRequest(req CreateNewProjectRequest) error {
	runner := errors.ResultRunnerWithParam[CreateNewProjectRequest]{}
	if req.DBLibrary == DBLibrarySqlc {
		runner.Do(req, renderSqlcConfig)
	}

	return runner.Error
}

func generateSSHKeyPair(dir string) error {
	keyPath := filepath.Join(dir, "id_rsa.pem")
	cmd := exec.Command("ssh-keygen", "-t", "ed25519", "-f", keyPath, "-N", "")
	return cmd.Run()
}

func generateSSHKeys(req CreateNewProjectRequest) error {
	stop := printer.Spin("Generating SSH keys")
	var err error
	if req.NeedsInfra(req.StagingDeploymentType) {
		dir := filepath.Join(req.RootDirPath, DEPLOYMENTS, TERRAFORM, STAGING)
		err = generateSSHKeyPair(dir)
	}
	if err == nil && req.NeedsInfra(req.ProductionDeploymentType) {
		dir := filepath.Join(req.RootDirPath, DEPLOYMENTS, TERRAFORM, PRODUCTION)
		err = generateSSHKeyPair(dir)
	}
	stop(err)
	return err
}

func generateCodeFromRequest(req CreateNewProjectRequest) error {
	runner := errors.ResultRunnerWithParam[CreateNewProjectRequest]{}
	runner.Do(req, generateRootFromRequest)
	runner.Do(req, generateInternalFromRequest)
	runner.Do(req, generateCmdFromRequest)
	runner.Do(req, generateDeploymentsFromRequest)
	runner.Do(req, generateConfigFromRequest)
	runner.Do(req, generateGithubFromRequest)

	if runner.Error != nil {
		return runner.Error
	}

	printer.Success("Generated go code")
	return nil
}

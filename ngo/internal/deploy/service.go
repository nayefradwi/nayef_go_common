package deploy

import (
	"fmt"
	"os"

	"github.com/nayefradwi/nayef_go_common/errors"
)

func generateDeployment(req CreateDeploymentRequest) error {
	if err := checkEnvironmentNotExists(req); err != nil {
		return err
	}

	if err := checkResourceNameLength(req); err != nil {
		return err
	}

	runner := errors.ResultRunnerWithParam[CreateDeploymentRequest]{}
	runner.Do(req, renderTerraform)
	runner.Do(req, renderTerraformVars)
	runner.Do(req, renderRunnerInit)
	runner.Do(req, renderDockerCompose)
	runner.Do(req, renderContinousDelivery)
	runner.Do(req, renderDeployScript)
	runner.Do(req, renderTodo)
	if runner.Error != nil {
		os.RemoveAll(envDir(req))
	}

	return runner.Error
}

func checkEnvironmentNotExists(req CreateDeploymentRequest) error {
	dir := envDir(req)
	if _, err := os.Stat(dir); err == nil {
		return fmt.Errorf("%s already exists, pick another environment name", dir)
	}

	return nil
}

func checkResourceNameLength(req CreateDeploymentRequest) error {
	name := req.Name + "-" + req.Environment
	if len(name) > MAX_RESOURCE_NAME {
		return fmt.Errorf("%q is %d chars, cloud resource names allow %d: pick a shorter environment name", name, len(name), MAX_RESOURCE_NAME)
	}

	return nil
}

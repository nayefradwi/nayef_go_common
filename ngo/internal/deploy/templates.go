package deploy

import (
	"embed"
	"fmt"
	"path"
	"path/filepath"

	"github.com/nayefradwi/nayef_go_common/ngo/internal/common"
)

//go:embed templates/*
var templatesFs embed.FS

func envDir(req CreateDeploymentRequest) string {
	return filepath.Join(req.RootDirPath, DEPLOYMENTS, req.Environment)
}

func providerDir(provider CloudProvider) (string, error) {
	switch provider {
	case Aws:
		return PROVIDER_AWS, nil
	case Gcp:
		return PROVIDER_GCP, nil
	case Azure:
		return PROVIDER_AZURE, nil
	case Vps:
		return PROVIDER_VPS, nil
	default:
		return "", fmt.Errorf("unknown provider %q", provider)
	}
}

func terraformSize(size DeploymentSize) (string, error) {
	switch size {
	case Small:
		return TF_SMALL, nil
	case Medium:
		return TF_MEDIUM, nil
	case Large:
		return TF_LARGE, nil
	default:
		return "", fmt.Errorf("unknown deployment size %q", size)
	}
}

func terraformTemplate(provider CloudProvider, size DeploymentSize) (string, error) {
	if provider == Vps {
		return "", fmt.Errorf("no terraform for provider %q", provider)
	}

	dir, err := providerDir(provider)
	if err != nil {
		return "", err
	}

	s, err := terraformSize(size)
	if err != nil {
		return "", err
	}

	return path.Join(dir, s), nil
}

func renderTerraform(req CreateDeploymentRequest) error {
	if req.Provider == Vps {
		return nil
	}

	tmpl, err := terraformTemplate(req.Provider, req.Size)
	if err != nil {
		return err
	}

	filePath := filepath.Join(envDir(req), TERRAFORM, "main.tf")
	return common.RenderToFile(templatesFs, tmpl, filePath, req)
}

func renderTerraformVars(req CreateDeploymentRequest) error {
	if req.Provider == Vps {
		return nil
	}

	dir, err := providerDir(req.Provider)
	if err != nil {
		return err
	}

	filePath := filepath.Join(envDir(req), TERRAFORM, "terraform.tfvars")
	return common.RenderToFile(templatesFs, path.Join(dir, TF_VARS), filePath, req)
}

func renderDockerCompose(req CreateDeploymentRequest) error {
	if req.Size != Small {
		return nil
	}

	filePath := filepath.Join(envDir(req), "docker-compose.yaml")
	return common.RenderToFile(templatesFs, TMPL_DOCKER_COMPOSE, filePath, req)
}

// small runs compose on a self-hosted runner on every provider, so it shares the vps workflow
func renderContinousDelivery(req CreateDeploymentRequest) error {
	dir, err := providerDir(req.Provider)
	if err != nil {
		return err
	}

	if req.Size == Small {
		dir = PROVIDER_VPS
	}

	tmpl := path.Join(dir, TMPL_DEPLOY)
	filePath := filepath.Join(req.RootDirPath, GITHUB, WORKFLOWS, "deploy-"+req.Environment+".yml")
	return common.RenderToFile(templatesFs, tmpl, filePath, req)
}

func renderTodo(req CreateDeploymentRequest) error {
	dir, err := providerDir(req.Provider)
	if err != nil {
		return err
	}

	filePath := filepath.Join(envDir(req), "TODO.md")
	return common.RenderToFile(templatesFs, path.Join(dir, TMPL_TODO), filePath, req)
}

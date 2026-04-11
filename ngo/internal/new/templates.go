package new

import (
	"embed"
	"os"
	"path/filepath"
	"text/template"

	"github.com/nayefradwi/nayef_go_common/errors"
)

//go:embed templates/*
var templatesFs embed.FS

func renderToFile[T any](tmplName, filePath string, input T) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}

	tmpl, err := template.ParseFS(templatesFs, "templates/"+tmplName)
	if err != nil {
		return err
	}

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, input)
}

func renderMain(req CreateNewProjectRequest) error {
	filePath := filepath.Join(req.RootDirPath, CMD, API, MAIN+"."+GO)
	return renderToFile(TMPL_MAIN, filePath, req)
}

func renderBootstrap(req CreateNewProjectRequest) error {
	filePath := filepath.Join(req.RootDirPath, CMD, API, BOOTSTRAP+"."+GO)
	view := newBootstrapView(req)
	return renderToFile(TMPL_BOOTSTRAP, filePath, view)
}

func renderSqlcConfig(req CreateNewProjectRequest) error {
	sqlcPath := filepath.Join(req.RootDirPath, INTERNAL, INFRA, SQLC, QUERIES)
	if err := os.MkdirAll(sqlcPath, 0755); err != nil {
		return err
	}

	migrationsPath := filepath.Join(req.RootDirPath, INTERNAL, INFRA, MIGRATIONS)
	if err := os.MkdirAll(migrationsPath, 0755); err != nil {
		return err
	}

	filePath := filepath.Join(req.RootDirPath, CONFIG, SQLC+"."+YAML)
	return renderToFile(TMPL_SQLC, filePath, "")
}

func renderEnv(req CreateNewProjectRequest) error {
	filepath := filepath.Join(req.RootDirPath, ENV)
	view := newEnvView(req)
	return renderToFile(TMPL_ENV, filepath, view)
}

func renderGitIgnore(req CreateNewProjectRequest) error {
	filepath := filepath.Join(req.RootDirPath, GITIGNORE)
	return renderToFile(TMPL_GITIGNORE, filepath, "")
}

func renderConfig(req CreateNewProjectRequest) error {
	filepath := filepath.Join(req.RootDirPath, INTERNAL, CONFIG, CONFIG+"."+GO)
	view := newConfigView(req)
	return renderToFile(TMPL_CONFIG, filepath, view)
}

func renderDi(req CreateNewProjectRequest) error {
	filepath := filepath.Join(req.RootDirPath, INTERNAL, DI, DI+"."+GO)
	view := newDiView(req)
	return renderToFile(TMPL_DI, filepath, view)
}

func renderDockerfile(req CreateNewProjectRequest) error {
	filePath := filepath.Join(req.RootDirPath, BUILD, DOCKERFILE)
	return renderToFile(TMPL_DOCKERFILE, filePath, req)
}

func renderLocalDockerCompose(req CreateNewProjectRequest) error {
	filePath := filepath.Join(req.RootDirPath, DEPLOYMENTS, LOCAL, DOCKER_COMPOSE+"."+YAML)
	view := newLocalDockerComposeView(req)
	return renderToFile(TMPL_DOCKER_COMPOSE, filePath, view)
}

func renderLocalEnv(req CreateNewProjectRequest) error {
	filePath := filepath.Join(req.RootDirPath, DEPLOYMENTS, LOCAL, ENV)
	view := newLocalEnvView(req)
	return renderToFile(TMPL_LOCAL_ENV, filePath, view)
}

func renderHealth(req CreateNewProjectRequest) error {
	filepath := filepath.Join(req.RootDirPath, INTERNAL, HEALTH, HANDLER+"."+GO)
	view := newHealthView(req)
	return renderToFile(TMPL_HEALTH, filepath, view)
}

func renderRouter(req CreateNewProjectRequest) error {
	filepath := filepath.Join(req.RootDirPath, CMD, API, ROUTER+"."+GO)
	view := newRouterView(req)
	return renderToFile(TMPL_ROUTER, filepath, view)
}

func renderAirToml(req CreateNewProjectRequest) error {
	filePath := filepath.Join(req.RootDirPath, AIR_TOML+"."+TOML)
	return renderToFile(TMPL_AIR_TOML, filePath, "")
}

func renderStagingDockerCompose(req CreateNewProjectRequest) error {
	filepath := filepath.Join(req.RootDirPath, DEPLOYMENTS, STAGING, DOCKER_COMPOSE+"."+YAML)
	view := newVpsDockerComposeView(req, STAGING)
	return renderToFile(TMPL_VPS_DOCKER_COMPOSE, filepath, view)
}

func renderProductionDockerCompose(req CreateNewProjectRequest) error {
	filepath := filepath.Join(req.RootDirPath, DEPLOYMENTS, PRODUCTION, DOCKER_COMPOSE+"."+YAML)
	view := newVpsDockerComposeView(req, PRODUCTION)
	return renderToFile(TMPL_VPS_DOCKER_COMPOSE, filepath, view)
}

func renderTerraformVpsModule(req CreateNewProjectRequest) error {
	moduleDir := filepath.Join(req.RootDirPath, DEPLOYMENTS, TERRAFORM, MODULES, VPS)
	runner := errors.ResultRunnerWithParam[TerraformView]{}
	view := TerraformView{Name: req.Name}
	runner.Do(view, func(v TerraformView) error {
		return renderToFile(TMPL_TF_VPS_MAIN, filepath.Join(moduleDir, "main.tf"), v)
	})
	runner.Do(view, func(v TerraformView) error {
		return renderToFile(TMPL_TF_VPS_VARIABLES, filepath.Join(moduleDir, "variables.tf"), v)
	})
	runner.Do(view, func(v TerraformView) error {
		return renderToFile(TMPL_TF_VPS_OUTPUTS, filepath.Join(moduleDir, "outputs.tf"), v)
	})
	return runner.Error
}

func renderTerraformEnv(req CreateNewProjectRequest, environment string) error {
	envDir := filepath.Join(req.RootDirPath, DEPLOYMENTS, TERRAFORM, environment)
	view := newTerraformView(req, environment)
	runner := errors.ResultRunnerWithParam[TerraformView]{}
	runner.Do(view, func(v TerraformView) error {
		return renderToFile(TMPL_TF_ENV_MAIN, filepath.Join(envDir, "main.tf"), v)
	})
	runner.Do(view, func(v TerraformView) error {
		return renderToFile(TMPL_TF_ENV_VARIABLES, filepath.Join(envDir, "variables.tf"), v)
	})
	runner.Do(view, func(v TerraformView) error {
		return renderToFile(TMPL_TF_ENV_TFVARS, filepath.Join(envDir, "terraform.tfvars.example"), v)
	})
	return runner.Error
}

func renderTerraformModules(req CreateNewProjectRequest) error {
	if !req.NeedsInfra(req.StagingDeploymentType) && !req.NeedsInfra(req.ProductionDeploymentType) {
		return nil
	}
	return renderTerraformVpsModule(req)
}

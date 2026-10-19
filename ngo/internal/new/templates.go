package new

import (
	"embed"
	"os"
	"path/filepath"

	"github.com/nayefradwi/nayef_go_common/ngo/internal/common"
)

//go:embed templates/*
var templatesFs embed.FS

func renderMain(req CreateNewProjectRequest) error {
	filePath := filepath.Join(req.RootDirPath, CMD, API, MAIN+"."+GO)
	return common.RenderToFile(templatesFs, TMPL_MAIN, filePath, req)
}

func renderBootstrap(req CreateNewProjectRequest) error {
	filePath := filepath.Join(req.RootDirPath, CMD, API, BOOTSTRAP+"."+GO)
	view := newBootstrapView(req)
	return common.RenderToFile(templatesFs, TMPL_BOOTSTRAP, filePath, view)
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

	filePath := filepath.Join(req.RootDirPath, SQLC+"."+YAML)
	return common.RenderToFile(templatesFs, TMPL_SQLC, filePath, "")
}

func renderEnv(req CreateNewProjectRequest) error {
	filepath := filepath.Join(req.RootDirPath, ENV)
	view := newEnvView(req)
	return common.RenderToFile(templatesFs, TMPL_ENV, filepath, view)
}

func renderGitIgnore(req CreateNewProjectRequest) error {
	filepath := filepath.Join(req.RootDirPath, GITIGNORE)
	return common.RenderToFile(templatesFs, TMPL_GITIGNORE, filepath, "")
}

func renderConfig(req CreateNewProjectRequest) error {
	filepath := filepath.Join(req.RootDirPath, INTERNAL, CONFIG, CONFIG+"."+GO)
	view := newConfigView(req)
	return common.RenderToFile(templatesFs, TMPL_CONFIG, filepath, view)
}

func renderDi(req CreateNewProjectRequest) error {
	filepath := filepath.Join(req.RootDirPath, INTERNAL, DI, DI+"."+GO)
	view := newDiView(req)
	return common.RenderToFile(templatesFs, TMPL_DI, filepath, view)
}

func renderDockerfile(req CreateNewProjectRequest) error {
	goVersion, err := readGoVersion(req.RootDirPath)
	if err != nil {
		return err
	}

	filePath := filepath.Join(req.RootDirPath, BUILD, DOCKERFILE)
	return common.RenderToFile(templatesFs, TMPL_DOCKERFILE, filePath, newDockerfileView(req, goVersion))
}

func renderMigrate(req CreateNewProjectRequest) error {
	if !req.HasPostgres() {
		return nil
	}

	goVersion, err := readGoVersion(req.RootDirPath)
	if err != nil {
		return err
	}

	view := newDockerfileView(req, goVersion)
	dockerfile := filepath.Join(req.RootDirPath, BUILD, MIGRATE_DOCKERFILE)
	if err := common.RenderToFile(templatesFs, TMPL_MIGRATE_DOCKERFILE, dockerfile, view); err != nil {
		return err
	}

	script := filepath.Join(req.RootDirPath, BUILD, MIGRATE_SCRIPT)
	return common.RenderToFile(templatesFs, TMPL_MIGRATE_SCRIPT, script, view)
}

func renderLocalDockerCompose(req CreateNewProjectRequest) error {
	if !req.HasPostgres() && !req.HasRedis() {
		return nil
	}

	filePath := filepath.Join(req.RootDirPath, DEPLOYMENTS, LOCAL, DOCKER_COMPOSE+"."+YAML)
	view := newLocalDockerComposeView(req)
	return common.RenderToFile(templatesFs, TMPL_DOCKER_COMPOSE, filePath, view)
}

func renderHealth(req CreateNewProjectRequest) error {
	filepath := filepath.Join(req.RootDirPath, INTERNAL, HEALTH, HANDLER+"."+GO)
	view := newHealthView(req)
	return common.RenderToFile(templatesFs, TMPL_HEALTH, filepath, view)
}

func renderRouter(req CreateNewProjectRequest) error {
	filepath := filepath.Join(req.RootDirPath, CMD, API, ROUTER+"."+GO)
	view := newRouterView(req)
	return common.RenderToFile(templatesFs, TMPL_ROUTER, filepath, view)
}

func renderAirToml(req CreateNewProjectRequest) error {
	filePath := filepath.Join(req.RootDirPath, AIR_TOML+"."+TOML)
	return common.RenderToFile(templatesFs, TMPL_AIR_TOML, filePath, req)
}

func renderJustfile(req CreateNewProjectRequest) error {
	filePath := filepath.Join(req.RootDirPath, JUSTFILE)
	return common.RenderToFile(templatesFs, TMPL_JUSTFILE, filePath, req)
}

func renderGithubCI(req CreateNewProjectRequest) error {
	filePath := filepath.Join(req.RootDirPath, GITHUB, WORKFLOWS, CI+"."+YML)
	return common.RenderToFile(templatesFs, TMPL_GITHUB_CI, filePath, req)
}

func renderBufConfig(req CreateNewProjectRequest) error {
	filePath := filepath.Join(req.RootDirPath, "buf."+YAML)
	return common.RenderToFile(templatesFs, TMPL_BUF, filePath, "")
}

func renderBufGenConfig(req CreateNewProjectRequest) error {
	filePath := filepath.Join(req.RootDirPath, "buf.gen."+YAML)
	return common.RenderToFile(templatesFs, TMPL_BUF_GEN, filePath, newProtoView(req))
}

func renderProto(req CreateNewProjectRequest) error {
	filePath := filepath.Join(req.RootDirPath, PROTO, req.Name, V1, req.Name+"."+PROTO)
	return common.RenderToFile(templatesFs, TMPL_PROTO, filePath, newProtoView(req))
}

func renderServiceHandler(req CreateNewProjectRequest) error {
	filePath := filepath.Join(req.RootDirPath, INTERNAL, req.Name, HANDLER+"."+GO)
	return common.RenderToFile(templatesFs, TMPL_SERVICE_HANDLER, filePath, newProtoView(req))
}

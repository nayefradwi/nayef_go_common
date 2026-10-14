package new

import (
	"embed"
	"os"
	"path/filepath"
	"text/template"
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

	filePath := filepath.Join(req.RootDirPath, SQLC+"."+YAML)
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
	goVersion, err := readGoVersion(req.RootDirPath)
	if err != nil {
		return err
	}

	filePath := filepath.Join(req.RootDirPath, BUILD, DOCKERFILE)
	return renderToFile(TMPL_DOCKERFILE, filePath, newDockerfileView(req, goVersion))
}

func renderLocalDockerCompose(req CreateNewProjectRequest) error {
	if !req.HasPostgres() && !req.HasRedis() {
		return nil
	}

	filePath := filepath.Join(req.RootDirPath, DEPLOYMENTS, LOCAL, DOCKER_COMPOSE+"."+YAML)
	view := newLocalDockerComposeView(req)
	return renderToFile(TMPL_DOCKER_COMPOSE, filePath, view)
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
	return renderToFile(TMPL_AIR_TOML, filePath, req)
}

func renderJustfile(req CreateNewProjectRequest) error {
	filePath := filepath.Join(req.RootDirPath, JUSTFILE)
	return renderToFile(TMPL_JUSTFILE, filePath, req)
}

func renderGithubCI(req CreateNewProjectRequest) error {
	filePath := filepath.Join(req.RootDirPath, GITHUB, WORKFLOWS, CI+"."+YML)
	return renderToFile(TMPL_GITHUB_CI, filePath, req)
}

func renderBufConfig(req CreateNewProjectRequest) error {
	filePath := filepath.Join(req.RootDirPath, "buf."+YAML)
	return renderToFile(TMPL_BUF, filePath, "")
}

func renderBufGenConfig(req CreateNewProjectRequest) error {
	filePath := filepath.Join(req.RootDirPath, "buf.gen."+YAML)
	return renderToFile(TMPL_BUF_GEN, filePath, newProtoView(req))
}

func renderProto(req CreateNewProjectRequest) error {
	filePath := filepath.Join(req.RootDirPath, PROTO, req.Name, V1, req.Name+"."+PROTO)
	return renderToFile(TMPL_PROTO, filePath, newProtoView(req))
}

func renderServiceHandler(req CreateNewProjectRequest) error {
	filePath := filepath.Join(req.RootDirPath, INTERNAL, req.Name, HANDLER+"."+GO)
	return renderToFile(TMPL_SERVICE_HANDLER, filePath, newProtoView(req))
}

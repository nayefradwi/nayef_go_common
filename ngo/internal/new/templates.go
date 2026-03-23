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
	return renderToFile("main.go.tmpl", filePath, req)
}

func renderBootstrap(req CreateNewProjectRequest) error {
	filePath := filepath.Join(req.RootDirPath, CMD, API, BOOTSTRAP+"."+GO)
	return renderToFile("bootstrap.go.tmpl", filePath, req)
}

func renderSqlcConfig(req CreateNewProjectRequest) error {
	filePath := filepath.Join(req.RootDirPath, CONFIG, SQLC+"."+YAML)
	return renderToFile("sqlc.yaml.tmpl", filePath, req)
}

func renderEnv(req CreateNewProjectRequest) error {
	filepath := filepath.Join(req.RootDirPath, ENV)
	return renderToFile(".env.tmpl", filepath, req)
}

func renderGitIgnore(req CreateNewProjectRequest) error {
	filepath := filepath.Join(req.RootDirPath, GITIGNORE)
	return renderToFile(".gitignore.tmpl", filepath, "")
}

func renderConfig(req CreateNewProjectRequest) error {
	filepath := filepath.Join(req.RootDirPath, INTERNAL, CONFIG, CONFIG+"."+GO)
	return renderToFile("config.go.tmpl", filepath, req)
}

func renderDi(req CreateNewProjectRequest) error {
	filepath := filepath.Join(req.RootDirPath, INTERNAL, DI, DI+"."+GO)
	return renderToFile("di.go.tmpl", filepath, req)

}

func renderDockerfile(req CreateNewProjectRequest) error {
	filePath := filepath.Join(req.RootDirPath, BUILD, DOCKERFILE)
	return renderToFile("Dockerfile.tmpl", filePath, req)
}

func renderDockerCompose(req CreateNewProjectRequest) error {
	filePath := filepath.Join(req.RootDirPath, DEPLOYMENTS, LOCAL, DOCKER_COMPOSE+"."+YAML)
	return renderToFile("docker-compose.yaml.tmpl", filePath, req)
}

func renderLocalEnv(req CreateNewProjectRequest) error {
	filePath := filepath.Join(req.RootDirPath, DEPLOYMENTS, LOCAL, ENV)
	return renderToFile("local.env.tmpl", filePath, req)
}

func renderHealth(req CreateNewProjectRequest) error {
	filepath := filepath.Join(req.RootDirPath, INTERNAL, HEALTH, HANDLER+"."+GO)
	return renderToFile("health.go.tmpl", filepath, req)
}

func renderRouter(req CreateNewProjectRequest) error {
	filepath := filepath.Join(req.RootDirPath, CMD, API, ROUTER+"."+GO)
	return renderToFile("router.go.tmpl", filepath, req)

}

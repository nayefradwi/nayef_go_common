package new

import (
	"embed"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed templates/*
var templatesFs embed.FS

func renderToFile(tmplName, filePath string, req CreateNewProjectRequest) error {
	tmpl, err := template.ParseFS(templatesFs, "templates/"+tmplName)
	if err != nil {
		return err
	}

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, req)
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
	filePath := filepath.Join(req.RootDirPath, INTERNAL, INFRA, SQLC, SQLC+"."+YAML)
	return renderToFile("sqlc.yaml.tmpl", filePath, req)
}

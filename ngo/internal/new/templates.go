package new

import (
	"embed"
	"os"
	"path/filepath"
	"slices"
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
	input := EnvTemplateInput{
		ShouldAddDb:     slices.Contains(req.InfraTypes, InfraTypePostgres),
		ShouldAddRedis:  slices.Contains(req.InfraTypes, InfraTypeRedis),
		ShouldAddSecret: req.AuthType != AuthTypeNone,
	}

	return renderToFile(".env.tmpl", filepath, input)
}

func renderGitIgnore(req CreateNewProjectRequest) error {
	filepath := filepath.Join(req.RootDirPath, GITIGNORE)
	return renderToFile(".gitignore.tmpl", filepath, "")
}

func renderConfig(req CreateNewProjectRequest) error {
	filepath := filepath.Join(req.RootDirPath, INTERNAL, CONFIG, CONFIG+"."+GO)
	input := EnvTemplateInput{
		ShouldAddDb:     slices.Contains(req.InfraTypes, InfraTypePostgres),
		ShouldAddRedis:  slices.Contains(req.InfraTypes, InfraTypeRedis),
		ShouldAddSecret: req.AuthType != AuthTypeNone,
	}

	return renderToFile("config.go.tmpl", filepath, input)
}

func renderDi(req CreateNewProjectRequest) error {
	filepath := filepath.Join(req.RootDirPath, INTERNAL, DI, DI+"."+GO)
	shouldAddDb := slices.Contains(req.InfraTypes, InfraTypePostgres)
	shouldAddRedis := slices.Contains(req.InfraTypes, InfraTypeRedis)

	imports := []string{"context", req.GoModule + "/" + INTERNAL + "/" + CONFIG}
	if shouldAddDb {
		imports = append(imports, PGUTIL, PGX+"/"+"pgxpool")
	}

	if shouldAddRedis {
		imports = append(imports, REDISUTIL, REDIS)
	}

	input := DiTemplateInput{
		Imports:        imports,
		ShouldAddDb:    shouldAddDb,
		ShouldAddRedis: shouldAddRedis,
	}

	return renderToFile("di.go.tmpl", filepath, input)

}

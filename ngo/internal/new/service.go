package new

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"

	"github.com/nayefradwi/nayef_go_common/errors"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/log"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/printer"
)

func populateRequestDetails(req *CreateNewProjectRequest) {
	wd, _ := os.Getwd()
	req.RootDirPath = filepath.Join(wd, req.Name)

	path := os.Getenv("GO_MOD_PATH")
	if path == "" {
		path = DEFAULT_MOD_PATH
	}

	module_path := path + "/" + req.Name
	req.GoModule = module_path

	directories, packages := baseDirectories.Clone(), slices.Clone(basePackages)

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

	if slices.Contains(req.InfraTypes, InfraTypeRedis) {
		packages = append(packages, REDIS, REDISUTIL)
	}

	if slices.Contains(req.InfraTypes, InfraTypePostgres) {
		packages = append(packages, PGX, PGUTIL)
		directories.AddSubDir([]string{INTERNAL, INFRA}, Dir{Name: MIGRATIONS})
		if req.DBLibrary == DBLibrarySqlc {
			directories.AddSubDir([]string{}, Dir{Name: CONFIG, Files: []File{{Name: SQLC, Extension: YAML}}})
			directories.AddSubDir([]string{INTERNAL, INFRA}, Dir{Name: SQLC, Directories: []Dir{{Name: QUERIES}}})
		}
	}

	for _, feature := range req.Features {
		packages = append(packages, featureToPackage[feature])
	}

	directories.Directories = append(directories.Directories,
		Dir{Name: BUILD},
		Dir{Name: DEPLOYMENTS, Directories: []Dir{{Name: LOCAL}}},
	)

	req.HeadDir = directories
	req.Packages = packages
	req.ShouldAddDb = slices.Contains(req.InfraTypes, InfraTypePostgres)
	req.ShouldAddRedis = slices.Contains(req.InfraTypes, InfraTypeRedis)
	req.ShouldAddSecret = req.AuthType != AuthTypeNone
	req.IsRest = req.ServiceType == ServiceTypeRest
	req.HasPagination = slices.Contains(req.Features, FeaturePagination)

	imports := []string{"context", req.GoModule + "/" + INTERNAL + "/" + CONFIG}
	if req.ShouldAddDb {
		imports = append(imports, PGUTIL, PGX+"/"+"pgxpool")
	}

	if req.ShouldAddRedis {
		imports = append(imports, REDISUTIL, REDIS)
	}

	req.DiImports = imports
}

func runGoTidy(dir string) {
	go func() {
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Dir = dir
		cmd.Run()
	}()
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

func installGoPackages(req CreateNewProjectRequest) error {
	stop := printer.Spin("Installing go packages")
	args := append([]string{"get"}, req.Packages...)
	cmd := exec.Command("go", args...)
	cmd.Dir = req.RootDirPath

	err := cmd.Run()
	stop(err)
	if err != nil {
		return fmt.Errorf("failed to get go packages %w", err)
	}

	return nil
}

func createDirStructure(current string, d Dir) error {
	for _, f := range d.Files {
		fileName := f.Name + "." + f.Extension
		filePath := filepath.Join(current, fileName)
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", filePath, err)
		}
		file.Close()
	}

	for _, sub := range d.Directories {
		subPath := filepath.Join(current, sub.Name)
		if err := os.MkdirAll(subPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", subPath, err)
		}
		if err := createDirStructure(subPath, sub); err != nil {
			return err
		}
	}

	return nil
}

func populateProjectStructure(req CreateNewProjectRequest) error {
	log.Info("creating project structure at", "path", req.RootDirPath)
	if err := os.MkdirAll(req.RootDirPath, 0755); err != nil {
		return fmt.Errorf("failed to create root directory %s: %w", req.RootDirPath, err)
	}

	if err := createDirStructure(req.RootDirPath, req.HeadDir); err != nil {
		return err
	}

	printer.Success("Created Project Structure")
	return nil
}

func generateCodeFromRequest(req CreateNewProjectRequest) error {
	runner := errors.ResultRunnerWithParam[CreateNewProjectRequest]{}
	runner.Do(req, renderGitIgnore)
	runner.Do(req, renderEnv)
	runner.Do(req, renderMain)
	runner.Do(req, renderBootstrap)
	runner.Do(req, renderConfig)
	runner.Do(req, renderDi)
	runner.Do(req, renderHealth)
	runner.Do(req, renderRouter)

	runner.Do(req, renderDockerfile)
	runner.Do(req, renderDockerCompose)
	runner.Do(req, renderLocalEnv)
	runner.Do(req, renderAirToml)

	if req.DBLibrary == DBLibrarySqlc {
		runner.Do(req, renderSqlcConfig)
	}

	if runner.Error != nil {
		return runner.Error
	}

	printer.Success("Generated go code")
	return nil
}

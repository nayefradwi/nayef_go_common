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
	directories, packages := baseDirectories.clone(), slices.Clone(basePackages)

	if req.ServiceType == ServiceTypeRest || req.ServiceType == ServiceTypeBoth {
		packages = append(packages, CHI, HTTPUTIL)
	}

	if req.ServiceType == ServiceTypeGrpc || req.ServiceType == ServiceTypeBoth {
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
		directories.addSubDir([]string{INFRA}, Dir{Name: MIGRATIONS})
		if req.DBLibrary == DBLibrarySqlc {
			directories.addSubDir([]string{}, Dir{Name: CONFIG, Files: []File{{Name: SQLC, Extension: YAML}}})
			directories.addSubDir([]string{INFRA}, Dir{Name: SQLC, Directories: []Dir{{Name: QUERIES}}})
		}
	}

	for _, feature := range req.Features {
		packages = append(packages, featureToPackage[feature])
	}

	req.HeadDir = directories
	req.Packages = packages

	wd, _ := os.Getwd()
	req.RootDirPath = filepath.Join(wd, req.Name)
}

func runGoTidy(dir string) {
	go func() {
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Dir = dir
		cmd.Run()
	}()
}

func createGoMod(req CreateNewProjectRequest) error {
	path := os.Getenv("GO_MOD_PATH")
	if path == "" {
		path = DEFAULT_MOD_PATH
	}

	module_path := path + "/" + req.Name
	log.Info("running go mod init at", "path", module_path)
	cmd := exec.Command("go", "mod", "init", module_path)
	cmd.Dir = req.RootDirPath

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to initialize go module at %s %w", req.RootDirPath, err)
	}

	printer.Success("Initialized Go module")
	return nil
}

func installGoPackages(req CreateNewProjectRequest) error {
	args := append([]string{"get"}, req.Packages...)
	cmd := exec.Command("go", args...)
	cmd.Dir = req.RootDirPath

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to get go packages %w", err)
	}

	printer.Success("Installed go packages")
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
	runner.Do(req, renderMain)
	runner.Do(req, renderBootstrap)
	if req.DBLibrary == DBLibrarySqlc {
		runner.Do(req, renderSqlcConfig)
	}
	if runner.Error != nil {
		return runner.Error
	}

	printer.Success("Generated go code")
	return nil
}

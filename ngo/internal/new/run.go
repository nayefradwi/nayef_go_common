package new

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/nayefradwi/nayef_go_common/errors"
)

func createGoMod(_ CreateNewProjectRequest) error {
	return nil
}

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
		directories.addSubDir([]string{INFRA}, dir{name: MIGRATIONS})
		if req.DBLibrary == DBLibrarySqlc {
			directories.addSubDir([]string{INFRA}, dir{name: SQLC})
		}
	}

	for _, feature := range req.Features {
		packages = append(packages, featureToPackage[feature])
	}

	req.headDir = directories
	req.packages = packages
}

func createDirStructure(current string, d dir) error {
	for _, f := range d.files {
		fileName := f.name + "." + f.extension
		filePath := filepath.Join(current, fileName)
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", filePath, err)
		}
		file.Close()
	}

	for _, sub := range d.directories {
		subPath := filepath.Join(current, sub.name)
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
	wd, _ := os.Getwd()
	rootDir := filepath.Join(wd, req.Name)
	return createDirStructure(rootDir, req.headDir)
}

func generateCodeFromRequest(_ CreateNewProjectRequest) error {
	return nil
}

func runGoTidy() {}

func Run() error {
	req, err := RunForm()

	if err != nil {
		return err
	}

	populateRequestDetails(req)
	if err := createGoMod(*req); err != nil {
		return err
	}

	defer runGoTidy()

	runner := errors.ResultRunnerWithParam[CreateNewProjectRequest]{}
	runner.Do(*req, populateProjectStructure)
	runner.Do(*req, generateCodeFromRequest)

	return runner.Error
}

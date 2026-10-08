package common

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v3"
)

const PROJECT_FILE = "ngo.yaml"

type Project struct {
	ServiceType          ServiceType `yaml:"service_type"`
	InfraTypes           []InfraType `yaml:"infra"`
	AuthType             AuthType    `yaml:"auth"`
	Features             []Feature   `yaml:"features"`
	Provider             string      `yaml:"provider"`
	StagingDeployment    string      `yaml:"staging_deployment"`
	ProductionDeployment string      `yaml:"production_deployment"`
}

func WriteProject(dir string, p Project) error {
	out, err := yaml.Marshal(p)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, PROJECT_FILE), out, 0644)
}

func ReadProject(dir string) (Project, error) {
	var p Project
	in, err := os.ReadFile(filepath.Join(dir, PROJECT_FILE))
	if errors.Is(err, fs.ErrNotExist) {
		return p, fmt.Errorf("%s not found in %s, run ngo add from a project created by ngo new", PROJECT_FILE, dir)
	}
	if err != nil {
		return p, err
	}

	return p, yaml.Unmarshal(in, &p)
}

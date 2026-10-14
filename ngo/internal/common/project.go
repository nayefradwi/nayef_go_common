package common

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

const PROJECT_FILE = "ngo.yaml"

type Project struct {
	ServiceType ServiceType `yaml:"service_type"`
	InfraTypes  []InfraType `yaml:"infra"`
	Auth        *Auth       `yaml:"auth,omitempty"`
	Features    []Feature   `yaml:"features"`
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
		return p, fmt.Errorf("%s not found in %s, run ngo from a project created by ngo new", PROJECT_FILE, dir)
	}
	if err != nil {
		return p, err
	}

	return p, yaml.Unmarshal(in, &p)
}

func ReadGoModule(dir string) (string, error) {
	f, err := os.Open(filepath.Join(dir, "go.mod"))
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if module, found := strings.CutPrefix(line, "module "); found {
			return strings.TrimSpace(module), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", os.ErrNotExist
}

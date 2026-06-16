package add

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/common"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/printer"
)

func RunForm() (*CreateFeatureRequest, error) {
	req := &CreateFeatureRequest{}

	form := huh.NewForm(
		huh.NewGroup(common.NameInput(&req.Name)),
		huh.NewGroup(common.InfraTypeInput(&req.InfraTypes)),
		huh.NewGroup(common.AuthInput(&req.AuthType)),
		huh.NewGroup(common.FeatureInput(&req.Features)),
	)
	if err := form.Run(); err != nil {
		printer.Error(err.Error())
		return nil, err
	}

	return setRequestPathDetails(req)
}

func setRequestPathDetails(req *CreateFeatureRequest) (*CreateFeatureRequest, error) {
	wd, err := os.Getwd()
	if err != nil {
		printer.Error(err.Error())
		return nil, err
	}
	req.RootDirPath = wd

	goModule, err := readGoModule(filepath.Join(wd, "go.mod"))
	if err != nil {
		printer.Error(err.Error())
		return nil, err
	}
	req.GoModule = goModule

	return req, nil
}

func readGoModule(path string) (string, error) {
	f, err := os.Open(path)
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

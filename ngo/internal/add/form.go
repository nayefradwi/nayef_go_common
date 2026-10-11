package add

import (
	"os"

	"github.com/charmbracelet/huh"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/common"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/printer"
)

func RunForm() (*CreateFeatureRequest, error) {
	wd, err := os.Getwd()
	if err != nil {
		printer.Error(err.Error())
		return nil, err
	}

	project, err := common.ReadProject(wd)
	if err != nil {
		printer.Error(err.Error())
		return nil, err
	}

	req := &CreateFeatureRequest{ServiceType: project.ServiceType, RootDirPath: wd}
	form := huh.NewForm(
		huh.NewGroup(common.NameInput(&req.Name)),
		huh.NewGroup(common.InfraTypeInput(&req.InfraTypes, project.InfraTypes)).
			WithHideFunc(func() bool { return len(project.InfraTypes) == 0 }),
		huh.NewGroup(common.FeatureInput(&req.Features, project.Features)).
			WithHideFunc(func() bool { return len(project.Features) == 0 }),
	)
	if err := form.Run(); err != nil {
		printer.Error(err.Error())
		return nil, err
	}

	return setGoModule(req)
}

func setGoModule(req *CreateFeatureRequest) (*CreateFeatureRequest, error) {
	wd := req.RootDirPath

	goModule, err := common.ReadGoModule(wd)
	if err != nil {
		printer.Error(err.Error())
		return nil, err
	}
	req.GoModule = goModule

	return req, nil
}

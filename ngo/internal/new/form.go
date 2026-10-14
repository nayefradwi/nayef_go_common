package new

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/common"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/printer"
)

func RunForm() (*CreateNewProjectRequest, error) {
	req := &CreateNewProjectRequest{}

	form := huh.NewForm(
		huh.NewGroup(
			common.NameInput(&req.Name),
		),
		huh.NewGroup(
			huh.
				NewSelect[common.ServiceType]().
				Title("Service Type").
				Options(
					huh.NewOption(string(common.ServiceTypeRest), common.ServiceTypeRest),
					huh.NewOption(string(common.ServiceTypeGrpc), common.ServiceTypeGrpc),
				).Value(&req.ServiceType),
		),
		huh.NewGroup(common.InfraTypeInput(&req.InfraTypes, common.AllInfraTypes)),
		huh.NewGroup(common.FeatureInput(&req.Features, common.AllFeatures)),
	)

	if err := form.Run(); err != nil {
		printer.Error(err.Error())
		return nil, err
	}

	return setRequestPathDetails(req)
}

func setRequestPathDetails(req *CreateNewProjectRequest) (*CreateNewProjectRequest, error) {
	wd, _ := os.Getwd()
	req.RootDirPath = filepath.Join(wd, req.Name)

	path := os.Getenv("GO_MOD_PATH")
	if path == "" {
		path = DEFAULT_MOD_PATH
	}

	req.GoModule = path + "/" + req.Name
	return req, os.MkdirAll(req.RootDirPath, 0755)
}

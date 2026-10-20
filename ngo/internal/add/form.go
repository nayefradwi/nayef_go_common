package add

import (
	"os"

	"github.com/charmbracelet/huh"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/common"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/printer"
	"github.com/spf13/pflag"
)

func RunForm(flags *pflag.FlagSet) (*CreateFeatureRequest, error) {
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
	groups, err := fillFromFlags(req, project, flags)
	if err != nil {
		printer.Error(err.Error())
		return nil, err
	}

	if err := huh.NewForm(groups...).Run(); err != nil {
		printer.Error(err.Error())
		return nil, err
	}

	return setGoModule(req)
}

func fillFromFlags(req *CreateFeatureRequest, project common.Project, flags *pflag.FlagSet) ([]*huh.Group, error) {
	return common.FillFromFlags(flags, []common.FlagField{
		{
			Name: common.FLAG_NAME,
			Set: func() (err error) {
				req.Name, err = common.ParseRegexFlag(flags, common.FLAG_NAME, common.NameRegex)
				return
			},
			Group: func() *huh.Group {
				return huh.NewGroup(common.NameInput(&req.Name))
			},
		},
		{
			Name: common.FLAG_INFRA_TYPES,
			Set: func() (err error) {
				req.InfraTypes, err = common.ParseListFlag(flags, common.FLAG_INFRA_TYPES, project.InfraTypes)
				return
			},
			Group: func() *huh.Group {
				if len(project.InfraTypes) == 0 {
					return nil
				}
				return huh.NewGroup(common.InfraTypeInput(&req.InfraTypes, project.InfraTypes))
			},
		},
		{
			Name: common.FLAG_FEATURES,
			Set: func() (err error) {
				req.Features, err = common.ParseListFlag(flags, common.FLAG_FEATURES, project.Features)
				return
			},
			Group: func() *huh.Group {
				if len(project.Features) == 0 {
					return nil
				}
				return huh.NewGroup(common.FeatureInput(&req.Features, project.Features))
			},
		},
	})
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

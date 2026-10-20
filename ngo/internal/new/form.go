package new

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/common"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/printer"
	"github.com/spf13/pflag"
)

func RunForm(flags *pflag.FlagSet) (*CreateNewProjectRequest, error) {
	req := &CreateNewProjectRequest{}
	groups, err := fillFromFlags(req, flags)
	if err != nil {
		printer.Error(err.Error())
		return nil, err
	}

	if err := huh.NewForm(groups...).Run(); err != nil {
		printer.Error(err.Error())
		return nil, err
	}

	return setRequestPathDetails(req)
}

func fillFromFlags(req *CreateNewProjectRequest, flags *pflag.FlagSet) ([]*huh.Group, error) {
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
			Name: common.FLAG_SERVICE_TYPE,
			Set: func() (err error) {
				req.ServiceType, err = common.ParseFlag(flags, common.FLAG_SERVICE_TYPE, common.AllServiceTypes)
				return
			},
			Group: func() *huh.Group {
				return huh.NewGroup(
					huh.NewSelect[common.ServiceType]().
						Title("Service Type").
						Options(huh.NewOptions(common.AllServiceTypes...)...).
						Value(&req.ServiceType),
				)
			},
		},
		{
			Name: common.FLAG_INFRA_TYPES,
			Set: func() (err error) {
				req.InfraTypes, err = common.ParseListFlag(flags, common.FLAG_INFRA_TYPES, common.AllInfraTypes)
				return
			},
			Group: func() *huh.Group {
				return huh.NewGroup(common.InfraTypeInput(&req.InfraTypes, common.AllInfraTypes))
			},
		},
		{
			Name: common.FLAG_FEATURES,
			Set: func() (err error) {
				req.Features, err = common.ParseListFlag(flags, common.FLAG_FEATURES, common.AllFeatures)
				return
			},
			Group: func() *huh.Group {
				return huh.NewGroup(common.FeatureInput(&req.Features, common.AllFeatures))
			},
		},
	})
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

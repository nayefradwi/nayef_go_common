package deploy

import (
	"fmt"
	"os"
	"path"

	"github.com/charmbracelet/huh"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/common"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/printer"
	"github.com/spf13/pflag"
)

func RunForm(flags *pflag.FlagSet) (*CreateDeploymentRequest, error) {
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

	goModule, err := common.ReadGoModule(wd)
	if err != nil {
		printer.Error(err.Error())
		return nil, err
	}

	req := &CreateDeploymentRequest{
		InfraTypes:  project.InfraTypes,
		Name:        path.Base(goModule),
		ServiceType: project.ServiceType,
		RootDirPath: wd,
	}
	groups, err := fillFromFlags(req, flags)
	if err != nil {
		printer.Error(err.Error())
		return nil, err
	}

	if err := huh.NewForm(groups...).Run(); err != nil {
		printer.Error(err.Error())
		return nil, err
	}

	// provider can come from a flag while size comes from the form
	if req.Provider == Vps && !req.IsSmall() {
		err := fmt.Errorf("--%s %s only works with --%s %s", FLAG_PROVIDER, PROVIDER_VPS, FLAG_SIZE, Small)
		printer.Error(err.Error())
		return nil, err
	}

	return req, nil
}

func fillFromFlags(req *CreateDeploymentRequest, flags *pflag.FlagSet) ([]*huh.Group, error) {
	return common.FillFromFlags(flags, []common.FlagField{
		{
			Name: FLAG_ENVIRONMENT,
			Set: func() (err error) {
				req.Environment, err = common.ParseRegexFlag(flags, FLAG_ENVIRONMENT, common.NameRegex)
				return
			},
			Group: func() *huh.Group {
				return huh.NewGroup(common.NameInput(&req.Environment))
			},
		},
		{
			Name: FLAG_SIZE,
			Set: func() (err error) {
				req.Size, err = common.ParseFlag(flags, FLAG_SIZE, AllSizes)
				return
			},
			Group: func() *huh.Group {
				return huh.NewGroup(DeploymentSizeInput(&req.Size))
			},
		},
		{
			Name: FLAG_PROVIDER,
			Set: func() (err error) {
				req.Provider, err = common.ParseFlag(flags, FLAG_PROVIDER, AllProviders)
				return
			},
			Group: func() *huh.Group {
				return huh.NewGroup(CloudProviderInput(&req.Provider, &req.Size))
			},
		},
	})
}

func DeploymentSizeInput(value *DeploymentSize) *huh.Select[DeploymentSize] {
	return huh.NewSelect[DeploymentSize]().
		Title("Deployment Size").
		Description("based on this the infra will be generated").
		Options(
			huh.NewOption("Small $25 - $40", Small),
			huh.NewOption("Medium $125 - $175", Medium),
			huh.NewOption("Large $300 - $500", Large),
		).
		Value(value)
}

func CloudProviderInput(value *CloudProvider, size *DeploymentSize) *huh.Select[CloudProvider] {
	options := func() []huh.Option[CloudProvider] {
		options := []huh.Option[CloudProvider]{
			huh.NewOption("AWS", Aws),
			huh.NewOption("GCP", Gcp),
			huh.NewOption("Azure", Azure),
		}
		if *size == Small {
			options = append(options, huh.NewOption("VPS", Vps))
		}
		return options
	}

	return huh.NewSelect[CloudProvider]().
		Title("Cloud Provider").
		Description("which cloud provider will be used").
		OptionsFunc(options, size).
		Value(value)
}

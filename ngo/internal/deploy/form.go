package deploy

import (
	"os"
	"path"

	"github.com/charmbracelet/huh"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/common"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/printer"
)

func RunForm() (*CreateDeploymentRequest, error) {
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
	form := huh.NewForm(
		huh.NewGroup(common.NameInput(&req.Environment)),
		huh.NewGroup(DeploymentSizeInput(&req.Size)),
		huh.NewGroup(CloudProviderInput(&req.Provider, &req.Size)),
	)

	if err := form.Run(); err != nil {
		printer.Error(err.Error())
		return nil, err
	}

	return req, nil
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

package new

import (
	"os"
	"path/filepath"
	"slices"

	"github.com/charmbracelet/huh"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/common"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/printer"
)

func deploymentOptions(provider ProviderType) []huh.Option[DeploymentType] {
	opts := []huh.Option[DeploymentType]{
		huh.NewOption(string(DeploymentTypeManual), DeploymentTypeManual),
	}
	for _, dt := range providerToDeploymentType[provider] {
		opts = append(opts, huh.NewOption(string(dt), dt))
	}
	return opts
}

func RunForm() (*CreateNewProjectRequest, error) {
	req := &CreateNewProjectRequest{DBLibrary: DBLibraryNone}

	form := huh.NewForm(
		huh.NewGroup(
			common.NameInput(&req.Name),
		),
		huh.NewGroup(
			huh.
				NewSelect[ServiceType]().
				Title("Service Type").
				Options(
					huh.NewOption(string(ServiceTypeRest), ServiceTypeRest),
					huh.NewOption(string(ServiceTypeGrpc), ServiceTypeGrpc),
				).Value(&req.ServiceType),
		),
		huh.NewGroup(common.InfraTypeInput(&req.InfraTypes)),
		huh.NewGroup(
			huh.
				NewSelect[DBLibrary]().
				Title("DB Library").
				Description("Choose the query library for PostgreSQL").
				Options(
					huh.NewOption(string(DBLibrarySqlc), DBLibrarySqlc),
					huh.NewOption(string(DBLibraryNone), DBLibraryNone),
				).
				Value(&req.DBLibrary),
		).WithHideFunc(func() bool {
			return !slices.Contains(req.InfraTypes, common.InfraTypePostgres)
		}),
		huh.NewGroup(common.AuthInput(&req.AuthType)),
		huh.NewGroup(common.FeatureInput(&req.Features)),
		huh.NewGroup(
			huh.
				NewSelect[ProviderType]().
				Title("Cloud Provider").
				Options(
					huh.NewOption(string(ProviderTypeAWS), ProviderTypeAWS),
				).Value(&req.ProviderType),
		),
		huh.NewGroup(
			huh.
				NewSelect[DeploymentType]().
				Title("Staging Deployment Type").
				OptionsFunc(func() []huh.Option[DeploymentType] {
					return deploymentOptions(req.ProviderType)
				}, &req.ProviderType).
				Value(&req.StagingDeploymentType),
		).WithHideFunc(func() bool {
			return req.ProviderType == ""
		}),
		huh.NewGroup(
			huh.
				NewSelect[DeploymentType]().
				Title("Production Deployment Type").
				OptionsFunc(func() []huh.Option[DeploymentType] {
					return deploymentOptions(req.ProviderType)
				}, &req.ProviderType).
				Value(&req.ProductionDeploymentType),
		).WithHideFunc(func() bool {
			return req.ProviderType == ""
		}),
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

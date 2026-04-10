package new

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"

	"github.com/charmbracelet/huh"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/printer"
)

var projectNameRegex = regexp.MustCompile(`^[a-z][a-z0-9]*$`)

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
			huh.
				NewInput().
				Title("Project Name").
				Placeholder("myservice").
				Description("Go package name: lowercase letters and digits only, no separators").
				Validate(func(s string) error {
					if !projectNameRegex.MatchString(s) {
						return fmt.Errorf("project name must be lowercase letters and digits only (Go package naming convention)")
					}
					return nil
				}).
				Value(&req.Name),
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
		huh.NewGroup(
			huh.
				NewMultiSelect[InfraType]().
				Title("Infra required").
				Description("based on this common modules will be imported").
				Options(
					huh.NewOption(string(InfraTypePostgres), InfraTypePostgres),
					huh.NewOption(string(InfraTypeRedis), InfraTypeRedis),
				).
				Value(&req.InfraTypes),
		),
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
			return !slices.Contains(req.InfraTypes, InfraTypePostgres)
		}),
		huh.NewGroup(
			huh.
				NewSelect[AuthType]().
				Title("Auth Type").
				Description("Choose the level of 'revokness', or None to skip").
				Options(
					huh.NewOption(string(AuthTypeJWT), AuthTypeJWT),
					huh.NewOption(string(AuthTypeRefresh), AuthTypeRefresh),
					huh.NewOption(string(AuthTypeOpaque), AuthTypeOpaque),
					huh.NewOption(string(AuthTypeNone), AuthTypeNone),
				).Value(&req.AuthType),
		),
		huh.NewGroup(
			huh.
				NewMultiSelect[Feature]().
				Title("Additional Features").
				Description("based on this common modules will be imported").
				Options(
					huh.NewOption(string(FeatureLocking), FeatureLocking),
					huh.NewOption(string(FeatureOtp), FeatureOtp),
					huh.NewOption(string(FeaturePagination), FeaturePagination),
				).
				Value(&req.Features),
		),
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

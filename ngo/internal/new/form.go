package new

import (
	"fmt"
	"regexp"
	"slices"

	"github.com/charmbracelet/huh"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/printer"
)

var projectNameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)

func RunForm() (*CreateNewProjectRequest, error) {
	req := &CreateNewProjectRequest{DBLibrary: DBLibraryNone}

	form := huh.NewForm(
		huh.NewGroup(
			huh.
				NewInput().
				Title("Project Name").
				Placeholder("my-service").
				Description("Must start with a letter, then letters/digits/underscores/hyphens").
				Validate(func(s string) error {
					if !projectNameRegex.MatchString(s) {
						return fmt.Errorf("project name must start with a letter and contain only letters, digits, underscores, or hyphens")
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
	)

	if err := form.Run(); err != nil {
		printer.Error(err.Error())
		return nil, err
	}

	return req, nil
}

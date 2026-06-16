package common

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

func NameInput(value *string) *huh.Input {
	return huh.NewInput().
		Title("Name").
		Placeholder("myservice").
		Description("Name: lowercase letters and digits only, no separators").
		Validate(func(s string) error {
			if !NameRegex.MatchString(s) {
				return fmt.Errorf("name must be lowercase letters and digits only (Go package naming convention)")
			}
			return nil
		}).
		Value(value)
}

func AuthInput(value *AuthType) *huh.Select[AuthType] {
	return huh.
		NewSelect[AuthType]().
		Title("Auth Type").
		Description("Choose the level of revocability, or Something Else to skip").
		Options(
			huh.NewOption(string(AuthTypeJWT), AuthTypeJWT),
			huh.NewOption(string(AuthTypeRefresh), AuthTypeRefresh),
			huh.NewOption(string(AuthTypeOpaque), AuthTypeOpaque),
			huh.NewOption(string(AuthTypeNone), AuthTypeNone),
		).Value(value)
}

func InfraTypeInput(value *[]InfraType) *huh.MultiSelect[InfraType] {
	return huh.
		NewMultiSelect[InfraType]().
		Title("Infra required").
		Description("based on this common modules will be imported").
		Options(
			huh.NewOption(string(InfraTypePostgres), InfraTypePostgres),
			huh.NewOption(string(InfraTypeRedis), InfraTypeRedis),
		).
		Value(value)
}

func FeatureInput(value *[]Feature) *huh.MultiSelect[Feature] {
	return huh.
		NewMultiSelect[Feature]().
		Title("Additional Features").
		Description("based on this common modules will be imported").
		Options(
			huh.NewOption(string(FeatureLocking), FeatureLocking),
			huh.NewOption(string(FeatureOtp), FeatureOtp),
			huh.NewOption(string(FeaturePagination), FeaturePagination),
		).
		Value(value)
}

package common

import (
	"github.com/charmbracelet/huh"
)

func NameInput(value *string) *huh.Input {
	return huh.NewInput().
		Title("Name").
		Placeholder("myservice").
		Description("Name: lowercase letters and digits only, no separators").
		Validate(ValidateName).
		Value(value)
}

func InfraTypeInput(value *[]InfraType, options []InfraType) *huh.MultiSelect[InfraType] {
	return huh.
		NewMultiSelect[InfraType]().
		Title("Infra required").
		Description("based on this common modules will be imported").
		Options(huh.NewOptions(options...)...).
		Value(value)
}

func FeatureInput(value *[]Feature, options []Feature) *huh.MultiSelect[Feature] {
	return huh.
		NewMultiSelect[Feature]().
		Title("Additional Features").
		Description("based on this common modules will be imported").
		Options(huh.NewOptions(options...)...).
		Value(value)
}

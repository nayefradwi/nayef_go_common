package common

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/pflag"
)

const (
	FLAG_NAME         = "name"
	FLAG_SERVICE_TYPE = "service-type"
	FLAG_INFRA_TYPES  = "infra-types"
	FLAG_FEATURES     = "features"
)

type FlagValueSetter[T any] func(value T) error

type FlagField struct {
	Name  string
	Set   func() error
	Group func() *huh.Group
}

func FillFromFlags(flags *pflag.FlagSet, fields []FlagField) ([]*huh.Group, error) {
	var groups []*huh.Group
	for _, f := range fields {
		if flags.Changed(f.Name) {
			if err := f.Set(); err != nil {
				return nil, err
			}
			continue
		}
		if g := f.Group(); g != nil {
			groups = append(groups, g)
		}
	}
	return groups, nil
}

func ParseFlag[T ~string](flags *pflag.FlagSet, name string, allowed []T) (T, error) {
	value, err := flags.GetString(name)
	if err != nil {
		return "", err
	}

	return parseValue(name, value, allowed)
}

func ParseListFlag[T ~string](flags *pflag.FlagSet, name string, allowed []T) ([]T, error) {
	values, err := flags.GetStringSlice(name)
	if err != nil {
		return nil, err
	}

	parsed := make([]T, 0, len(values))
	for _, value := range values {
		v, err := parseValue(name, value, allowed)
		if err != nil {
			return nil, err
		}
		parsed = append(parsed, v)
	}

	return parsed, nil
}

func ParseRegexFlag(flags *pflag.FlagSet, name string, regex *regexp.Regexp) (string, error) {
	value, err := flags.GetString(name)
	if err != nil {
		return "", err
	}

	if !regex.MatchString(value) {
		return "", fmt.Errorf("invalid --%s %q, must match %s", name, value, regex)
	}

	return value, nil
}

func ValidateName(s string) error {
	if !NameRegex.MatchString(s) {
		return fmt.Errorf("name must be lowercase letters and digits only (Go package naming convention)")
	}
	return nil
}

func FlagValues[T ~string](allowed []T) string {
	names := make([]string, len(allowed))
	for i, a := range allowed {
		names[i] = strings.ToLower(string(a))
	}
	return strings.Join(names, ", ")
}

func parseValue[T ~string](name, value string, allowed []T) (T, error) {
	for _, a := range allowed {
		if strings.EqualFold(string(a), strings.TrimSpace(value)) {
			return a, nil
		}
	}

	return "", fmt.Errorf("invalid --%s %q, want one of: %s", name, value, FlagValues(allowed))
}

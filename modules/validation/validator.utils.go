package validation

import (
	"net/url"
	"regexp"

	"github.com/nayefradwi/nayef_go_common/core"
)

const (
	INVALID_DATA_TYPE = "INVALID_DATA_TYPE"
)

func NotEmptyString(opts ValidateOption) ValidationFunc {
	vf := func(opts ValidateOption) core.ErrorDetails {
		str, ok := opts.Data.(string)
		if !ok {
			return core.ErrorDetails{Field: opts.Field, Message: "Invalid data type", Code: INVALID_DATA_TYPE}
		}

		if str == "" {
			return core.ErrorDetails{Field: opts.Field, Message: opts.Message}
		}

		return core.ErrorDetails{}
	}

	return ValidationFunc{Opts: opts, fn: vf}
}

func NotEmptySlice[T any](opts ValidateOption) ValidationFunc {
	vf := func(opts ValidateOption) core.ErrorDetails {
		slice, ok := opts.Data.([]T)
		if !ok {
			return core.ErrorDetails{Field: opts.Field, Message: "Invalid data type", Code: INVALID_DATA_TYPE}
		}

		if len(slice) == 0 {
			return core.ErrorDetails{Field: opts.Field, Message: opts.Message}
		}

		return core.ErrorDetails{}
	}

	return ValidationFunc{Opts: opts, fn: vf}
}

func NotEmptyMap[K comparable, V any](opts ValidateOption) ValidationFunc {

	vf := func(opts ValidateOption) core.ErrorDetails {
		if opts.Data == nil {
			return core.ErrorDetails{Field: opts.Field, Message: opts.Message}
		}

		m, ok := opts.Data.(map[K]V)
		if !ok {
			return core.ErrorDetails{Field: opts.Field, Message: "Invalid data type", Code: INVALID_DATA_TYPE}
		}

		if len(m) == 0 {
			return core.ErrorDetails{Field: opts.Field, Message: opts.Message}
		}

		return core.ErrorDetails{}
	}

	return ValidationFunc{Opts: opts, fn: vf}
}

func IsUrl(opts ValidateOption) ValidationFunc {
	vf := func(opts ValidateOption) core.ErrorDetails {
		str, ok := opts.Data.(string)
		if !ok {
			return core.ErrorDetails{Field: opts.Field, Message: "Invalid data type", Code: INVALID_DATA_TYPE}
		}

		url, err := url.Parse(str)
		if err != nil || url.Host == "" || url.Scheme == "" {
			return core.ErrorDetails{Field: opts.Field, Message: opts.Message}
		}

		return core.ErrorDetails{}
	}

	return ValidationFunc{Opts: opts, fn: vf}
}

func IsEmail(opts ValidateOption) ValidationFunc {
	vf := func(opts ValidateOption) core.ErrorDetails {
		str, ok := opts.Data.(string)
		if !ok {
			return core.ErrorDetails{Field: opts.Field, Message: "Invalid data type", Code: INVALID_DATA_TYPE}
		}

		emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		matched, err := regexp.MatchString(emailRegex, str)
		if err != nil || !matched {
			return core.ErrorDetails{Field: opts.Field, Message: opts.Message}
		}

		return core.ErrorDetails{}
	}

	return ValidationFunc{Opts: opts, fn: vf}
}

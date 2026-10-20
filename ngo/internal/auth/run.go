package auth

import (
	"github.com/nayefradwi/nayef_go_common/ngo/internal/printer"
	"github.com/spf13/pflag"
)

func Run(flags *pflag.FlagSet) error {
	req, err := RunForm(flags)
	if err != nil {
		return err
	}

	err = scaffold(*req)
	if err != nil {
		printer.Error(err.Error())
	}

	return err
}

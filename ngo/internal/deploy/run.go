package deploy

import "github.com/spf13/pflag"

func Run(flags *pflag.FlagSet) error {
	req, err := RunForm(flags)
	if err != nil {
		return err
	}

	return generateDeployment(*req)
}

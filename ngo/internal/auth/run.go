package auth

import "github.com/nayefradwi/nayef_go_common/ngo/internal/printer"

func Run() error {
	req, err := RunForm()
	if err != nil {
		return err
	}

	err = scaffold(*req)
	if err != nil {
		printer.Error(err.Error())
	}

	return err
}

package new

import (
	"github.com/nayefradwi/nayef_go_common/errors"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/printer"
)

func Run() error {
	req, err := RunForm()

	if err != nil {
		return err
	}

	populateRequestDetails(req)
	defer runGoTidy(req.RootDirPath)

	runner := errors.ResultRunnerWithParam[CreateNewProjectRequest]{}
	runner.Do(*req, populateProjectStructure)
	runner.Do(*req, createGoMod)
	runner.Do(*req, installGoPackages)
	runner.Do(*req, generateCodeFromRequest)

	if runner.Error != nil {
		printer.Error(runner.Error.Error())
	}

	return runner.Error
}

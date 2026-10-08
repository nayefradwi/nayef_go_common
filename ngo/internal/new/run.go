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

	err = scaffold(*req)
	if err != nil {
		printer.Error(err.Error())
	}

	return err
}

func scaffold(req CreateNewProjectRequest) error {
	runner := errors.ResultRunnerWithParam[CreateNewProjectRequest]{}
	runner.Do(req, checkBufInstalled)
	runner.Do(req, createGoMod)
	runner.Do(req, installGoPackages)
	runner.Do(req, installGoTools)
	runner.Do(req, generateCodeFromRequest)
	runner.Do(req, writeProjectFile)
	runner.Do(req, runBufGenerate)
	runner.Do(req, runGoFmt)
	runner.Do(req, runGoTidy)
	runner.Do(req, generateSSHKeys)
	return runner.Error
}

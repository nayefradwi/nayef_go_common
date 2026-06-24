package add

import "github.com/nayefradwi/nayef_go_common/errors"

func generateServiceClass(req CreateFeatureRequest) error {
	runner := errors.ResultRunnerWithParam[CreateFeatureRequest]{}
	runner.Do(req, renderService)
	runner.Do(req, renderHandler)
	return runner.Error
}

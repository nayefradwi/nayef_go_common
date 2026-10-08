package add

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nayefradwi/nayef_go_common/errors"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/common"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/printer"
)

func generateServiceClass(req CreateFeatureRequest) error {
	runner := errors.ResultRunnerWithParam[CreateFeatureRequest]{}
	runner.Do(req, checkFeatureNotExists)
	runner.Do(req, checkBufInstalled)
	runner.Do(req, renderService)
	runner.Do(req, renderHandler)
	runner.Do(req, renderProto)
	runner.Do(req, runBufGenerate)
	runner.Do(req, printWiring)
	return runner.Error
}

func checkFeatureNotExists(req CreateFeatureRequest) error {
	dir := filepath.Join(req.RootDirPath, INTERNAL, req.Name)
	if _, err := os.Stat(dir); err == nil {
		return fmt.Errorf("%s already exists, pick another name", dir)
	}

	return nil
}

func checkBufInstalled(req CreateFeatureRequest) error {
	if !req.IsGrpc() {
		return nil
	}

	return common.CheckBufInstalled()
}

func runBufGenerate(req CreateFeatureRequest) error {
	if !req.IsGrpc() {
		return nil
	}

	return common.RunBufGenerate(req.RootDirPath)
}

func printWiring(req CreateFeatureRequest) error {
	v := newHandlerView(req)
	svc := fmt.Sprintf("%s.New%sService(%s)", v.Package, v.Name, serviceArgs(req))
	handler := fmt.Sprintf("%s.New%sHandler(%s)", v.Package, v.Name, svc)

	printer.Success("added " + req.Name)
	if !req.IsGrpc() {
		printer.Info("add to setup() in cmd/api/router.go:")
		printer.Info(fmt.Sprintf("ar.mux.Mount(\"/%s\", %s.Routes())", v.Package, handler))
		return nil
	}

	printer.Info("add to setup() in cmd/api/router.go:")
	printer.Info(fmt.Sprintf("ar.mux.Handle(%sv1connect.New%sServiceHandler(%s, opts))", v.Package, v.Name, handler))
	printer.Info("add to NewStaticChecker in internal/health/handler.go:")
	printer.Info(fmt.Sprintf("%sv1connect.%sServiceName", v.Package, v.Name))
	return nil
}

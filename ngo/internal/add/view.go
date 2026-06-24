package add

import (
	"strings"

	"github.com/nayefradwi/nayef_go_common/ngo/internal/common"
)

type ServiceView struct {
	Name             string
	Package          string
	GoModule         string
	Imports          []string
	ShouldAddDb      bool
	ShouldAddRedis   bool
	ShouldAddLocking bool
	ShouldAddOtp     bool
}

func newServiceView(req CreateFeatureRequest) ServiceView {
	imports := []string{}
	if req.HasPostgres() {
		imports = append(imports, PGXPOOL, req.GoModule+"/"+DB_GEN_SUBPATH)
	}

	if req.HasRedis() {
		imports = append(imports, REDIS)
	}

	if req.HasFeature(common.FeatureLocking) {
		imports = append(imports, LOCKING)
	}

	if req.HasFeature(common.FeatureOtp) {
		imports = append(imports, OTP)
	}

	return ServiceView{
		Name:             strings.ToUpper(req.Name[:1]) + req.Name[1:],
		Package:          req.Name,
		GoModule:         req.GoModule,
		Imports:          imports,
		ShouldAddDb:      req.HasPostgres(),
		ShouldAddRedis:   req.HasRedis(),
		ShouldAddLocking: req.HasFeature(common.FeatureLocking),
		ShouldAddOtp:     req.HasFeature(common.FeatureOtp),
	}
}

type HandlerView struct {
	Name    string
	Package string
}

func newHandlerView(req CreateFeatureRequest) HandlerView {
	return HandlerView{
		Name:    strings.ToUpper(req.Name[:1]) + req.Name[1:],
		Package: req.Name,
	}
}

package add

import (
	"embed"
	"path/filepath"

	"github.com/nayefradwi/nayef_go_common/ngo/internal/common"
)

//go:embed templates/*
var templatesFs embed.FS

func renderService(req CreateFeatureRequest) error {
	filePath := filepath.Join(req.RootDirPath, INTERNAL, req.Name, SERVICE+"."+GO)
	return common.RenderToFile(templatesFs, TMPL_SERVICE, filePath, newServiceView(req))
}

func renderHandler(req CreateFeatureRequest) error {
	filePath := filepath.Join(req.RootDirPath, INTERNAL, req.Name, HANDLER+"."+GO)
	tmpl := TMPL_HANDLER
	if req.IsGrpc() {
		tmpl = TMPL_GRPC_HANDLER
	}

	return common.RenderToFile(templatesFs, tmpl, filePath, newHandlerView(req))
}

func renderProto(req CreateFeatureRequest) error {
	if !req.IsGrpc() {
		return nil
	}

	filePath := filepath.Join(req.RootDirPath, PROTO, req.Name, V1, req.Name+"."+PROTO)
	return common.RenderToFile(templatesFs, TMPL_PROTO, filePath, newHandlerView(req))
}

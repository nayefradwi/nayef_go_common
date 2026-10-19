package auth

import (
	"embed"
	"path/filepath"

	"github.com/nayefradwi/nayef_go_common/ngo/internal/common"
)

//go:embed templates/*
var templatesFs embed.FS

func authDir(req CreateAuthRequest) string {
	return filepath.Join(req.RootDirPath, INTERNAL, AUTH)
}

// password + opaque + no MFA needs no secrets, so there is no config to read
func renderConfig(req CreateAuthRequest) error {
	if len(envKeys(req)) == 0 {
		return nil
	}

	return common.RenderToFile(templatesFs, TMPL_CONFIG, filepath.Join(authDir(req), "config.go"), newAuthView(req))
}

func renderModule(req CreateAuthRequest) error {
	handler := TMPL_HANDLER
	if req.IsGrpc() {
		handler = TMPL_GRPC_HANDLER
	}

	files := map[string]string{
		TMPL_ERRORS:   "errors.go",
		TMPL_REQUEST:  "request.go",
		TMPL_RESPONSE: "response.go",
		TMPL_SERVICE:  "service.go",
		handler:       "handler.go",
		TMPL_TODO:     "TODO.md",
	}
	for tmpl, name := range files {
		if err := common.RenderToFile(templatesFs, tmpl, filepath.Join(authDir(req), name), newAuthView(req)); err != nil {
			return err
		}
	}

	return nil
}

func queriesPath(req CreateAuthRequest) string {
	return filepath.Join(req.RootDirPath, INTERNAL, INFRA, SQLC, QUERIES, "users.sql")
}

func renderQueries(req CreateAuthRequest) error {
	return common.RenderToFile(templatesFs, TMPL_QUERIES, queriesPath(req), newAuthView(req))
}

func renderProto(req CreateAuthRequest) error {
	if !req.IsGrpc() {
		return nil
	}

	filePath := filepath.Join(req.RootDirPath, PROTO, AUTH, V1, AUTH+"."+PROTO)
	return common.RenderToFile(templatesFs, TMPL_PROTO, filePath, newAuthView(req))
}

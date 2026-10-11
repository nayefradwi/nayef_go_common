package auth

import (
	"bytes"
	"embed"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed templates/*
var templatesFs embed.FS

func renderToString[T any](tmplName string, input T) (string, error) {
	return renderNamed(tmplName, tmplName, input)
}

func renderNamed[T any](tmplName, name string, input T) (string, error) {
	tmpl, err := template.ParseFS(templatesFs, "templates/"+tmplName)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, input); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func renderToFile[T any](tmplName, filePath string, input T) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}

	out, err := renderToString(tmplName, input)
	if err != nil {
		return err
	}

	content := []byte(out)
	if strings.HasSuffix(filePath, ".go") {
		content, err = format.Source(content)
		if err != nil {
			return err
		}
	}

	return os.WriteFile(filePath, content, 0644)
}

func authDir(req CreateAuthRequest) string {
	return filepath.Join(req.RootDirPath, INTERNAL, AUTH)
}

// password + opaque + no MFA needs no secrets, so there is no config to read
func renderConfig(req CreateAuthRequest) error {
	if len(envKeys(req)) == 0 {
		return nil
	}

	return renderToFile(TMPL_CONFIG, filepath.Join(authDir(req), "config.go"), newAuthView(req))
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
	}
	for tmpl, name := range files {
		if err := renderToFile(tmpl, filepath.Join(authDir(req), name), newAuthView(req)); err != nil {
			return err
		}
	}

	return nil
}

func queriesPath(req CreateAuthRequest) string {
	return filepath.Join(req.RootDirPath, INTERNAL, INFRA, SQLC, QUERIES, "users.sql")
}

func renderQueries(req CreateAuthRequest) error {
	return renderToFile(TMPL_QUERIES, queriesPath(req), newAuthView(req))
}

func renderProto(req CreateAuthRequest) error {
	if !req.IsGrpc() {
		return nil
	}

	filePath := filepath.Join(req.RootDirPath, PROTO, AUTH, V1, AUTH+"."+PROTO)
	return renderToFile(TMPL_PROTO, filePath, newAuthView(req))
}

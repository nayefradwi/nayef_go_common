package add

import (
	"embed"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed templates/*
var templatesFs embed.FS

func renderToFile[T any](tmplName, filePath string, input T) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}

	tmpl, err := template.ParseFS(templatesFs, "templates/"+tmplName)
	if err != nil {
		return err
	}

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, input)
}

func renderService(req CreateFeatureRequest) error {
	filePath := filepath.Join(req.RootDirPath, INTERNAL, req.Name, SERVICE+"."+GO)
	return renderToFile(TMPL_SERVICE, filePath, newServiceView(req))
}

func renderHandler(req CreateFeatureRequest) error {
	filePath := filepath.Join(req.RootDirPath, INTERNAL, req.Name, HANDLER+"."+GO)
	tmpl := TMPL_HANDLER
	if req.IsGrpc() {
		tmpl = TMPL_GRPC_HANDLER
	}

	return renderToFile(tmpl, filePath, newHandlerView(req))
}

func renderProto(req CreateFeatureRequest) error {
	if !req.IsGrpc() {
		return nil
	}

	filePath := filepath.Join(req.RootDirPath, PROTO, req.Name, V1, req.Name+"."+PROTO)
	return renderToFile(TMPL_PROTO, filePath, newHandlerView(req))
}

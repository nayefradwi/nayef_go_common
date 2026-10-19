package common

import (
	"bytes"
	"go/format"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func RenderToString[T any](fsys fs.FS, tmplName string, input T) (string, error) {
	tmpl, err := template.ParseFS(fsys, "templates/"+tmplName)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, input); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func RenderToFile[T any](fsys fs.FS, tmplName, filePath string, input T) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}

	out, err := RenderToString(fsys, tmplName, input)
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

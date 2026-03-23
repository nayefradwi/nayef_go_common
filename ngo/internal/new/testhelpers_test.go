package new

import (
	"bytes"
	"go/format"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func normalizeGoCode(t *testing.T, src string) string {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	require.NoError(t, err, "failed to parse Go source:\n%s", src)
	var buf bytes.Buffer
	require.NoError(t, format.Node(&buf, fset, f))
	return buf.String()
}

func assertGoCodeEqual(t *testing.T, expected, actual string) {
	t.Helper()
	assert.Equal(t, normalizeGoCode(t, expected), normalizeGoCode(t, actual))
}

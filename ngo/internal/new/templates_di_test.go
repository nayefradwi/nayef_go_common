package new

import (
	"bytes"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// normalizeGoCode parses src as a Go source file and re-formats it with
// go/format, producing a canonical representation that is insensitive to
// whitespace, indentation, and blank-line differences.
func normalizeGoCode(t *testing.T, src string) string {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	require.NoError(t, err, "failed to parse Go source:\n%s", src)
	var buf bytes.Buffer
	require.NoError(t, format.Node(&buf, fset, f))
	return buf.String()
}

// assertGoCodeEqual compares two Go source strings after normalizing both
// through go/format, so formatting differences do not cause false failures.
func assertGoCodeEqual(t *testing.T, expected, actual string) {
	t.Helper()
	assert.Equal(t, normalizeGoCode(t, expected), normalizeGoCode(t, actual))
}

func setupDiDir(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, INTERNAL, DI), 0o755))
	return root
}

func readDiFile(t *testing.T, root string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(root, INTERNAL, DI, DI+"."+GO))
	require.NoError(t, err)
	return string(content)
}

func TestRenderDi_NoDbNoRedis(t *testing.T) {
	root := setupDiDir(t)
	req := CreateNewProjectRequest{
		RootDirPath:    root,
		ShouldAddDb:    false,
		ShouldAddRedis: false,
		DiImports:      []string{"context"},
	}

	require.NoError(t, renderDi(req))

	expected := `package di

import (
	"context"
)

type Di struct{}

func RegisterServices(ctx context.Context, config config.Config) *Di {
	di := Di{}
	return &di
}

func (d *Di) Dispose() {}
`
	assertGoCodeEqual(t, expected, readDiFile(t, root))
}

func TestRenderDi_WithDbOnly(t *testing.T) {
	root := setupDiDir(t)
	req := CreateNewProjectRequest{
		RootDirPath:    root,
		ShouldAddDb:    true,
		ShouldAddRedis: false,
		DiImports:      []string{"context", "github.com/jackc/pgx/v5/pgxpool"},
	}

	require.NoError(t, renderDi(req))

	expected := `package di

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Di struct {
	Pool *pgxpool.Pool
}

func RegisterServices(ctx context.Context, config config.Config) *Di {
	di := Di{}
	di.connectToDb(ctx, config.DatabaseUrl)
	return &di
}

func (d *Di) Dispose() {
	d.Pool.Close()
}

func (d *Di) connectToDb(ctx context.Context, connectionUrl string) {
	d.Pool = pgutil.ConnectToPostgres(ctx, connectionUrl)
}
`
	assertGoCodeEqual(t, expected, readDiFile(t, root))
}

func TestRenderDi_WithRedisOnly(t *testing.T) {
	root := setupDiDir(t)
	req := CreateNewProjectRequest{
		RootDirPath:    root,
		ShouldAddDb:    false,
		ShouldAddRedis: true,
		DiImports:      []string{"context", "github.com/redis/go-redis/v9"},
	}

	require.NoError(t, renderDi(req))

	expected := `package di

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type Di struct {
	Redis *redis.Client
}

func RegisterServices(ctx context.Context, config config.Config) *Di {
	di := Di{}
	di.connectToRedis(ctx, config.RedisUrl)
	return &di
}

func (d *Di) Dispose() {
	d.Redis.Close()
}

func (d *Di) connectToRedis(ctx context.Context, connectionUrl string) {
	d.Redis = redisutil.ConnectToRedis(ctx, connectionUrl)
}
`
	assertGoCodeEqual(t, expected, readDiFile(t, root))
}

func TestRenderDi_WithDbAndRedis(t *testing.T) {
	root := setupDiDir(t)
	req := CreateNewProjectRequest{
		RootDirPath:    root,
		ShouldAddDb:    true,
		ShouldAddRedis: true,
		DiImports: []string{
			"context",
			"github.com/jackc/pgx/v5/pgxpool",
			"github.com/redis/go-redis/v9",
		},
	}

	require.NoError(t, renderDi(req))

	expected := `package di

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Di struct {
	Pool  *pgxpool.Pool
	Redis *redis.Client
}

func RegisterServices(ctx context.Context, config config.Config) *Di {
	di := Di{}
	di.connectToDb(ctx, config.DatabaseUrl)
	di.connectToRedis(ctx, config.RedisUrl)
	return &di
}

func (d *Di) Dispose() {
	d.Pool.Close()
	d.Redis.Close()
}

func (d *Di) connectToDb(ctx context.Context, connectionUrl string) {
	d.Pool = pgutil.ConnectToPostgres(ctx, connectionUrl)
}

func (d *Di) connectToRedis(ctx context.Context, connectionUrl string) {
	d.Redis = redisutil.ConnectToRedis(ctx, connectionUrl)
}
`
	assertGoCodeEqual(t, expected, readDiFile(t, root))
}

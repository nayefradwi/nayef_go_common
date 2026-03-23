package new

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupDiDir(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	diDir := filepath.Join(root, INTERNAL, DI)
	require.NoError(t, os.MkdirAll(diDir, 0o755))
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

	err := renderDi(req)
	require.NoError(t, err)

	content := readDiFile(t, root)
	assert.Contains(t, content, "package di")
	assert.Contains(t, content, "type Di struct {")
	assert.Contains(t, content, "func RegisterServices(")
	assert.Contains(t, content, "func (d *Di) Dispose() {")

	// No db or redis fields/methods
	assert.NotContains(t, content, "Pool *pgxpool.Pool")
	assert.NotContains(t, content, "Redis *redis.Client")
	assert.NotContains(t, content, "connectToDb")
	assert.NotContains(t, content, "connectToRedis")
	assert.NotContains(t, content, "d.Pool.Close()")
	assert.NotContains(t, content, "d.Redis.Close()")
}

func TestRenderDi_WithDbOnly(t *testing.T) {
	root := setupDiDir(t)
	req := CreateNewProjectRequest{
		RootDirPath:    root,
		ShouldAddDb:    true,
		ShouldAddRedis: false,
		DiImports:      []string{"context", "github.com/jackc/pgx/v5/pgxpool"},
	}

	err := renderDi(req)
	require.NoError(t, err)

	content := readDiFile(t, root)
	assert.Contains(t, content, "Pool *pgxpool.Pool")
	assert.Contains(t, content, "di.connectToDb(ctx, config.DatabaseUrl)")
	assert.Contains(t, content, "d.Pool.Close()")
	assert.Contains(t, content, "func (d *Di) connectToDb(ctx context.Context, connectionUrl string)")
	assert.Contains(t, content, "pgutil.ConnectToPostgres(ctx, connectionUrl)")

	// No redis-related content
	assert.NotContains(t, content, "Redis *redis.Client")
	assert.NotContains(t, content, "connectToRedis")
	assert.NotContains(t, content, "d.Redis.Close()")
}

func TestRenderDi_WithRedisOnly(t *testing.T) {
	root := setupDiDir(t)
	req := CreateNewProjectRequest{
		RootDirPath:    root,
		ShouldAddDb:    false,
		ShouldAddRedis: true,
		DiImports:      []string{"context", "github.com/redis/go-redis/v9"},
	}

	err := renderDi(req)
	require.NoError(t, err)

	content := readDiFile(t, root)
	assert.Contains(t, content, "Redis *redis.Client")
	assert.Contains(t, content, "di.connectToRedis(ctx, config.RedisUrl)")
	assert.Contains(t, content, "d.Redis.Close()")
	assert.Contains(t, content, "func (d *Di) connectToRedis(ctx context.Context, connectionUrl string)")
	assert.Contains(t, content, "redisutil.ConnectToRedis(ctx, connectionUrl)")

	// No db-related content
	assert.NotContains(t, content, "Pool *pgxpool.Pool")
	assert.NotContains(t, content, "connectToDb")
	assert.NotContains(t, content, "d.Pool.Close()")
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

	err := renderDi(req)
	require.NoError(t, err)

	content := readDiFile(t, root)
	// Both db and redis present
	assert.Contains(t, content, "Pool *pgxpool.Pool")
	assert.Contains(t, content, "Redis *redis.Client")
	assert.Contains(t, content, "di.connectToDb(ctx, config.DatabaseUrl)")
	assert.Contains(t, content, "di.connectToRedis(ctx, config.RedisUrl)")
	assert.Contains(t, content, "d.Pool.Close()")
	assert.Contains(t, content, "d.Redis.Close()")
	assert.Contains(t, content, "func (d *Di) connectToDb(ctx context.Context, connectionUrl string)")
	assert.Contains(t, content, "func (d *Di) connectToRedis(ctx context.Context, connectionUrl string)")
}

func TestRenderDi_ImportsRendered(t *testing.T) {
	root := setupDiDir(t)
	imports := []string{
		"context",
		"github.com/jackc/pgx/v5/pgxpool",
		"github.com/nayefradwi/nayef_go_common/pgutil",
	}
	req := CreateNewProjectRequest{
		RootDirPath:    root,
		ShouldAddDb:    true,
		ShouldAddRedis: false,
		DiImports:      imports,
	}

	err := renderDi(req)
	require.NoError(t, err)

	content := readDiFile(t, root)
	for _, imp := range imports {
		assert.Contains(t, content, `"`+imp+`"`, "expected import %q to be present", imp)
	}
}

func TestRenderDi_EmptyImports(t *testing.T) {
	root := setupDiDir(t)
	req := CreateNewProjectRequest{
		RootDirPath:    root,
		ShouldAddDb:    false,
		ShouldAddRedis: false,
		DiImports:      []string{},
	}

	err := renderDi(req)
	require.NoError(t, err)

	content := readDiFile(t, root)
	// import block should be present but empty
	assert.Contains(t, content, "import (")
	assert.Contains(t, content, ")")
}

func TestRenderDi_FileCreatedAtCorrectPath(t *testing.T) {
	root := setupDiDir(t)
	req := CreateNewProjectRequest{
		RootDirPath:    root,
		ShouldAddDb:    false,
		ShouldAddRedis: false,
		DiImports:      []string{},
	}

	err := renderDi(req)
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(root, INTERNAL, DI, DI+"."+GO))
}

func TestRenderDi_RegisterServicesReturnsPointer(t *testing.T) {
	root := setupDiDir(t)
	req := CreateNewProjectRequest{
		RootDirPath:    root,
		ShouldAddDb:    false,
		ShouldAddRedis: false,
		DiImports:      []string{"context"},
	}

	err := renderDi(req)
	require.NoError(t, err)

	content := readDiFile(t, root)
	assert.True(t,
		strings.Contains(content, "return &di"),
		"RegisterServices should return a pointer to Di",
	)
}

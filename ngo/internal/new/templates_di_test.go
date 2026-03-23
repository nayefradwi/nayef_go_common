package new

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

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

type Di struct {
}

func RegisterServices(ctx context.Context, config config.Config) *Di {
	di := Di{}
	return &di
}

func (d *Di) Dispose() {
}
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

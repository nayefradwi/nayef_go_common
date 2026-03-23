package new

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateDirStructure_EmptyDir(t *testing.T) {
	root := t.TempDir()
	err := createDirStructure(root, Dir{})
	require.NoError(t, err)
}

func TestCreateDirStructure_SingleNestedDir(t *testing.T) {
	root := t.TempDir()
	dir := Dir{
		Directories: []Dir{
			{Name: "infra"},
		},
	}
	err := createDirStructure(root, dir)
	require.NoError(t, err)
	assert.DirExists(t, filepath.Join(root, "infra"))
}

func TestCreateDirStructure_DeeplyNestedDirs(t *testing.T) {
	root := t.TempDir()
	dir := Dir{
		Directories: []Dir{
			{
				Name: "internal",
				Directories: []Dir{
					{
						Name: "infra",
						Directories: []Dir{
							{Name: "migrations"},
						},
					},
				},
			},
		},
	}
	err := createDirStructure(root, dir)
	require.NoError(t, err)
	assert.DirExists(t, filepath.Join(root, "internal"))
	assert.DirExists(t, filepath.Join(root, "internal", "infra"))
	assert.DirExists(t, filepath.Join(root, "internal", "infra", "migrations"))
}

func TestCreateDirStructure_WithFiles(t *testing.T) {
	root := t.TempDir()
	dir := Dir{
		Files: []File{
			{Name: "main", Extension: "go"},
		},
	}
	err := createDirStructure(root, dir)
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(root, "main.go"))
}

func TestCreateDirStructure_MixedFilesAndSubdirs(t *testing.T) {
	root := t.TempDir()
	dir := Dir{
		Files: []File{
			{Name: "config", Extension: "go"},
		},
		Directories: []Dir{
			{
				Name:  "health",
				Files: []File{{Name: "handler", Extension: "go"}},
			},
		},
	}
	err := createDirStructure(root, dir)
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(root, "config.go"))
	assert.DirExists(t, filepath.Join(root, "health"))
	assert.FileExists(t, filepath.Join(root, "health", "handler.go"))
}

func TestCreateDirStructure_PostgresSqlcStructure(t *testing.T) {
	root := t.TempDir()
	dir := Dir{
		Directories: []Dir{
			{
				Name: "internal",
				Directories: []Dir{
					{
						Name: "infra",
						Directories: []Dir{
							{Name: "migrations"},
							{
								Name: "sqlc",
								Directories: []Dir{
									{Name: "queries"},
								},
							},
						},
					},
				},
			},
		},
	}
	err := createDirStructure(root, dir)
	require.NoError(t, err)
	assert.DirExists(t, filepath.Join(root, "internal", "infra", "migrations"))
	assert.DirExists(t, filepath.Join(root, "internal", "infra", "sqlc"))
	assert.DirExists(t, filepath.Join(root, "internal", "infra", "sqlc", "queries"))
}

package new

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- Dir.Clone() ---

func TestDir_Clone_Empty(t *testing.T) {
	original := Dir{}
	clone := original.Clone()
	// Clone always allocates non-nil slices; compare field by field
	assert.Equal(t, original.Name, clone.Name)
	assert.Empty(t, clone.Files)
	assert.Empty(t, clone.Directories)
}

func TestDir_Clone_WithFiles(t *testing.T) {
	original := Dir{
		Name:  "root",
		Files: []File{{Name: "main", Extension: "go"}},
	}
	clone := original.Clone()
	assert.Equal(t, original.Name, clone.Name)
	assert.Equal(t, original.Files, clone.Files)

	// Mutate clone's Files — original must be unaffected
	clone.Files[0].Name = "changed"
	assert.Equal(t, "main", original.Files[0].Name)
}

func TestDir_Clone_WithSubDirs(t *testing.T) {
	original := Dir{
		Name: "root",
		Directories: []Dir{
			{Name: "internal"},
			{Name: "infra"},
		},
	}
	clone := original.Clone()
	assert.Equal(t, original.Name, clone.Name)
	assert.Equal(t, len(original.Directories), len(clone.Directories))
	assert.Equal(t, original.Directories[0].Name, clone.Directories[0].Name)
	assert.Equal(t, original.Directories[1].Name, clone.Directories[1].Name)

	// Mutate clone's subdirectory — original must be unaffected
	clone.Directories[0].Name = "changed"
	assert.Equal(t, "internal", original.Directories[0].Name)
}

func TestDir_Clone_DeepMutation(t *testing.T) {
	original := Dir{
		Name: "root",
		Directories: []Dir{
			{
				Name: "internal",
				Directories: []Dir{
					{Name: "infra"},
				},
			},
		},
	}
	clone := original.Clone()
	assert.Equal(t, "root", clone.Name)
	assert.Equal(t, "internal", clone.Directories[0].Name)
	assert.Equal(t, "infra", clone.Directories[0].Directories[0].Name)

	// Mutate deeply nested dir in clone — original must be unaffected
	clone.Directories[0].Directories[0].Name = "changed"
	assert.Equal(t, "infra", original.Directories[0].Directories[0].Name)
}

// --- Dir.AddSubDir() ---

func TestDir_AddSubDir_EmptyPath(t *testing.T) {
	root := Dir{Name: "root"}
	node := Dir{Name: "new"}
	root.AddSubDir([]string{}, node)
	assert.Equal(t, []Dir{{Name: "new"}}, root.Directories)
}

func TestDir_AddSubDir_SingleLevel(t *testing.T) {
	root := Dir{
		Name:        "root",
		Directories: []Dir{{Name: "internal"}},
	}
	node := Dir{Name: "infra"}
	root.AddSubDir([]string{"internal"}, node)
	assert.Equal(t, []Dir{{Name: "infra"}}, root.Directories[0].Directories)
}

func TestDir_AddSubDir_MultiLevel(t *testing.T) {
	root := Dir{
		Name: "root",
		Directories: []Dir{
			{
				Name:        "internal",
				Directories: []Dir{{Name: "infra"}},
			},
		},
	}
	node := Dir{Name: "migrations"}
	root.AddSubDir([]string{"internal", "infra"}, node)
	assert.Equal(t, []Dir{{Name: "migrations"}}, root.Directories[0].Directories[0].Directories)
}

func TestDir_AddSubDir_NoMatchingChild(t *testing.T) {
	root := Dir{
		Name:        "root",
		Directories: []Dir{{Name: "internal"}},
	}
	node := Dir{Name: "orphan"}
	// Path does not match any child — should be a no-op
	root.AddSubDir([]string{"nonexistent"}, node)
	assert.Equal(t, []Dir{{Name: "internal"}}, root.Directories)
}

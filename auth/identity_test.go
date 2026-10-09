package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIdentity_HasClaim(t *testing.T) {
	id := Identity{Claims: map[string]any{
		"verified": true,
		"banned":   false,
		"nothing":  nil,
		"role":     "admin",
		"roles":    []any{"member", "owner", 3},
		"scopes":   []string{"read"},
		"level":    3.0,
	}}

	cases := []struct {
		name   string
		claim  string
		values []string
		want   bool
	}{
		{"missing", "x", nil, false},
		{"false", "banned", nil, false},
		{"nil", "nothing", nil, false},
		{"present", "verified", nil, true},
		{"string match", "role", []string{"user", "admin"}, true},
		{"string miss", "role", []string{"user"}, false},
		{"any list contains", "roles", []string{"owner"}, true},
		{"any list miss", "roles", []string{"admin"}, false},
		{"string list contains", "scopes", []string{"read"}, true},
		{"non-string with values", "level", []string{"3"}, false},
		{"false with values", "banned", []string{"false"}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, id.HasClaim(c.claim, c.values...))
		})
	}
}

func TestGetIdentity_Missing(t *testing.T) {
	_, ok := GetIdentity(context.Background())
	assert.False(t, ok)
}

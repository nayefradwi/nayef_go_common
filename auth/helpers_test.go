package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

var testOwner = uuid.MustParse("00000000-0000-0000-0000-000000000001")

// testTokenID is a valid UUID for use as a reference-token id in middleware tests.
var testTokenID = uuid.MustParse("00000000-0000-0000-0000-0000000000aa")

func mustUUID(s string) uuid.UUID { return uuid.MustParse(s) }

func mustConfig(t *testing.T) JwtTokenProviderConfig {
	t.Helper()
	cfg, err := NewJwtTokenProviderConfig("test-secret-key-min-length", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

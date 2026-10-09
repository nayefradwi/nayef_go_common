package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

var testOwner = uuid.MustParse("00000000-0000-0000-0000-000000000001")

// testTokenID is a valid UUID for use as a reference-token id in middleware tests.
var testTokenID = uuid.MustParse("00000000-0000-0000-0000-0000000000aa")

const testSecret = "test-secret-key-at-least-32-bytes"

func mustUUID(s string) uuid.UUID { return uuid.MustParse(s) }

func mustConfig(t *testing.T) JwtTokenProviderConfig {
	t.Helper()
	cfg, err := NewJwtTokenProviderConfig(testSecret, time.Hour, AccessTokenType)
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

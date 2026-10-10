package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

var testOwner = uuid.MustParse("00000000-0000-0000-0000-000000000001")

const testSecret = "test-secret-key-at-least-32-bytes"

func mustConfig(t *testing.T) JwtTokenProviderConfig {
	t.Helper()
	cfg, err := NewJwtTokenProviderConfig(testSecret, time.Hour, AccessTokenType)
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

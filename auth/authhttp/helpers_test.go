package authhttp

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nayefradwi/nayef_go_common/auth"
)

var testOwner = uuid.MustParse("00000000-0000-0000-0000-000000000001")

var testTokenID = uuid.MustParse("00000000-0000-0000-0000-0000000000aa")

const testSecret = "test-secret-key-at-least-32-bytes"

func mustConfig(t *testing.T) auth.JwtTokenProviderConfig {
	t.Helper()
	cfg, err := auth.NewJwtTokenProviderConfig(testSecret, time.Hour, auth.AccessTokenType)
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

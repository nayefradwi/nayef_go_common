package auth

import (
	"testing"
	"time"
)

func mustConfig(t *testing.T) JwtTokenProviderConfig {
	t.Helper()
	cfg, err := NewJwtTokenProviderConfig("test-secret-key-min-length", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

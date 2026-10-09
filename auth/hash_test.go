package auth

import (
	"strings"
	"testing"

	. "github.com/nayefradwi/nayef_go_common/errors"
	"github.com/stretchr/testify/assert"
)

func TestHash_PasswordTooLong(t *testing.T) {
	hc := NewHashingConfig(10)
	longPassword := strings.Repeat("a", 73)
	_, err := hc.Hash(longPassword)
	if err == nil {
		t.Fatal("expected error for password exceeding 72 bytes")
	}
}

func TestHash_ExactlyMaxLength(t *testing.T) {
	hc := NewHashingConfig(10)
	password := strings.Repeat("a", 72)
	hash, err := hc.Hash(password)
	if err != nil {
		t.Fatalf("expected no error for 72-byte password: %v", err)
	}
	if !CompareHash(password, hash) {
		t.Fatal("hash comparison should succeed")
	}
}

func TestHash_RoundTrip(t *testing.T) {
	hc := DefaultHashingConfig
	password := "my-secure-password"
	hash, err := hc.Hash(password)
	if err != nil {
		t.Fatal(err)
	}
	if !CompareHash(password, hash) {
		t.Fatal("hash comparison should succeed")
	}
	if CompareHash("wrong-password", hash) {
		t.Fatal("hash comparison should fail for wrong password")
	}
}

func TestPasswordPolicy_Validate(t *testing.T) {
	p := DefaultPasswordPolicy
	cases := []struct {
		name     string
		password string
		ok       bool
	}{
		{"empty", "", false},
		{"7 runes", "1234567", false},
		{"8 runes", "12345678", true},
		{"8 multibyte runes", "éééééééé", true},
		{"7 multibyte runes over 8 bytes", "ééééééé", false},
		{"72 bytes", strings.Repeat("a", 72), true},
		{"73 bytes", strings.Repeat("a", 73), false},
		{"40 runes over 72 bytes", strings.Repeat("é", 40), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := p.Validate(c.password)
			if c.ok {
				assert.NoError(t, err)
				return
			}
			var re *ResultError
			assert.ErrorAs(t, err, &re)
		})
	}
}

package auth

import (
	"strings"
	"testing"
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

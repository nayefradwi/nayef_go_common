package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestNewJwtTokenProviderConfig_SecretKeyTooShort(t *testing.T) {
	_, err := NewJwtTokenProviderConfig(strings.Repeat("a", 31), time.Hour, AccessTokenType)
	if err == nil {
		t.Fatal("expected error for 31-byte secret key")
	}
	if _, err := NewJwtTokenProviderConfig(strings.Repeat("a", 32), time.Hour, AccessTokenType); err != nil {
		t.Fatal(err)
	}
}

func TestNewJwtTokenProviderConfig_InvalidTokenType(t *testing.T) {
	if _, err := NewJwtTokenProviderConfig(testSecret, time.Hour, 0); err == nil {
		t.Fatal("expected error for missing token type")
	}
}

func TestSetSecretKey_Empty(t *testing.T) {
	cfg := mustConfig(t)
	_, err := cfg.SetSecretKey("")
	if err == nil {
		t.Fatal("expected error for empty secret key")
	}
}

func TestSetSecretKey_UpdatesParserAndSigner(t *testing.T) {
	cfg := mustConfig(t)
	provider := NewJwtTokenProvider(cfg)

	token, err := provider.SignClaims(testOwner, map[string]any{})
	if err != nil {
		t.Fatal(err)
	}

	newCfg, err := cfg.SetSecretKey("different-secret-key-at-least-32-bytes")
	if err != nil {
		t.Fatal(err)
	}
	newProvider := NewJwtTokenProvider(newCfg)

	_, err = newProvider.GetClaims(token)
	if err == nil {
		t.Fatal("expected error parsing token with different key")
	}
}

func TestSetRSASigningMethod_NilKeys(t *testing.T) {
	cfg := mustConfig(t)

	_, err := cfg.SetRSASigningMethod(jwt.SigningMethodRS256, nil, nil)
	if err == nil {
		t.Fatal("expected error for nil RSA private key")
	}

	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	_, err = cfg.SetRSASigningMethod(jwt.SigningMethodRS256, key, nil)
	if err == nil {
		t.Fatal("expected error for nil RSA public key")
	}
}

func TestSetRSASigningMethod_InvalidAlg(t *testing.T) {
	cfg := mustConfig(t)
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	_, err := cfg.SetRSASigningMethod(jwt.SigningMethodES256, key, &key.PublicKey)
	if err == nil {
		t.Fatal("expected error for non-RSA signing method")
	}
}

func TestSetECDSASigningMethod_NilKeys(t *testing.T) {
	cfg := mustConfig(t)

	_, err := cfg.SetECDSASigningMethod(jwt.SigningMethodES256, nil, nil)
	if err == nil {
		t.Fatal("expected error for nil ECDSA private key")
	}

	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	_, err = cfg.SetECDSASigningMethod(jwt.SigningMethodES256, key, nil)
	if err == nil {
		t.Fatal("expected error for nil ECDSA public key")
	}
}

func TestSetECDSASigningMethod_InvalidAlg(t *testing.T) {
	cfg := mustConfig(t)
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	_, err := cfg.SetECDSASigningMethod(jwt.SigningMethodRS256, key, &key.PublicKey)
	if err == nil {
		t.Fatal("expected error for non-ECDSA signing method")
	}
}

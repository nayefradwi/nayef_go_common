package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestNewJwtTokenProviderConfig_EmptySecretKey(t *testing.T) {
	_, err := NewJwtTokenProviderConfig("", time.Hour)
	if err == nil {
		t.Fatal("expected error for empty secret key")
	}
}

func TestNewJwtTokenProviderConfig_ValidKey(t *testing.T) {
	cfg, err := NewJwtTokenProviderConfig("my-secret", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.SecretKey != "my-secret" {
		t.Fatalf("expected SecretKey 'my-secret', got %q", cfg.SecretKey)
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

	token, err := provider.SignClaims("owner1", map[string]any{})
	if err != nil {
		t.Fatal(err)
	}

	newCfg, err := cfg.SetSecretKey("different-secret-key")
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

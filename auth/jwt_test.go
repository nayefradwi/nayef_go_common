package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestSignClaims_DoesNotMutateInput(t *testing.T) {
	cfg := mustConfig(t)
	provider := NewJwtTokenProvider(cfg)

	claims := map[string]any{"role": "admin"}
	_, err := provider.SignClaims("owner1", claims)
	if err != nil {
		t.Fatal(err)
	}

	if len(claims) != 1 {
		t.Fatalf("expected claims to have 1 entry, got %d", len(claims))
	}
	if claims["role"] != "admin" {
		t.Fatal("expected claims to be unchanged")
	}
}

func TestTokenType_Enforcement(t *testing.T) {
	cfg := mustConfig(t)

	accessCfg := cfg.SetTokenType(AccessTokenType)
	refreshCfg := cfg.SetTokenType(RefreshTokenType)

	accessProvider := NewJwtTokenProvider(accessCfg)
	refreshProvider := NewJwtTokenProvider(refreshCfg)

	accessToken, err := accessProvider.SignClaims("owner1", map[string]any{})
	if err != nil {
		t.Fatal(err)
	}

	// Access provider should accept access token
	_, err = accessProvider.GetClaims(accessToken)
	if err != nil {
		t.Fatalf("access provider should accept access token: %v", err)
	}

	// Refresh provider should reject access token
	_, err = refreshProvider.GetClaims(accessToken)
	if err == nil {
		t.Fatal("refresh provider should reject access token")
	}
}

func TestTokenType_NotEnforced_WhenZero(t *testing.T) {
	cfg := mustConfig(t)

	typedProvider := NewJwtTokenProvider(cfg.SetTokenType(AccessTokenType))
	untypedProvider := NewJwtTokenProvider(cfg)

	token, err := typedProvider.SignClaims("owner1", map[string]any{})
	if err != nil {
		t.Fatal(err)
	}

	// Untyped provider (TokenType=0) should accept any token
	_, err = untypedProvider.GetClaims(token)
	if err != nil {
		t.Fatalf("untyped provider should accept any token: %v", err)
	}
}

func TestAudience_Enforcement(t *testing.T) {
	cfg := mustConfig(t)
	cfgWithAud := cfg.SetAudience("service-a")

	provider := NewJwtTokenProvider(cfgWithAud)
	token, err := provider.SignClaims("owner1", map[string]any{})
	if err != nil {
		t.Fatal(err)
	}

	// Same audience should work
	_, err = provider.GetClaims(token)
	if err != nil {
		t.Fatalf("expected token to be valid with matching audience: %v", err)
	}

	// Different audience should fail
	diffAudCfg := cfg.SetAudience("service-b")
	diffProvider := NewJwtTokenProvider(diffAudCfg)
	_, err = diffProvider.GetClaims(token)
	if err == nil {
		t.Fatal("expected error for mismatched audience")
	}
}

func TestRSASigningMethod_SignAndParse(t *testing.T) {
	cfg := mustConfig(t)
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	rsaCfg, err := cfg.SetRSASigningMethod(jwt.SigningMethodRS256, key, &key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}

	provider := NewJwtTokenProvider(rsaCfg)
	token, err := provider.SignClaims("owner1", map[string]any{"role": "admin"})
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := provider.GetClaims(token)
	if err != nil {
		t.Fatalf("failed to parse RSA token: %v", err)
	}
	if parsed.OwnerId != "owner1" {
		t.Fatalf("expected owner 'owner1', got %q", parsed.OwnerId)
	}
}

func TestECDSASigningMethod_SignAndParse(t *testing.T) {
	cfg := mustConfig(t)
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	ecCfg, err := cfg.SetECDSASigningMethod(jwt.SigningMethodES256, key, &key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}

	provider := NewJwtTokenProvider(ecCfg)
	token, err := provider.SignClaims("owner1", map[string]any{})
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := provider.GetClaims(token)
	if err != nil {
		t.Fatalf("failed to parse ECDSA token: %v", err)
	}
	if parsed.OwnerId != "owner1" {
		t.Fatalf("expected owner 'owner1', got %q", parsed.OwnerId)
	}
}

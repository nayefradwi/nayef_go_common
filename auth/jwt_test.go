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

func TestSignClaims_DoesNotMutateInput(t *testing.T) {
	cfg := mustConfig(t)
	provider := NewJwtTokenProvider(cfg)

	claims := map[string]any{"role": "admin"}
	_, err := provider.SignClaims(testOwner, claims)
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

func TestSignClaims_NilClaims(t *testing.T) {
	cfg := mustConfig(t)
	provider := NewJwtTokenProvider(cfg)

	signed, err := provider.SignClaims(testOwner, nil)
	if err != nil {
		t.Fatal(err)
	}

	token, err := provider.GetClaims(signed)
	if err != nil {
		t.Fatal(err)
	}
	if token.OwnerId != testOwner {
		t.Fatalf("expected owner %v, got %v", testOwner, token.OwnerId)
	}
}

func TestSignClaims_EmptyClaims(t *testing.T) {
	cfg := mustConfig(t)
	provider := NewJwtTokenProvider(cfg)

	signed, err := provider.SignClaims(testOwner, map[string]any{})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := provider.GetClaims(signed); err != nil {
		t.Fatal(err)
	}
}

func TestTokenType_Enforcement(t *testing.T) {
	accessCfg := mustConfig(t)
	refreshCfg, err := NewJwtTokenProviderConfig(testSecret, 24*time.Hour, RefreshTokenType)
	if err != nil {
		t.Fatal(err)
	}

	accessProvider := NewJwtTokenProvider(accessCfg)
	refreshProvider := NewJwtTokenProvider(refreshCfg)

	refreshToken, err := refreshProvider.SignClaims(testOwner, map[string]any{})
	if err != nil {
		t.Fatal(err)
	}

	// same secret on both sides, as in the README setup
	if _, err := accessProvider.GetClaims(refreshToken); err == nil {
		t.Fatal("access provider should reject refresh token")
	}
}

func TestGetClaims_WrongIssuerRejected(t *testing.T) {
	cfg := mustConfig(t)
	token, err := NewJwtTokenProvider(cfg.SetIssuer("other")).SignClaims(testOwner, nil)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := NewJwtTokenProvider(cfg).GetClaims(token); err == nil {
		t.Fatal("expected error for wrong issuer")
	}
}

func TestGetClaims_MissingExpRejected(t *testing.T) {
	cfg := mustConfig(t)
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		ownerClaimKey:     testOwner.String(),
		issuerClaimKey:    cfg.Issuer,
		tokenTypeClaimKey: AccessTokenType,
	}).SignedString([]byte(testSecret))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := NewJwtTokenProvider(cfg).GetClaims(token); err == nil {
		t.Fatal("expected error for token without exp")
	}
}

func TestGetClaims_OtherAlgInSameFamilyRejected(t *testing.T) {
	cfg := mustConfig(t)
	token, err := NewJwtTokenProvider(cfg.SetHS384SigningMethod()).SignClaims(testOwner, nil)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := NewJwtTokenProvider(cfg).GetClaims(token); err == nil {
		t.Fatal("HS256 config should reject HS384 token")
	}
}

func TestAudience_Enforcement(t *testing.T) {
	cfg := mustConfig(t)
	cfgWithAud := cfg.SetAudience("service-a")

	provider := NewJwtTokenProvider(cfgWithAud)
	token, err := provider.SignClaims(testOwner, map[string]any{})
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
	token, err := provider.SignClaims(testOwner, map[string]any{"role": "admin"})
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := provider.GetClaims(token)
	if err != nil {
		t.Fatalf("failed to parse RSA token: %v", err)
	}
	if parsed.OwnerId != testOwner {
		t.Fatalf("expected owner %v, got %v", testOwner, parsed.OwnerId)
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
	token, err := provider.SignClaims(testOwner, map[string]any{})
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := provider.GetClaims(token)
	if err != nil {
		t.Fatalf("failed to parse ECDSA token: %v", err)
	}
	if parsed.OwnerId != testOwner {
		t.Fatalf("expected owner %v, got %v", testOwner, parsed.OwnerId)
	}
}

package authhttp

import (
	"context"
	"errors"
	"github.com/nayefradwi/nayef_go_common/auth"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

type stubTokenProvider struct {
	token auth.Token
	err   error
}

func (s stubTokenProvider) GetClaims(_ string) (auth.Token, error) {
	return s.token, s.err
}

func (s stubTokenProvider) SignClaims(_ uuid.UUID, _ map[string]any) (string, error) {
	return "", nil
}

type stubVerifier struct {
	token auth.Token
	err   error
}

func (s stubVerifier) VerifyAccess(_ context.Context, _ string) (auth.Token, error) {
	return s.token, s.err
}

func unexpiredToken() auth.Token {
	return auth.Token{
		OwnerId:   testOwner,
		Claims:    map[string]any{},
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}
}

func expiredToken() auth.Token {
	return auth.Token{
		OwnerId:   testOwner,
		Claims:    map[string]any{},
		ExpiresAt: time.Now().UTC().Add(-time.Hour),
	}
}

func nextHandler(t *testing.T, called *bool) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*called = true
		w.WriteHeader(http.StatusOK)
	})
}

// --- JwtMiddleware ---

func TestJwtMiddleware_ValidToken(t *testing.T) {
	cfg := mustConfig(t)
	provider := auth.NewJwtTokenProvider(cfg)
	tokenStr, err := provider.SignClaims(testOwner, map[string]any{})
	if err != nil {
		t.Fatal(err)
	}

	called := false
	m := NewJwtMiddleware(provider)
	handler := m.UseAuthentication(nextHandler(t, &called))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !called {
		t.Error("expected next handler to be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestJwtMiddleware_MissingToken(t *testing.T) {
	called := false
	m := NewJwtMiddleware(stubTokenProvider{})
	handler := m.UseAuthentication(nextHandler(t, &called))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if called {
		t.Error("expected next handler NOT to be called")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestJwtMiddleware_InvalidToken(t *testing.T) {
	called := false
	m := NewJwtMiddleware(stubTokenProvider{err: errors.New("bad token")})
	handler := m.UseAuthentication(nextHandler(t, &called))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if called {
		t.Error("expected next handler NOT to be called")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestJwtMiddleware_TokenInContext(t *testing.T) {
	expectedToken := unexpiredToken()
	stub := stubTokenProvider{token: expectedToken}
	m := NewJwtMiddleware(stub)

	var gotIdentity auth.Identity
	capture := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotIdentity, _ = auth.GetIdentity(r.Context())
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer sometoken")
	w := httptest.NewRecorder()

	m.UseAuthentication(capture).ServeHTTP(w, req)

	if gotIdentity.OwnerId != expectedToken.OwnerId || gotIdentity.Kind != auth.KindUser {
		t.Errorf("expected user identity for %v, got %+v", expectedToken.OwnerId, gotIdentity)
	}
}

func TestJwtMiddleware_ExpiredToken(t *testing.T) {
	stub := stubTokenProvider{token: expiredToken()}
	called := false
	m := NewJwtMiddleware(stub)
	handler := m.UseAuthentication(nextHandler(t, &called))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer sometoken")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if called {
		t.Error("expected next handler NOT to be called")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// A token whose ExpiresAt is the zero value (a JWT carrying no exp claim) is
// treated as expired rather than as never-expiring.
func TestJwtMiddleware_ZeroExpiryIsRejected(t *testing.T) {
	stub := stubTokenProvider{token: auth.Token{OwnerId: testOwner, Claims: map[string]any{}}}
	called := false
	m := NewJwtMiddleware(stub)
	handler := m.UseAuthentication(nextHandler(t, &called))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer sometoken")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if called {
		t.Error("expected next handler NOT to be called")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// --- OpaqueMiddleware ---

func TestOpaqueMiddleware_ValidToken(t *testing.T) {
	stub := stubVerifier{token: unexpiredToken()}
	called := false
	m := NewOpaqueMiddleware(stub)
	handler := m.UseAuthentication(nextHandler(t, &called))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer opaque")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !called {
		t.Error("expected next handler to be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestOpaqueMiddleware_MissingToken(t *testing.T) {
	called := false
	m := NewOpaqueMiddleware(stubVerifier{})
	handler := m.UseAuthentication(nextHandler(t, &called))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if called {
		t.Error("expected next handler NOT to be called")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestOpaqueMiddleware_InvalidToken(t *testing.T) {
	stub := stubVerifier{err: errors.New("not found")}
	called := false
	m := NewOpaqueMiddleware(stub)
	handler := m.UseAuthentication(nextHandler(t, &called))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer opaque")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if called {
		t.Error("expected next handler NOT to be called")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestOpaqueMiddleware_ExpiredToken(t *testing.T) {
	stub := stubVerifier{token: expiredToken()}
	called := false
	m := NewOpaqueMiddleware(stub)
	handler := m.UseAuthentication(nextHandler(t, &called))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer opaque")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if called {
		t.Error("expected next handler NOT to be called")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestOpaqueMiddleware_TokenInContext(t *testing.T) {
	expectedToken := unexpiredToken()
	stub := stubVerifier{token: expectedToken}
	m := NewOpaqueMiddleware(stub)

	var gotIdentity auth.Identity
	capture := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotIdentity, _ = auth.GetIdentity(r.Context())
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer opaque")
	w := httptest.NewRecorder()

	m.UseAuthentication(capture).ServeHTTP(w, req)

	if gotIdentity.OwnerId != expectedToken.OwnerId || gotIdentity.Kind != auth.KindUser {
		t.Errorf("expected user identity for %v, got %+v", expectedToken.OwnerId, gotIdentity)
	}
}

// --- ApiKeyMiddleware ---

type stubKeyVerifier struct {
	identity auth.Identity
	err      error
}

func (s stubKeyVerifier) VerifyKey(context.Context, string) (auth.Identity, error) {
	return s.identity, s.err
}

func TestApiKeyMiddleware(t *testing.T) {
	called := false
	denied := NewApiKeyMiddleware(stubKeyVerifier{err: errors.New("not found")}).UseAuthentication(nextHandler(t, &called))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer sk_live_x")
	w := httptest.NewRecorder()
	denied.ServeHTTP(w, req)
	if called || w.Code != http.StatusUnauthorized {
		t.Fatalf("verifier error should 401, got %d called=%v", w.Code, called)
	}

	var got auth.Identity
	capture := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		got, _ = auth.GetIdentity(r.Context())
	})
	stub := stubKeyVerifier{identity: auth.Identity{OwnerId: testOwner, Kind: auth.KindKey}}
	w = httptest.NewRecorder()
	NewApiKeyMiddleware(stub).UseAuthentication(capture).ServeHTTP(w, req)
	if got.Kind != auth.KindKey || got.OwnerId != testOwner {
		t.Fatalf("expected key identity, got %+v", got)
	}
}

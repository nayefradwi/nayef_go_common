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

type stubReferenceTokenProvider struct {
	token auth.Token
	err   error
}

func (s stubReferenceTokenProvider) GenerateId() (uuid.UUID, error) { return uuid.Nil, nil }
func (s stubReferenceTokenProvider) GenerateToken(_ context.Context, _ uuid.UUID, _ map[string]any) (auth.TokenDTO, error) {
	return auth.TokenDTO{}, nil
}
func (s stubReferenceTokenProvider) GetAccessToken(_ context.Context, _ uuid.UUID) (auth.Token, error) {
	return s.token, s.err
}
func (s stubReferenceTokenProvider) GetRefreshToken(_ context.Context, _ uuid.UUID) (auth.Token, error) {
	return s.token, s.err
}
func (s stubReferenceTokenProvider) RevokeToken(_ context.Context, _ uuid.UUID) error { return nil }
func (s stubReferenceTokenProvider) RevokeOwner(_ context.Context, _ uuid.UUID) error { return nil }
func (s stubReferenceTokenProvider) GetAccessTokenProvider() auth.ITokenProvider      { return nil }

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

	var gotToken auth.Token
	capture := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotToken, _ = auth.GetToken(r.Context())
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer sometoken")
	w := httptest.NewRecorder()

	m.UseAuthentication(capture).ServeHTTP(w, req)

	if gotToken.OwnerId != expectedToken.OwnerId {
		t.Errorf("expected owner %v, got %v", expectedToken.OwnerId, gotToken.OwnerId)
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

// --- ReferenceTokenMiddleware ---

func TestReferenceTokenMiddleware_ValidToken(t *testing.T) {
	stub := stubReferenceTokenProvider{token: unexpiredToken()}
	called := false
	m := NewReferenceTokenMiddleware(stub)
	handler := m.UseAuthentication(nextHandler(t, &called))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+testTokenID.String())
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !called {
		t.Error("expected next handler to be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestReferenceTokenMiddleware_MissingToken(t *testing.T) {
	called := false
	m := NewReferenceTokenMiddleware(stubReferenceTokenProvider{})
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

func TestReferenceTokenMiddleware_InvalidToken(t *testing.T) {
	stub := stubReferenceTokenProvider{err: errors.New("not found")}
	called := false
	m := NewReferenceTokenMiddleware(stub)
	handler := m.UseAuthentication(nextHandler(t, &called))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer bad-id")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if called {
		t.Error("expected next handler NOT to be called")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestReferenceTokenMiddleware_ExpiredToken(t *testing.T) {
	stub := stubReferenceTokenProvider{token: expiredToken()}
	called := false
	m := NewReferenceTokenMiddleware(stub)
	handler := m.UseAuthentication(nextHandler(t, &called))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+testTokenID.String())
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if called {
		t.Error("expected next handler NOT to be called")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestReferenceTokenMiddleware_TokenInContext(t *testing.T) {
	expectedToken := unexpiredToken()
	stub := stubReferenceTokenProvider{token: expectedToken}
	m := NewReferenceTokenMiddleware(stub)

	var gotToken auth.Token
	capture := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotToken, _ = auth.GetToken(r.Context())
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+testTokenID.String())
	w := httptest.NewRecorder()

	m.UseAuthentication(capture).ServeHTTP(w, req)

	if gotToken.OwnerId != expectedToken.OwnerId {
		t.Errorf("expected owner %v, got %v", expectedToken.OwnerId, gotToken.OwnerId)
	}
}

func TestGetToken_MissingReturnsNotOk(t *testing.T) {
	if _, ok := auth.GetToken(context.Background()); ok {
		t.Fatal("expected ok=false when no token is in context")
	}
}

package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubTokenProvider struct {
	token Token
	err   error
}

func (s stubTokenProvider) GetClaims(_ string) (Token, error) {
	return s.token, s.err
}

func (s stubTokenProvider) SignClaims(_ string, _ map[string]any) (string, error) {
	return "", nil
}

type stubReferenceTokenProvider struct {
	token Token
	err   error
}

func (s stubReferenceTokenProvider) GenerateId() (string, error) { return "", nil }
func (s stubReferenceTokenProvider) GenerateToken(_ string, _ map[string]any) (TokenDTO, error) {
	return TokenDTO{}, nil
}
func (s stubReferenceTokenProvider) GetAccessToken(_ string) (Token, error)  { return s.token, s.err }
func (s stubReferenceTokenProvider) GetRefreshToken(_ string) (Token, error) { return s.token, s.err }
func (s stubReferenceTokenProvider) RevokeToken(_ string) error              { return nil }
func (s stubReferenceTokenProvider) RevokeOwner(_ string) error              { return nil }
func (s stubReferenceTokenProvider) GetAccessTokenProvider() ITokenProvider  { return nil }

func nextHandler(t *testing.T, called *bool) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*called = true
		w.WriteHeader(http.StatusOK)
	})
}

// --- JwtAuthenticationMiddleware ---

func TestJwtAuthenticationMiddleware_ValidToken(t *testing.T) {
	cfg := mustConfig(t)
	provider := NewJwtTokenProvider(cfg)
	tokenStr, err := provider.SignClaims("owner1", map[string]any{})
	if err != nil {
		t.Fatal(err)
	}

	called := false
	m := NewJwtAuthenticationMiddleware(provider)
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

func TestJwtAuthenticationMiddleware_MissingToken(t *testing.T) {
	called := false
	m := NewJwtAuthenticationMiddleware(stubTokenProvider{})
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

func TestJwtAuthenticationMiddleware_InvalidToken(t *testing.T) {
	called := false
	m := NewJwtAuthenticationMiddleware(stubTokenProvider{err: errors.New("bad token")})
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

func TestJwtAuthenticationMiddleware_TokenInContext(t *testing.T) {
	expectedToken := Token{OwnerId: "user42", Claims: map[string]any{}}
	stub := stubTokenProvider{token: expectedToken}
	m := NewJwtAuthenticationMiddleware(stub)

	var gotToken Token
	capture := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotToken = GetToken(r.Context())
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer sometoken")
	w := httptest.NewRecorder()

	m.UseAuthentication(capture).ServeHTTP(w, req)

	if gotToken.OwnerId != expectedToken.OwnerId {
		t.Errorf("expected owner %q, got %q", expectedToken.OwnerId, gotToken.OwnerId)
	}
}

// --- JwtReferenceTokenAuthenticationMiddleware ---

func TestJwtReferenceTokenAuthenticationMiddleware_ValidToken(t *testing.T) {
	stub := stubReferenceTokenProvider{token: Token{OwnerId: "user1", Claims: map[string]any{}}}
	called := false
	m := NewJwtReferenceTokenAuthenticationMiddleware(stub)
	handler := m.UseAuthentication(nextHandler(t, &called))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer some-ref-id")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !called {
		t.Error("expected next handler to be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestJwtReferenceTokenAuthenticationMiddleware_MissingToken(t *testing.T) {
	called := false
	m := NewJwtReferenceTokenAuthenticationMiddleware(stubReferenceTokenProvider{})
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

func TestJwtReferenceTokenAuthenticationMiddleware_InvalidToken(t *testing.T) {
	stub := stubReferenceTokenProvider{err: errors.New("not found")}
	called := false
	m := NewJwtReferenceTokenAuthenticationMiddleware(stub)
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

func TestJwtReferenceTokenAuthenticationMiddleware_TokenInContext(t *testing.T) {
	expectedToken := Token{OwnerId: "ref-owner", Claims: map[string]any{}}
	stub := stubReferenceTokenProvider{token: expectedToken}
	m := NewJwtReferenceTokenAuthenticationMiddleware(stub)

	var gotToken Token
	capture := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotToken = GetToken(r.Context())
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer some-ref-id")
	w := httptest.NewRecorder()

	m.UseAuthentication(capture).ServeHTTP(w, req)

	if gotToken.OwnerId != expectedToken.OwnerId {
		t.Errorf("expected owner %q, got %q", expectedToken.OwnerId, gotToken.OwnerId)
	}
}

// Email + password login: signup, login, refresh, logout, logout-all, me.
package main

import (
	"context"
	"crypto/rand"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nayefradwi/nayef_go_common/auth"
	"github.com/nayefradwi/nayef_go_common/auth/authhttp"
	"github.com/nayefradwi/nayef_go_common/auth/authpg"
	"github.com/nayefradwi/nayef_go_common/auth/examples/shared"
	"github.com/nayefradwi/nayef_go_common/errors"
	"github.com/nayefradwi/nayef_go_common/httputil"
)

var errInvalidLogin = errors.UnauthorizedError("invalid email or password")

type app struct {
	users     shared.Users
	sessions  auth.SessionManager
	logins    auth.Limiter
	policy    auth.PasswordPolicy
	hashing   auth.HashingConfig
	dummyHash string
}

func main() {
	ctx := context.Background()
	pool := shared.Connect(ctx)
	defer pool.Close()

	// redis instead of postgres: rdb := shared.Redis(ctx), then the authredis swap on each store
	sessionStore := shared.Must(authpg.NewSessionStore(pool)) // swap: authredis.NewSessionStore(rdb)
	attemptStore := authpg.NewAttemptStore(pool)              // swap: authredis.NewAttemptStore(rdb)

	sessions, authenticate := opaqueSessions(sessionStore) // swap: jwtSessions(sessionStore, os.Getenv("JWT_SECRET"))
	hashing := auth.DefaultHashingConfig
	a := app{
		users:    shared.NewUsers(pool),
		sessions: sessions,
		logins:   shared.Must(auth.NewLimiter(attemptStore, 5, 15*time.Minute)),
		policy:   auth.DefaultPasswordPolicy,
		hashing:  hashing,
		// checked against for unknown emails, so a miss costs the same bcrypt time as a hit
		dummyHash: shared.Must(hashing.Hash(rand.Text())),
	}

	r := chi.NewRouter()
	r.Post("/signup", a.signup)
	r.Post("/login", a.login)
	r.Post("/refresh", a.refresh)
	r.Post("/logout", a.logout)
	r.Group(func(r chi.Router) {
		r.Use(authenticate)
		r.Get("/me", a.me)
		r.Post("/logout-all", a.logoutAll)
	})

	shared.Serve(r)
}

const (
	accessTTL  = 15 * time.Minute
	refreshTTL = 30 * 24 * time.Hour
)

// opaque access tokens are looked up on every request, so logout ends them at once
func opaqueSessions(store auth.SessionStore) (auth.SessionManager, func(http.Handler) http.Handler) {
	manager := shared.Must(auth.NewOpaqueSessionManager(store, accessTTL, refreshTTL))
	return manager, authhttp.NewOpaqueMiddleware(manager).UseAuthentication
}

// jwt access tokens are checked without a store read, so logout only ends the
// refresh token; access tokens already issued live until accessTTL
func jwtSessions(store auth.SessionStore, secret string) (auth.SessionManager, func(http.Handler) http.Handler) {
	config := shared.Must(auth.NewJwtTokenProviderConfig(secret, accessTTL, auth.AccessTokenType))
	provider := auth.NewJwtTokenProvider(config)
	manager := shared.Must(auth.NewJwtSessionManager(store, provider, refreshTTL))
	return manager, authhttp.NewJwtMiddleware(provider).UseAuthentication
}

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshBody struct {
	RefreshToken string `json:"refresh_token"`
}

func (a app) signup(w http.ResponseWriter, r *http.Request) {
	jw := httputil.NewJsonResponseWriter(w)
	body, err := shared.Decode[credentials](r)
	if err != nil {
		jw.WriteError(err)
		return
	}

	email, err := shared.ParseEmail(body.Email)
	if err != nil {
		jw.WriteError(err)
		return
	}

	if err := a.policy.Validate(body.Password); err != nil {
		jw.WriteError(err)
		return
	}

	hash, err := a.hashing.Hash(body.Password)
	if err != nil {
		jw.WriteError(err)
		return
	}

	user, err := a.users.Create(r.Context(), email, hash, true)
	if err != nil {
		jw.WriteError(err)
		return
	}

	tokens, err := a.sessions.Issue(r.Context(), user.Id, nil)
	jw.WithSuccessStatus(http.StatusCreated).WriteJsonResponse(tokens, err)
}

func (a app) login(w http.ResponseWriter, r *http.Request) {
	jw := httputil.NewJsonResponseWriter(w)
	body, err := shared.Decode[credentials](r)
	if err != nil {
		jw.WriteError(err)
		return
	}

	email, err := shared.ParseEmail(body.Email)
	if err != nil {
		jw.WriteError(err)
		return
	}

	// keyed on the email, not the IP: caps guesses against one account from anywhere
	allowed, _, err := a.logins.Allow(r.Context(), "login:"+email)
	if err != nil {
		jw.WriteError(err)
		return
	}

	if !allowed {
		jw.WriteError(shared.ErrTooManyRequests)
		return
	}

	user, err := a.users.ByEmail(r.Context(), email)
	if errors.Is(err, shared.ErrUserNotFound) {
		auth.CompareHash(body.Password, a.dummyHash)
		jw.WriteError(errInvalidLogin)
		return
	}

	if err != nil {
		jw.WriteError(err)
		return
	}

	if !auth.CompareHash(body.Password, user.PasswordHash) {
		jw.WriteError(errInvalidLogin)
		return
	}

	tokens, err := a.sessions.Issue(r.Context(), user.Id, nil)
	jw.WriteJsonResponse(tokens, err)
}

func (a app) refresh(w http.ResponseWriter, r *http.Request) {
	jw := httputil.NewJsonResponseWriter(w)
	body, err := shared.Decode[refreshBody](r)
	if err != nil {
		jw.WriteError(err)
		return
	}

	// a refresh token used twice ends the whole login (reuse detection)
	tokens, err := a.sessions.Refresh(r.Context(), body.RefreshToken)
	jw.WriteJsonResponse(tokens, err)
}

func (a app) logout(w http.ResponseWriter, r *http.Request) {
	jw := httputil.NewJsonResponseWriter(w)
	body, err := shared.Decode[refreshBody](r)
	if err != nil {
		jw.WriteError(err)
		return
	}

	err = a.sessions.Revoke(r.Context(), body.RefreshToken)
	jw.WriteSuccessMessage("logged out", err)
}

func (a app) me(w http.ResponseWriter, r *http.Request) {
	jw := httputil.NewJsonResponseWriter(w)
	id, ok := auth.GetIdentity(r.Context())
	if !ok {
		jw.WriteError(errors.UnauthorizedError("not logged in"))
		return
	}

	user, err := a.users.ById(r.Context(), id.OwnerId)
	jw.WriteJsonResponse(user, err)
}

func (a app) logoutAll(w http.ResponseWriter, r *http.Request) {
	jw := httputil.NewJsonResponseWriter(w)
	id, ok := auth.GetIdentity(r.Context())
	if !ok {
		jw.WriteError(errors.UnauthorizedError("not logged in"))
		return
	}

	err := a.sessions.RevokeOwner(r.Context(), id.OwnerId)
	jw.WriteSuccessMessage("logged out everywhere", err)
}

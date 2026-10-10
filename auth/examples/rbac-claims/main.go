// RBAC without a DB: the role and its scopes go into the session claims at login,
// routes check claims only. A role change shows up at the next login.
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

var roleScopes = map[string][]string{
	"member": {"reports:read"},
	"admin":  {"reports:read", "reports:write"},
}

type app struct {
	users     shared.Users
	sessions  auth.SessionManager
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

	sessions, authenticate := opaqueSessions(sessionStore) // swap: jwtSessions(sessionStore, os.Getenv("JWT_SECRET"))
	hashing := auth.DefaultHashingConfig
	a := app{
		users:    shared.NewUsers(pool),
		sessions: sessions,
		policy:   auth.DefaultPasswordPolicy,
		hashing:  hashing,
		// checked against for unknown emails, so a miss costs the same bcrypt time as a hit
		dummyHash: shared.Must(hashing.Hash(rand.Text())),
	}

	r := chi.NewRouter()
	r.Post("/signup", a.signup)
	r.Post("/login", a.login)
	r.Post("/refresh", a.refresh)
	r.Group(func(r chi.Router) {
		r.Use(authenticate)
		r.Get("/me", a.me)
		r.With(authhttp.Has(authhttp.HasClaim("scopes", "reports:read"))).Get("/reports", a.listReports)
		r.With(authhttp.Has(authhttp.HasClaim("scopes", "reports:write"))).Post("/reports", a.createReport)
		r.With(authhttp.VerifyClaim("role", "admin")).Get("/admin", a.admin)
	})

	shared.Serve(r)
}

const (
	accessTTL  = 15 * time.Minute
	refreshTTL = 30 * 24 * time.Hour
)

// opaque sessions read claims from the session row on every request
func opaqueSessions(store auth.SessionStore) (auth.SessionManager, func(http.Handler) http.Handler) {
	manager := shared.Must(auth.NewOpaqueSessionManager(store, accessTTL, refreshTTL))
	return manager, authhttp.NewOpaqueMiddleware(manager).UseAuthentication
}

// jwt sessions sign the claims into the access token
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

	claims := map[string]any{"role": user.Role, "scopes": roleScopes[user.Role]}
	tokens, err := a.sessions.Issue(r.Context(), user.Id, claims)
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

	// read from the users table only here: the session keeps these claims until it ends
	claims := map[string]any{"role": user.Role, "scopes": roleScopes[user.Role]}
	tokens, err := a.sessions.Issue(r.Context(), user.Id, claims)
	jw.WriteJsonResponse(tokens, err)
}

// the new tokens carry the claims from login, not the user's current role
func (a app) refresh(w http.ResponseWriter, r *http.Request) {
	jw := httputil.NewJsonResponseWriter(w)
	body, err := shared.Decode[refreshBody](r)
	if err != nil {
		jw.WriteError(err)
		return
	}

	tokens, err := a.sessions.Refresh(r.Context(), body.RefreshToken)
	jw.WriteJsonResponse(tokens, err)
}

func (a app) me(w http.ResponseWriter, r *http.Request) {
	jw := httputil.NewJsonResponseWriter(w)
	id, ok := auth.GetIdentity(r.Context())
	if !ok {
		jw.WriteError(errors.UnauthorizedError("not logged in"))
		return
	}

	jw.WriteData(map[string]any{"id": id.OwnerId, "role": id.Claims["role"], "scopes": id.Claims["scopes"]})
}

func (a app) listReports(w http.ResponseWriter, r *http.Request) {
	httputil.NewJsonResponseWriter(w).WriteData([]string{"q1", "q2"})
}

func (a app) createReport(w http.ResponseWriter, r *http.Request) {
	httputil.NewJsonResponseWriter(w).WithSuccessStatus(http.StatusCreated).WriteSuccessMessage("report created", nil)
}

func (a app) admin(w http.ResponseWriter, r *http.Request) {
	httputil.NewJsonResponseWriter(w).WriteSuccessMessage("hello admin", nil)
}

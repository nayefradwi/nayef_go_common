// API keys / m2m: users manage their keys behind a session; key-only routes take
// just keys; mixed routes take either and branch on the identity kind.
package main

import (
	"context"
	"crypto/rand"
	"net/http"
	"slices"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nayefradwi/nayef_go_common/auth"
	"github.com/nayefradwi/nayef_go_common/auth/authhttp"
	"github.com/nayefradwi/nayef_go_common/auth/authpg"
	"github.com/nayefradwi/nayef_go_common/auth/examples/shared"
	"github.com/nayefradwi/nayef_go_common/errors"
	"github.com/nayefradwi/nayef_go_common/httputil"
)

var errInvalidLogin = errors.UnauthorizedError("invalid email or password")

var keyScopes = []string{"reports:read", "reports:write"}

type app struct {
	users     shared.Users
	sessions  auth.SessionManager
	keys      auth.ApiKeyManager
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
	apiKeyStore := shared.Must(authpg.NewApiKeyStore(pool))   // swap: authredis.NewApiKeyStore(rdb)

	keys := shared.Must(auth.NewApiKeyManager(apiKeyStore, "sk_live"))
	sessions, authenticate, either := opaqueSessions(sessionStore, keys) // swap: jwtSessions(sessionStore, keys, os.Getenv("JWT_SECRET"))
	hashing := auth.DefaultHashingConfig
	a := app{
		users:    shared.NewUsers(pool),
		sessions: sessions,
		keys:     keys,
		policy:   auth.DefaultPasswordPolicy,
		hashing:  hashing,
		// checked against for unknown emails, so a miss costs the same bcrypt time as a hit
		dummyHash: shared.Must(hashing.Hash(rand.Text())),
	}

	r := chi.NewRouter()
	r.Post("/signup", a.signup)
	r.Post("/login", a.login)

	// users only: managing keys needs a session, a key can't mint keys
	r.Group(func(r chi.Router) {
		r.Use(authenticate)
		r.Post("/keys", a.createKey)
		r.Get("/keys", a.listKeys)
		r.Delete("/keys/{id}", a.revokeKey)
	})

	// keys only
	r.Group(func(r chi.Router) {
		r.Use(authhttp.NewApiKeyMiddleware(keys).UseAuthentication)
		r.Get("/m2m/ping", a.ping)
	})

	// users or keys, told apart by the key prefix with no extra DB read
	r.Group(func(r chi.Router) {
		r.Use(either)
		r.Get("/reports", a.listReports)
		r.With(authhttp.Has(authhttp.IsKey)).Post("/reports/ingest", a.ingestReports)
		r.With(authhttp.Has(authhttp.ByKind(authhttp.IsUser, authhttp.HasClaim("scopes", "reports:write")))).Post("/reports", a.createReport)
	})

	shared.Serve(r)
}

const (
	accessTTL  = 15 * time.Minute
	refreshTTL = 30 * 24 * time.Hour
)

type middleware = func(http.Handler) http.Handler

func opaqueSessions(store auth.SessionStore, keys auth.ApiKeyManager) (auth.SessionManager, middleware, middleware) {
	manager := shared.Must(auth.NewOpaqueSessionManager(store, accessTTL, refreshTTL))
	return manager,
		authhttp.NewOpaqueMiddleware(manager).UseAuthentication,
		authhttp.NewEitherMiddleware(manager, keys).UseAuthentication
}

func jwtSessions(store auth.SessionStore, keys auth.ApiKeyManager, secret string) (auth.SessionManager, middleware, middleware) {
	config := shared.Must(auth.NewJwtTokenProviderConfig(secret, accessTTL, auth.AccessTokenType))
	provider := auth.NewJwtTokenProvider(config)
	manager := shared.Must(auth.NewJwtSessionManager(store, provider, refreshTTL))
	return manager,
		authhttp.NewJwtMiddleware(provider).UseAuthentication,
		authhttp.NewJwtEitherMiddleware(provider, keys).UseAuthentication
}

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type createKeyBody struct {
	Name      string    `json:"name"`
	Scopes    []string  `json:"scopes"`
	ExpiresAt time.Time `json:"expires_at"` // zero = never
}

type keyView struct {
	Id         uuid.UUID      `json:"id"`
	Name       string         `json:"name"`
	Claims     map[string]any `json:"claims"`
	ExpiresAt  time.Time      `json:"expires_at,omitzero"`
	LastUsedAt time.Time      `json:"last_used_at,omitzero"`
	CreatedAt  time.Time      `json:"created_at"`
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

// the raw key is in this response only; the store keeps its hash
func (a app) createKey(w http.ResponseWriter, r *http.Request) {
	jw := httputil.NewJsonResponseWriter(w)
	id, ok := auth.GetIdentity(r.Context())
	if !ok {
		jw.WriteError(errors.UnauthorizedError("not logged in"))
		return
	}

	body, err := shared.Decode[createKeyBody](r)
	if err != nil {
		jw.WriteError(err)
		return
	}

	if body.Name == "" {
		jw.WriteError(errors.NewValidationError(errors.Field("name", "required", "name is required")))
		return
	}

	for _, scope := range body.Scopes {
		if !slices.Contains(keyScopes, scope) {
			jw.WriteError(errors.NewValidationError(errors.Field("scopes", "invalid", "unknown scope "+scope)))
			return
		}
	}

	raw, key, err := a.keys.Issue(r.Context(), id.OwnerId, body.Name, map[string]any{"scopes": body.Scopes}, body.ExpiresAt)
	if err != nil {
		jw.WriteError(err)
		return
	}

	jw.WithSuccessStatus(http.StatusCreated).WriteData(map[string]any{"key": raw, "id": key.Id, "name": key.Name})
}

func (a app) listKeys(w http.ResponseWriter, r *http.Request) {
	jw := httputil.NewJsonResponseWriter(w)
	id, ok := auth.GetIdentity(r.Context())
	if !ok {
		jw.WriteError(errors.UnauthorizedError("not logged in"))
		return
	}

	keys, err := a.keys.List(r.Context(), id.OwnerId)
	if err != nil {
		jw.WriteError(err)
		return
	}

	// never return ApiKey as is: it carries the hash
	views := make([]keyView, len(keys))
	for i, k := range keys {
		views[i] = keyView{Id: k.Id, Name: k.Name, Claims: k.Claims, ExpiresAt: k.ExpiresAt, LastUsedAt: k.LastUsedAt, CreatedAt: k.CreatedAt}
	}

	jw.WriteData(views)
}

// scoped to the caller: another user's key id is a 404
func (a app) revokeKey(w http.ResponseWriter, r *http.Request) {
	jw := httputil.NewJsonResponseWriter(w)
	id, ok := auth.GetIdentity(r.Context())
	if !ok {
		jw.WriteError(errors.UnauthorizedError("not logged in"))
		return
	}

	keyId, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jw.WriteError(errors.NotFoundError("api key not found"))
		return
	}

	err = a.keys.Revoke(r.Context(), id.OwnerId, keyId)
	jw.WriteSuccessMessage("api key revoked", err)
}

func (a app) ping(w http.ResponseWriter, r *http.Request) {
	jw := httputil.NewJsonResponseWriter(w)
	id, ok := auth.GetIdentity(r.Context())
	if !ok {
		jw.WriteError(errors.UnauthorizedError("no api key"))
		return
	}

	jw.WriteData(map[string]any{"owner": id.OwnerId, "claims": id.Claims})
}

func (a app) listReports(w http.ResponseWriter, r *http.Request) {
	jw := httputil.NewJsonResponseWriter(w)
	id, ok := auth.GetIdentity(r.Context())
	if !ok {
		jw.WriteError(errors.UnauthorizedError("not logged in"))
		return
	}

	jw.WriteData(map[string]any{"kind": id.Kind, "reports": []string{"q1", "q2"}})
}

func (a app) ingestReports(w http.ResponseWriter, r *http.Request) {
	httputil.NewJsonResponseWriter(w).WithSuccessStatus(http.StatusAccepted).WriteSuccessMessage("ingested", nil)
}

func (a app) createReport(w http.ResponseWriter, r *http.Request) {
	httputil.NewJsonResponseWriter(w).WithSuccessStatus(http.StatusCreated).WriteSuccessMessage("report created", nil)
}

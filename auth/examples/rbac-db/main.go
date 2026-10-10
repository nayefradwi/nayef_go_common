// RBAC with a DB: each request loads the user's permissions from a role table,
// so a role change applies on the next request. Route groups layer the checks.
package main

import (
	"context"
	"crypto/rand"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nayefradwi/nayef_go_common/auth"
	"github.com/nayefradwi/nayef_go_common/auth/authhttp"
	"github.com/nayefradwi/nayef_go_common/auth/authpg"
	"github.com/nayefradwi/nayef_go_common/auth/examples/shared"
	"github.com/nayefradwi/nayef_go_common/errors"
	"github.com/nayefradwi/nayef_go_common/httputil"
)

const schema = `
CREATE TABLE IF NOT EXISTS role_permissions (
    role TEXT NOT NULL,
    permission TEXT NOT NULL,
    PRIMARY KEY (role, permission)
);
INSERT INTO role_permissions VALUES
    ('member', 'reports:read'),
    ('admin', 'reports:read'),
    ('admin', 'reports:write')
ON CONFLICT DO NOTHING;`

var errInvalidLogin = errors.UnauthorizedError("invalid email or password")

type app struct {
	pool      *pgxpool.Pool
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
	shared.Must(pool.Exec(ctx, schema))

	// redis instead of postgres: rdb := shared.Redis(ctx), then the authredis swap on each store
	sessionStore := shared.Must(authpg.NewSessionStore(pool)) // swap: authredis.NewSessionStore(rdb)

	sessions, authenticate := opaqueSessions(sessionStore) // swap: jwtSessions(sessionStore, os.Getenv("JWT_SECRET"))
	hashing := auth.DefaultHashingConfig
	a := app{
		pool:     pool,
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
	r.Group(func(r chi.Router) {
		r.Use(authenticate)
		r.Get("/me", a.me)

		r.Group(func(r chi.Router) {
			r.Use(authhttp.Has(a.can("reports:read")))
			r.Get("/reports", a.listReports)

			r.With(authhttp.Has(a.can("reports:write"))).Post("/reports", a.createReport)
		})
	})

	shared.Serve(r)
}

// can is the Check: one query per request, true when the user's role grants the permission
func (a app) can(permission string) authhttp.Check {
	return func(ctx context.Context, id auth.Identity, _ *http.Request) (bool, error) {
		var allowed bool
		err := a.pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM users u JOIN role_permissions p ON p.role = u.role
				WHERE u.id = $1 AND p.permission = $2
			)`, id.OwnerId, permission,
		).Scan(&allowed)
		return allowed, err
	}
}

const (
	accessTTL  = 15 * time.Minute
	refreshTTL = 30 * 24 * time.Hour
)

func opaqueSessions(store auth.SessionStore) (auth.SessionManager, func(http.Handler) http.Handler) {
	manager := shared.Must(auth.NewOpaqueSessionManager(store, accessTTL, refreshTTL))
	return manager, authhttp.NewOpaqueMiddleware(manager).UseAuthentication
}

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

func (a app) listReports(w http.ResponseWriter, r *http.Request) {
	httputil.NewJsonResponseWriter(w).WriteData([]string{"q1", "q2"})
}

func (a app) createReport(w http.ResponseWriter, r *http.Request) {
	httputil.NewJsonResponseWriter(w).WithSuccessStatus(http.StatusCreated).WriteSuccessMessage("report created", nil)
}

// Account activation: signup leaves the account inactive until the emailed link is confirmed.
package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nayefradwi/nayef_go_common/auth"
	"github.com/nayefradwi/nayef_go_common/auth/authpg"
	"github.com/nayefradwi/nayef_go_common/auth/examples/shared"
	"github.com/nayefradwi/nayef_go_common/errors"
	"github.com/nayefradwi/nayef_go_common/httputil"
)

var errInvalidLogin = errors.UnauthorizedError("invalid email or password")

type app struct {
	users       shared.Users
	sessions    auth.SessionManager
	activations auth.OneTimeTokenManager
	policy      auth.PasswordPolicy
	hashing     auth.HashingConfig
	dummyHash   string
}

func main() {
	ctx := context.Background()
	pool := shared.Connect(ctx)
	defer pool.Close()

	// redis instead of postgres: rdb := shared.Redis(ctx), then the authredis swap on each store
	sessionStore := shared.Must(authpg.NewSessionStore(pool)) // swap: authredis.NewSessionStore(rdb)
	codeStore := shared.Must(authpg.NewCodeStore(pool))       // swap: authredis.NewCodeStore(rdb)

	sessions := opaqueSessions(sessionStore) // swap: jwtSessions(sessionStore, os.Getenv("JWT_SECRET"))
	hashing := auth.DefaultHashingConfig
	a := app{
		users:    shared.NewUsers(pool),
		sessions: sessions,
		activations: shared.Must(auth.NewOneTimeTokenManager(codeStore, auth.CodeLimits{
			TTL: 24 * time.Hour, ResendAfter: time.Minute, MaxSends: 5, MaxAttempts: 5,
		})),
		policy:  auth.DefaultPasswordPolicy,
		hashing: hashing,
		// checked against for unknown emails, so a miss costs the same bcrypt time as a hit
		dummyHash: shared.Must(hashing.Hash(rand.Text())),
	}

	r := chi.NewRouter()
	r.Post("/signup", a.signup)
	r.Post("/activate", a.activate)
	r.Post("/login", a.login)

	shared.Serve(r)
}

const (
	accessTTL  = 15 * time.Minute
	refreshTTL = 30 * 24 * time.Hour
)

func opaqueSessions(store auth.SessionStore) auth.SessionManager {
	return shared.Must(auth.NewOpaqueSessionManager(store, accessTTL, refreshTTL))
}

func jwtSessions(store auth.SessionStore, secret string) auth.SessionManager {
	config := shared.Must(auth.NewJwtTokenProviderConfig(secret, accessTTL, auth.AccessTokenType))
	return shared.Must(auth.NewJwtSessionManager(store, auth.NewJwtTokenProvider(config), refreshTTL))
}

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type activateBody struct {
	Email string `json:"email"`
	Token string `json:"token"`
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

	if _, err := a.users.Create(r.Context(), email, hash, false); err != nil {
		jw.WriteError(err)
		return
	}

	token, err := a.activations.Issue(r.Context(), "activate:"+email)
	if err != nil {
		jw.WriteError(err)
		return
	}

	link := fmt.Sprintf("http://localhost:3000/activate?email=%s&token=%s", url.QueryEscape(email), token)
	shared.Deliver(email, "activate your account: "+link)
	jw.WithSuccessStatus(http.StatusCreated).WriteSuccessMessage("check your email to activate your account", nil)
}

func (a app) activate(w http.ResponseWriter, r *http.Request) {
	jw := httputil.NewJsonResponseWriter(w)
	body, err := shared.Decode[activateBody](r)
	if err != nil {
		jw.WriteError(err)
		return
	}

	email, err := shared.ParseEmail(body.Email)
	if err != nil {
		jw.WriteError(err)
		return
	}

	if err := a.activations.Verify(r.Context(), "activate:"+email, body.Token); err != nil {
		jw.WriteError(err)
		return
	}

	err = a.users.Activate(r.Context(), email)
	jw.WriteSuccessMessage("account activated", err)
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

	// after the password check, so only the owner learns the account is inactive
	if !user.Active {
		jw.WriteError(errors.ForbiddenError("account not activated"))
		return
	}

	tokens, err := a.sessions.Issue(r.Context(), user.Id, nil)
	jw.WriteJsonResponse(tokens, err)
}

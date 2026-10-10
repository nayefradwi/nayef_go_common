// Reset password: request a link by email, confirm it with a new password, end every session.
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
	resets    auth.OneTimeTokenManager
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
	codeStore := shared.Must(authpg.NewCodeStore(pool))       // swap: authredis.NewCodeStore(rdb)
	attemptStore := authpg.NewAttemptStore(pool)              // swap: authredis.NewAttemptStore(rdb)

	sessions := opaqueSessions(sessionStore) // swap: jwtSessions(sessionStore, os.Getenv("JWT_SECRET"))
	requests := shared.Must(auth.NewLimiter(attemptStore, 20, time.Hour))
	hashing := auth.DefaultHashingConfig
	a := app{
		users:    shared.NewUsers(pool),
		sessions: sessions,
		resets: shared.Must(auth.NewOneTimeTokenManager(codeStore, auth.CodeLimits{
			TTL: 30 * time.Minute, ResendAfter: time.Minute, MaxSends: 3, MaxAttempts: 5,
		})),
		policy:  auth.DefaultPasswordPolicy,
		hashing: hashing,
		// checked against for unknown emails, so a miss costs the same bcrypt time as a hit
		dummyHash: shared.Must(hashing.Hash(rand.Text())),
	}

	r := chi.NewRouter()
	r.Post("/signup", a.signup)
	r.Post("/login", a.login)
	r.Post("/refresh", a.refresh)
	r.With(authhttp.RateLimit(requests, authhttp.ByIP("reset"))).Post("/password/reset", a.requestReset)
	r.Post("/password/reset/confirm", a.confirmReset)

	shared.Serve(r)
}

const (
	accessTTL  = 15 * time.Minute
	refreshTTL = 30 * 24 * time.Hour
)

// opaque access tokens are looked up on every request, so revoking ends them at once
func opaqueSessions(store auth.SessionStore) auth.SessionManager {
	return shared.Must(auth.NewOpaqueSessionManager(store, accessTTL, refreshTTL))
}

// jwt access tokens are checked without a store read, so RevokeOwner only ends
// refresh tokens; access tokens already issued live until accessTTL
func jwtSessions(store auth.SessionStore, secret string) auth.SessionManager {
	config := shared.Must(auth.NewJwtTokenProviderConfig(secret, accessTTL, auth.AccessTokenType))
	return shared.Must(auth.NewJwtSessionManager(store, auth.NewJwtTokenProvider(config), refreshTTL))
}

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshBody struct {
	RefreshToken string `json:"refresh_token"`
}

type resetBody struct {
	Email    string `json:"email"`
	Token    string `json:"token"`
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

// answers the same whether or not the email has an account, so it can't be used to find accounts
func (a app) requestReset(w http.ResponseWriter, r *http.Request) {
	jw := httputil.NewJsonResponseWriter(w)
	body, err := shared.Decode[resetBody](r)
	if err != nil {
		jw.WriteError(err)
		return
	}

	email, err := shared.ParseEmail(body.Email)
	if err != nil {
		jw.WriteError(err)
		return
	}

	const sent = "if that email has an account, a reset link is on its way"
	_, err = a.users.ByEmail(r.Context(), email)
	if errors.Is(err, shared.ErrUserNotFound) {
		jw.WithSuccessStatus(http.StatusAccepted).WriteSuccessMessage(sent, nil)
		return
	}

	if err != nil {
		jw.WriteError(err)
		return
	}

	token, err := a.resets.Issue(r.Context(), "reset:"+email)
	if errors.Is(err, auth.ErrCodeResendBlocked) {
		jw.WithSuccessStatus(http.StatusAccepted).WriteSuccessMessage(sent, nil)
		return
	}

	if err != nil {
		jw.WriteError(err)
		return
	}

	link := fmt.Sprintf("http://localhost:3000/reset?email=%s&token=%s", url.QueryEscape(email), token)
	shared.Deliver(email, "reset your password: "+link)
	jw.WithSuccessStatus(http.StatusAccepted).WriteSuccessMessage(sent, nil)
}

func (a app) confirmReset(w http.ResponseWriter, r *http.Request) {
	jw := httputil.NewJsonResponseWriter(w)
	body, err := shared.Decode[resetBody](r)
	if err != nil {
		jw.WriteError(err)
		return
	}

	email, err := shared.ParseEmail(body.Email)
	if err != nil {
		jw.WriteError(err)
		return
	}

	// checked before Verify so a rejected password doesn't burn the link
	if err := a.policy.Validate(body.Password); err != nil {
		jw.WriteError(err)
		return
	}

	if err := a.resets.Verify(r.Context(), "reset:"+email, body.Token); err != nil {
		jw.WriteError(err)
		return
	}

	user, err := a.users.ByEmail(r.Context(), email)
	if err != nil {
		jw.WriteError(err)
		return
	}

	hash, err := a.hashing.Hash(body.Password)
	if err != nil {
		jw.WriteError(err)
		return
	}

	if err := a.users.SetPassword(r.Context(), user.Id, hash); err != nil {
		jw.WriteError(err)
		return
	}

	// whoever knew the old password may still hold a session
	err = a.sessions.RevokeOwner(r.Context(), user.Id)
	jw.WriteSuccessMessage("password updated, log in again", err)
}

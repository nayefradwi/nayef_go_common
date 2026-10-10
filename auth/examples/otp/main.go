// Passwordless login: request a code by email, verify it for a session.
package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nayefradwi/nayef_go_common/auth"
	"github.com/nayefradwi/nayef_go_common/auth/authhttp"
	"github.com/nayefradwi/nayef_go_common/auth/authpg"
	"github.com/nayefradwi/nayef_go_common/auth/examples/shared"
	"github.com/nayefradwi/nayef_go_common/errors"
	"github.com/nayefradwi/nayef_go_common/httputil"
)

type app struct {
	users    shared.Users
	sessions auth.SessionManager
	codes    auth.OtpManager
}

func main() {
	ctx := context.Background()
	pool := shared.Connect(ctx)
	defer pool.Close()

	// redis instead of postgres: rdb := shared.Redis(ctx), then the authredis swap on each store
	sessionStore := shared.Must(authpg.NewSessionStore(pool)) // swap: authredis.NewSessionStore(rdb)
	codeStore := shared.Must(authpg.NewCodeStore(pool))       // swap: authredis.NewCodeStore(rdb)
	attemptStore := authpg.NewAttemptStore(pool)              // swap: authredis.NewAttemptStore(rdb)

	sessions, authenticate := opaqueSessions(sessionStore) // swap: jwtSessions(sessionStore, os.Getenv("JWT_SECRET"))
	requests := shared.Must(auth.NewLimiter(attemptStore, 20, time.Hour))
	a := app{
		users:    shared.NewUsers(pool),
		sessions: sessions,
		codes: shared.Must(auth.NewOtpManager(codeStore, auth.OtpConfig{
			Secret:     []byte(os.Getenv("OTP_SECRET")),
			Length:     6,
			CodeLimits: auth.CodeLimits{TTL: 10 * time.Minute, ResendAfter: time.Minute, MaxSends: 5, MaxAttempts: 5},
		})),
	}

	r := chi.NewRouter()
	// the code's resend cap guards one email; the IP limit stops one client spraying many
	r.With(authhttp.RateLimit(requests, authhttp.ByIP("otp"))).Post("/otp/request", a.requestCode)
	r.Post("/otp/verify", a.verifyCode)
	r.With(authenticate).Get("/me", a.me)

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

type codeBody struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

func (a app) requestCode(w http.ResponseWriter, r *http.Request) {
	jw := httputil.NewJsonResponseWriter(w)
	body, err := shared.Decode[codeBody](r)
	if err != nil {
		jw.WriteError(err)
		return
	}

	email, err := shared.ParseEmail(body.Email)
	if err != nil {
		jw.WriteError(err)
		return
	}

	code, err := a.codes.Issue(r.Context(), "login:"+email)
	if errors.Is(err, auth.ErrCodeResendBlocked) {
		jw.WriteError(shared.ErrTooManyRequests)
		return
	}

	if err != nil {
		jw.WriteError(err)
		return
	}

	shared.Deliver(email, "your login code is "+code)
	jw.WithSuccessStatus(http.StatusAccepted).WriteSuccessMessage("code sent", nil)
}

func (a app) verifyCode(w http.ResponseWriter, r *http.Request) {
	jw := httputil.NewJsonResponseWriter(w)
	body, err := shared.Decode[codeBody](r)
	if err != nil {
		jw.WriteError(err)
		return
	}

	email, err := shared.ParseEmail(body.Email)
	if err != nil {
		jw.WriteError(err)
		return
	}

	if err := a.codes.Verify(r.Context(), "login:"+email, body.Code); err != nil {
		jw.WriteError(err)
		return
	}

	// first verified code creates the account
	user, err := a.users.FindOrCreate(r.Context(), email)
	if err != nil {
		jw.WriteError(err)
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

// TOTP: enroll, confirm (→ recovery codes), two-step login with a pending token,
// recovery code login, disable.
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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

var errInvalidLogin = errors.UnauthorizedError("invalid email or password")

type app struct {
	users     shared.Users
	sessions  auth.SessionManager
	totp      auth.TotpManager
	pending   auth.OneTimeTokenManager
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
	totpStore := shared.Must(authpg.NewTotpStore(pool))       // swap: authredis.NewTotpStore(rdb)
	codeStore := shared.Must(authpg.NewCodeStore(pool))       // swap: authredis.NewCodeStore(rdb)

	sessions, authenticate := opaqueSessions(sessionStore) // swap: jwtSessions(sessionStore, os.Getenv("JWT_SECRET"))
	hashing := auth.DefaultHashingConfig
	a := app{
		users:    shared.NewUsers(pool),
		sessions: sessions,
		// 32 random bytes, hex encoded: openssl rand -hex 32
		totp: shared.Must(auth.NewTotpManager(totpStore, auth.TotpConfig{
			Key:    shared.Must(hex.DecodeString(os.Getenv("TOTP_KEY"))),
			Issuer: "auth-examples",
		})),
		// MaxAttempts caps code guesses per password login; MaxSends caps logins per TTL
		pending: shared.Must(auth.NewOneTimeTokenManager(codeStore, auth.CodeLimits{
			TTL: 5 * time.Minute, ResendAfter: 0, MaxSends: 5, MaxAttempts: 5,
		})),
		policy:  auth.DefaultPasswordPolicy,
		hashing: hashing,
		// checked against for unknown emails, so a miss costs the same bcrypt time as a hit
		dummyHash: shared.Must(hashing.Hash(rand.Text())),
	}

	r := chi.NewRouter()
	r.Post("/signup", a.signup)
	r.Post("/login", a.login)
	r.Post("/mfa/challenge", a.challenge)
	r.Post("/mfa/recover", a.recover)
	r.Group(func(r chi.Router) {
		r.Use(authenticate)
		r.Post("/mfa/enroll", a.enroll)
		r.Post("/mfa/confirm", a.confirm)
		r.Post("/mfa/disable", a.disable)
	})

	shared.Serve(r)
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

type mfaBody struct {
	Email        string `json:"email"`
	MfaToken     string `json:"mfa_token"`
	Code         string `json:"code"`
	RecoveryCode string `json:"recovery_code"`
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

// with TOTP on, a correct password earns a short-lived pending token, not a session
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

	enabled, err := a.totp.Enabled(r.Context(), user.Id)
	if err != nil {
		jw.WriteError(err)
		return
	}

	if !enabled {
		tokens, err := a.sessions.Issue(r.Context(), user.Id, nil)
		jw.WriteJsonResponse(tokens, err)
		return
	}

	token, err := a.pending.Issue(r.Context(), "mfa:"+email)
	if errors.Is(err, auth.ErrCodeResendBlocked) {
		jw.WriteError(shared.ErrTooManyRequests)
		return
	}

	if err != nil {
		jw.WriteError(err)
		return
	}

	jw.WriteData(map[string]any{"mfa_required": true, "mfa_token": token})
}

// Check (counts an attempt) → TOTP → Consume: a wrong code keeps the pending token
// until its attempts run out, a right one spends it
func (a app) challenge(w http.ResponseWriter, r *http.Request) {
	jw := httputil.NewJsonResponseWriter(w)
	body, err := shared.Decode[mfaBody](r)
	if err != nil {
		jw.WriteError(err)
		return
	}

	email, err := shared.ParseEmail(body.Email)
	if err != nil {
		jw.WriteError(err)
		return
	}

	if err := a.pending.Check(r.Context(), "mfa:"+email, body.MfaToken); err != nil {
		jw.WriteError(err)
		return
	}

	user, err := a.users.ByEmail(r.Context(), email)
	if err != nil {
		jw.WriteError(err)
		return
	}

	if err := a.totp.Verify(r.Context(), user.Id, body.Code); err != nil {
		jw.WriteError(err)
		return
	}

	if err := a.pending.Consume(r.Context(), "mfa:"+email, body.MfaToken); err != nil {
		jw.WriteError(err)
		return
	}

	tokens, err := a.sessions.Issue(r.Context(), user.Id, nil)
	jw.WriteJsonResponse(tokens, err)
}

// same as challenge with a single-use recovery code in place of the TOTP code
func (a app) recover(w http.ResponseWriter, r *http.Request) {
	jw := httputil.NewJsonResponseWriter(w)
	body, err := shared.Decode[mfaBody](r)
	if err != nil {
		jw.WriteError(err)
		return
	}

	email, err := shared.ParseEmail(body.Email)
	if err != nil {
		jw.WriteError(err)
		return
	}

	if err := a.pending.Check(r.Context(), "mfa:"+email, body.MfaToken); err != nil {
		jw.WriteError(err)
		return
	}

	user, err := a.users.ByEmail(r.Context(), email)
	if err != nil {
		jw.WriteError(err)
		return
	}

	if err := a.totp.UseRecoveryCode(r.Context(), user.Id, body.RecoveryCode); err != nil {
		jw.WriteError(err)
		return
	}

	if err := a.pending.Consume(r.Context(), "mfa:"+email, body.MfaToken); err != nil {
		jw.WriteError(err)
		return
	}

	tokens, err := a.sessions.Issue(r.Context(), user.Id, nil)
	jw.WriteJsonResponse(tokens, err)
}

// returns the secret for manual entry and the otpauth:// uri for a QR code
func (a app) enroll(w http.ResponseWriter, r *http.Request) {
	jw := httputil.NewJsonResponseWriter(w)
	id, ok := auth.GetIdentity(r.Context())
	if !ok {
		jw.WriteError(errors.UnauthorizedError("not logged in"))
		return
	}

	user, err := a.users.ById(r.Context(), id.OwnerId)
	if err != nil {
		jw.WriteError(err)
		return
	}

	secret, uri, err := a.totp.Enroll(r.Context(), user.Id, user.Email)
	if errors.Is(err, auth.ErrTotpEnabled) {
		jw.WriteError(errors.BadRequestError("totp already enabled, disable it first"))
		return
	}

	if err != nil {
		jw.WriteError(err)
		return
	}

	jw.WriteData(map[string]any{"secret": secret, "uri": uri})
}

// the recovery codes are in this response only
func (a app) confirm(w http.ResponseWriter, r *http.Request) {
	jw := httputil.NewJsonResponseWriter(w)
	id, ok := auth.GetIdentity(r.Context())
	if !ok {
		jw.WriteError(errors.UnauthorizedError("not logged in"))
		return
	}

	body, err := shared.Decode[mfaBody](r)
	if err != nil {
		jw.WriteError(err)
		return
	}

	codes, err := a.totp.Confirm(r.Context(), id.OwnerId, body.Code)
	if err != nil {
		jw.WriteError(err)
		return
	}

	jw.WriteData(map[string]any{"recovery_codes": codes})
}

// asks for a current code so a stolen session alone can't turn TOTP off
func (a app) disable(w http.ResponseWriter, r *http.Request) {
	jw := httputil.NewJsonResponseWriter(w)
	id, ok := auth.GetIdentity(r.Context())
	if !ok {
		jw.WriteError(errors.UnauthorizedError("not logged in"))
		return
	}

	body, err := shared.Decode[mfaBody](r)
	if err != nil {
		jw.WriteError(err)
		return
	}

	if err := a.totp.Verify(r.Context(), id.OwnerId, body.Code); err != nil {
		jw.WriteError(err)
		return
	}

	err = a.totp.Disable(r.Context(), id.OwnerId)
	jw.WriteSuccessMessage("totp disabled", err)
}

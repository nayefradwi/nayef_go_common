# auth

Building blocks for auth. No flows: you combine them in your handlers. Each block that keeps state takes a store interface; `authpg` and `authredis` implement them all.

```
go get github.com/nayefradwi/nayef_go_common/auth
```

Runnable flows that use every block: [examples](examples).

| Block | Type | Store |
|---|---|---|
| Passwords | `HashingConfig`, `PasswordPolicy`, `CompareHash` | — |
| Sessions | `SessionManager` (opaque or JWT access, rotating refresh) | `SessionStore` |
| JWT | `JwtTokenProvider` | — |
| Rate limits | `Limiter` | `AttemptStore` |
| Codes by email/SMS | `OtpManager` | `CodeStore` |
| Links and pending steps | `OneTimeTokenManager` | `CodeStore` |
| TOTP + recovery codes | `TotpManager` | `TotpStore` |
| API keys | `ApiKeyManager` | `ApiKeyStore` |
| Who is calling | `Identity`, `GetIdentity` | — |

## Passwords

```go
if err := auth.DefaultPasswordPolicy.Validate(pw); err != nil { ... } // 8+ chars, ≤ 72 bytes
hash, err := auth.DefaultHashingConfig.Hash(pw)
ok := auth.CompareHash(pw, hash)
```

## Sessions

Opaque access tokens are looked up on every request, so logout ends them at once. JWT access tokens skip the lookup and live until they expire. Refresh tokens are always opaque and rotate on use; a spent one coming back ends the whole login.

```go
sessions, err := auth.NewOpaqueSessionManager(store, 15*time.Minute, 30*24*time.Hour)

// or JWT access tokens
config, err := auth.NewJwtTokenProviderConfig(secret, 15*time.Minute, auth.AccessTokenType) // secret ≥ 32 bytes
sessions, err := auth.NewJwtSessionManager(store, auth.NewJwtTokenProvider(config), 30*24*time.Hour)

tokens, err := sessions.Issue(ctx, userId, map[string]any{"role": "admin"})
tokens, err = sessions.Refresh(ctx, tokens.RefreshToken)
err = sessions.Revoke(ctx, refreshToken) // one login
err = sessions.RevokeOwner(ctx, userId)  // every login
```

Claims are copied at `Issue` and kept through every refresh. A role change shows up at the next login.

## Codes, links, pending steps

```go
otps, err := auth.NewOtpManager(store, auth.OtpConfig{
	Secret:     secret, // ≥ 32 bytes
	Length:     6,
	CodeLimits: auth.CodeLimits{TTL: 10 * time.Minute, ResendAfter: time.Minute, MaxSends: 5, MaxAttempts: 5},
})
code, err := otps.Issue(ctx, "login:"+email) // auth.ErrCodeResendBlocked → 429
err = otps.Verify(ctx, "login:"+email, code)

resets, err := auth.NewOneTimeTokenManager(store, limits)
token, err := resets.Issue(ctx, "reset:"+email)
err = resets.Verify(ctx, "reset:"+email, token)
```

`Verify` is `Check` (counts an attempt) then `Consume` (deletes). Call them apart for a second step, like MFA: `Check` the pending token, check the TOTP code, then `Consume`.

## TOTP

```go
totp, err := auth.NewTotpManager(store, auth.TotpConfig{Key: key32, Issuer: "myapp"}) // secrets are AES-GCM sealed with Key
secret, uri, err := totp.Enroll(ctx, userId, email) // uri → QR code
recoveryCodes, err := totp.Confirm(ctx, userId, code)
err = totp.Verify(ctx, userId, code) // each code works once
err = totp.UseRecoveryCode(ctx, userId, recoveryCode)
on, err := totp.Enabled(ctx, userId)
```

## API keys

```go
keys, err := auth.NewApiKeyManager(store, "sk_live")
raw, key, err := keys.Issue(ctx, ownerId, "ci", map[string]any{"scopes": []string{"reports:read"}}, time.Time{}) // raw is shown once; zero time = never expires
id, err := keys.VerifyKey(ctx, raw) // Identity{Kind: KindKey}
err = keys.Revoke(ctx, ownerId, key.Id) // owner-scoped
```

## Rate limits

```go
limiter, err := auth.NewLimiter(store, 5, 15*time.Minute)
allowed, retryAfter, err := limiter.Allow(ctx, "login:"+email)
```

## Your own access tokens

Anything with `VerifyAccess(ctx, raw) (Token, error)` works with the opaque middleware and interceptor. `SessionManager` is the built-in one. A custom one must set `ExpiresAt` (zero counts as expired), and its identity comes out as `KindUser`. For example, a single token from env:

```go
type envToken struct {
	token string
	owner uuid.UUID
}

func (e envToken) VerifyAccess(_ context.Context, raw string) (auth.Token, error) {
	if subtle.ConstantTimeCompare([]byte(raw), []byte(e.token)) != 1 {
		return auth.Token{}, errors.UnauthorizedError("Invalid token")
	}
	return auth.Token{OwnerId: e.owner, ExpiresAt: time.Now().Add(time.Minute)}, nil
}

r.Use(authhttp.NewOpaqueMiddleware(envToken{os.Getenv("ADMIN_TOKEN"), adminId}).UseAuthentication)
```

## Cleanup

Every manager has `DeleteExpired(ctx)`. Nothing calls it; schedule it yourself (a no-op on authredis).

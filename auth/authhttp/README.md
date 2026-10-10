# authhttp

net/http middleware for [auth](..): read the bearer token, put an `auth.Identity` on the context, then gate routes with `Has(check)`.

```
go get github.com/nayefradwi/nayef_go_common/auth/authhttp
```

## Who can call

| Middleware | Accepts |
|---|---|
| `NewOpaqueMiddleware(sessions)` | users, opaque access tokens (any `AccessVerifier`) |
| `NewJwtMiddleware(provider)` | users, JWT access tokens |
| `NewApiKeyMiddleware(keys)` | API keys only |
| `NewEitherMiddleware(sessions, keys)` | users or keys, split by the key prefix with no extra read |
| `NewJwtEitherMiddleware(provider, keys)` | same, JWT users |

A missing or bad token is a 401. Each has `.UseAuthentication`.

## What they can do

A check is a func you write: `func(ctx, auth.Identity, *http.Request) (bool, error)`. `Has(check)` returns 403 on false and 500 on error. It can read claims, query a DB, or read an id from the URL.

```go
r.Group(func(r chi.Router) {
	r.Use(authhttp.NewEitherMiddleware(sessions, keys).UseAuthentication)
	r.Get("/reports", listReports)                                         // users and keys
	r.With(authhttp.Has(authhttp.IsKey)).Post("/reports/ingest", ingest)   // keys only
	r.With(authhttp.Has(authhttp.ByKind(authhttp.IsUser, authhttp.HasClaim("scopes", "reports:write")))).
		Post("/reports", createReport)                                     // users, or keys with the scope

	r.Group(func(r chi.Router) {
		r.Use(authhttp.VerifyClaim("role", "admin")) // = Has(HasClaim("role", "admin"))
		r.Get("/admin", admin)
	})
})
```

In a handler: `id, ok := auth.GetIdentity(r.Context())`.

## Rate limits

```go
r.With(authhttp.RateLimit(limiter, authhttp.ByIP("otp"))).Post("/otp/request", requestCode)
```

Over the limit is a 429 with `Retry-After`. `ByIP` reads `r.RemoteAddr`; behind a proxy, mount a real-IP middleware first.

# authconnect

Connect interceptors for [auth](..). Same API as [authhttp](../authhttp): read the bearer token, put an `auth.Identity` on the context, gate with `Has(check)`.

```
go get github.com/nayefradwi/nayef_go_common/auth/authconnect
```

| Interceptor | Accepts |
|---|---|
| `NewOpaqueInterceptor(sessions)` | users, opaque access tokens (any `AccessVerifier`) |
| `NewJwtInterceptor(provider)` | users, JWT access tokens |
| `NewApiKeyInterceptor(keys)` | API keys only |
| `NewEitherInterceptor(sessions, keys)` | users or keys, split by the key prefix |
| `NewJwtEitherInterceptor(provider, keys)` | same, JWT users |

A missing or bad token is `Unauthenticated`. A check is `func(ctx, auth.Identity, connect.AnyRequest) (bool, error)`; `Has` returns `PermissionDenied` on false and `Internal` on error. `HasClaim`, `VerifyClaim`, `IsUser`, `IsKey` and `ByKind` work as in authhttp.

## Layering

Interceptors run in the order they are added, so put the base options first and narrow per service:

```go
base := connect.WithHandlerOptions(
	connect.WithInterceptors(authconnect.NewEitherInterceptor(sessions, keys)),
)
adminOnly := connect.WithHandlerOptions(base, connect.WithInterceptors(authconnect.VerifyClaim("role", "admin")))

mux.Handle(reportsv1connect.NewReportsServiceHandler(reports, base))
mux.Handle(adminv1connect.NewAdminServiceHandler(admin, adminOnly))
```

## Rate limits

```go
connect.WithInterceptors(authconnect.RateLimit(limiter, authconnect.ByIP("otp")))
```

Over the limit is `ResourceExhausted` with `Retry-After`.

## Streams

Auth interceptors cover streaming handlers too. `Has` and `RateLimit` are unary only: on a streaming RPC they fail closed (`PermissionDenied` / `Unimplemented`), so check streams inside the handler.

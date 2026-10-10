# auth examples

One runnable server per flow. Each `main.go` reads top to bottom: wiring, routes, handlers in flow order. Every auth call is in the flow's own file; `shared/` only holds plumbing that isn't about auth (DB connect + migrations, users SQL, JSON decode, a log stub for email).

| Flow | Shows |
|---|---|
| [password](password) | signup, login (rate limited), refresh, logout, logout-all, me |
| [otp](otp) | passwordless login with an emailed code |
| [reset](reset) | password reset link, end every session |
| [activation](activation) | inactive until the emailed link is confirmed |
| [rbac-claims](rbac-claims) | role and scopes in the session claims |
| [rbac-db](rbac-db) | permissions loaded per request from a role table |
| [apikeys](apikeys) | user-managed API keys, key-only and mixed routes |
| [totp](totp) | TOTP enroll/confirm, two-step login, recovery codes, disable |

## Run

```sh
docker run -d --name auth-examples-pg -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=auth_examples -p 5432:5432 postgres:16-alpine
go run ./password
```

| Env | Default | Used by |
|---|---|---|
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/auth_examples?sslmode=disable` | all |
| `ADDR` | `:8080` | all |
| `OTP_SECRET` | none, ≥ 32 bytes: `openssl rand -hex 32` | otp |
| `TOTP_KEY` | none, 32 bytes hex: `openssl rand -hex 32` | totp |
| `JWT_SECRET` | none, ≥ 32 bytes, only after the jwt swap | any |
| `REDIS_URL` | `redis://localhost:6379`, only after the redis swap | any |

Codes and links that would be emailed are printed in the server log (`msg=deliver`). The walkthroughs use `jq`.

## Swaps

Each choice is one line in `main.go` with its alternative in a `// swap:` comment. Handlers never change.

- **Opaque ↔ JWT access tokens:** `opaqueSessions(sessionStore)` → `jwtSessions(sessionStore, os.Getenv("JWT_SECRET"))`. Both functions sit under `main`, since the swap changes the manager and its middleware together. JWT access tokens are checked without a store read, so they outlive logout until they expire (15 min).
- **Postgres ↔ Redis stores:** uncomment `rdb := shared.Redis(ctx)` and switch each `authpg.NewXStore(pool)` to `authredis.NewXStore(rdb)`. The users table stays in Postgres.

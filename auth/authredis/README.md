# authredis

Redis stores for every [auth](..) store interface. Drop-in for [authpg](../authpg): same managers, swap the constructor.

```
go get github.com/nayefradwi/nayef_go_common/auth/authredis
```

```go
sessionStore, err := authredis.NewSessionStore(client)
attemptStore := authredis.NewAttemptStore(client)
codeStore, err := authredis.NewCodeStore(client)
apiKeyStore, err := authredis.NewApiKeyStore(client)
totpStore, err := authredis.NewTotpStore(client)
```

- Takes a `*redis.Client` (go-redis v9), which covers Sentinel. Needs Redis 7+. No Cluster: the Lua scripts touch related keys together.
- Expiry is Redis TTL, so `DeleteExpired` is a no-op and there are no migrations.
- Keys live under `auth:sessions:`, `auth:otps:`, `auth:api_keys:`, `auth:totp:` and `auth:attempts:`. Change a prefix with the store's config, e.g. `authredis.SessionStoreConfig{Prefix: "admin:sessions:"}` (lowercase, digits, `_` and `:`).
- Checks and writes that must happen together run as one Lua script, timed by the Redis clock.

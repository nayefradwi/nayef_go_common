# authpg

Postgres stores for every [auth](..) store interface, with their migrations. Built on pgx.

```
go get github.com/nayefradwi/nayef_go_common/auth/authpg
```

```go
sessionStore, err := authpg.NewSessionStore(pool)
attemptStore := authpg.NewAttemptStore(pool)
codeStore, err := authpg.NewCodeStore(pool)
apiKeyStore, err := authpg.NewApiKeyStore(pool)
totpStore, err := authpg.NewTotpStore(pool)

sessions, err := auth.NewOpaqueSessionManager(sessionStore, 15*time.Minute, 30*24*time.Hour)
```

## Migrations

The schema ships as goose files in `authpg.Migrations` (`embed.FS`), named `authpg_0002_attempts.sql` … `authpg_0006_totp.sql`. Copy them into your migrations folder with a goose timestamp in front (`20261009120000_authpg_0003_sessions.sql`) so they sort with yours. A release that changes the schema adds a new numbered file.

## Your own table names

Sessions, codes, API keys and TOTP take a table name. Render the matching migration with it, so the columns can't drift:

```go
sql, err := authpg.SessionMigration("admin_sessions") // write into your migrations folder
store, err := authpg.NewSessionStore(pool, authpg.SessionStoreConfig{Table: "admin_sessions"})
```

Same for `CodeMigration` + `CodeStoreConfig`, `ApiKeyMigration` + `ApiKeyStoreConfig`, `TotpMigration` + `TotpStoreConfig`. Names must be plain identifiers.

## Transactions

Stores never commit or roll back. `SessionStore` and `ApiKeyStore` join your transaction with `WithTx(tx)`. `AttemptStore`, `CodeStore` and `TotpStore` always use the pool, so your rollback can't undo a failed attempt or a used code.

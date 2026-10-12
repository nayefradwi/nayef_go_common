package auth

const (
	INTERNAL    = "internal"
	AUTH        = "auth"
	INFRA       = "infra"
	SQLC        = "sqlc"
	QUERIES     = "queries"
	MIGRATIONS  = "migrations"
	PROTO       = "proto"
	V1          = "v1"
	DEPLOYMENTS = "deployments"
	LOCAL       = "local"
	ENV         = ".env"
	ENV_EXAMPLE = ".env.example"
)

const (
	TMPL_CONFIG       = "config.go.tmpl"
	TMPL_HANDLER      = "handler_rest.go.tmpl"
	TMPL_GRPC_HANDLER = "handler_grpc.go.tmpl"
	TMPL_SERVICE      = "service.go.tmpl"
	TMPL_REQUEST      = "request.go.tmpl"
	TMPL_RESPONSE     = "response.go.tmpl"
	TMPL_ERRORS       = "errors.go.tmpl"
	TMPL_PROTO        = "auth.proto.tmpl"
	TMPL_QUERIES      = "users.sql.tmpl"
	TMPL_USERS_TABLE  = "users_table.sql.tmpl"
	TMPL_TODO         = "TODO.md.tmpl"
)

// Pinned so the copied authpg schema matches the store code the project runs;
// bump together with the authpg require in ngo's go.mod.
const AUTH_VERSION = "v1.0.1"

const (
	AUTH_PKG        = "github.com/nayefradwi/nayef_go_common/auth@" + AUTH_VERSION
	AUTHPG_PKG      = "github.com/nayefradwi/nayef_go_common/auth/authpg@" + AUTH_VERSION
	AUTHREDIS_PKG   = "github.com/nayefradwi/nayef_go_common/auth/authredis@" + AUTH_VERSION
	AUTHHTTP_PKG    = "github.com/nayefradwi/nayef_go_common/auth/authhttp@" + AUTH_VERSION
	AUTHCONNECT_PKG = "github.com/nayefradwi/nayef_go_common/auth/authconnect@" + AUTH_VERSION
)

const (
	// first pgutil release with UUID / UUIDFromPg
	PGUTIL_PKG     = "github.com/nayefradwi/nayef_go_common/pgutil@v0.1.10"
	VALIDATION_PKG = "github.com/nayefradwi/nayef_go_common/validation@v0.1.5"
)

var authpgMigrations = []string{"authpg_0002_attempts.sql", "authpg_0003_sessions.sql", "authpg_0004_otps.sql"}

const AUTHPG_TOTP_MIGRATION = "authpg_0006_totp.sql"

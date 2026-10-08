package add

const (
	INTERNAL = "internal"
	SERVICE  = "service"
	HANDLER  = "handler"
	GO       = "go"
	PROTO    = "proto"
	V1       = "v1"
)

const (
	TMPL_SERVICE      = "service.go.tmpl"
	TMPL_HANDLER      = "handler.go.tmpl"
	TMPL_GRPC_HANDLER = "grpc_handler.go.tmpl"
	TMPL_PROTO        = "service.proto.tmpl"
)

const (
	DB_GEN_SUBPATH = "internal/infra/sqlc/gen"
)

const (
	PGXPOOL = "github.com/jackc/pgx/v5/pgxpool"
	REDIS   = "github.com/redis/go-redis/v9"
	LOCKING = "github.com/nayefradwi/nayef_go_common/locking"
	OTP     = "github.com/nayefradwi/nayef_go_common/otp"
)

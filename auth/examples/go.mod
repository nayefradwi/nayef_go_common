module github.com/nayefradwi/nayef_go_common/auth/examples

go 1.27.0

// never tagged: builds against the auth modules in this repo
replace (
	github.com/nayefradwi/nayef_go_common/auth => ../
	github.com/nayefradwi/nayef_go_common/auth/authhttp => ../authhttp
	github.com/nayefradwi/nayef_go_common/auth/authpg => ../authpg
	github.com/nayefradwi/nayef_go_common/auth/authredis => ../authredis
)

require (
	github.com/go-chi/chi/v5 v5.3.2
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.11.0
	github.com/nayefradwi/nayef_go_common/auth v1.0.0
	github.com/nayefradwi/nayef_go_common/auth/authhttp v0.0.0-00010101000000-000000000000
	github.com/nayefradwi/nayef_go_common/auth/authpg v0.0.0-00010101000000-000000000000
	github.com/nayefradwi/nayef_go_common/auth/authredis v0.0.0-00010101000000-000000000000
	github.com/nayefradwi/nayef_go_common/errors v1.0.8
	github.com/nayefradwi/nayef_go_common/httputil v1.0.7
	github.com/nayefradwi/nayef_go_common/pgutil v0.1.9
	github.com/nayefradwi/nayef_go_common/redisutil v0.1.4
	github.com/redis/go-redis/v9 v9.23.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	go.uber.org/atomic v1.12.0 // indirect
	golang.org/x/crypto v0.54.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/text v0.40.0 // indirect
)

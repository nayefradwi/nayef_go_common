# Restructuring Plan for nayef_go_common

## Current Problems

- All packages are part of a single Go module — consumers pull every dependency (pgx, redis, grpc, jwt, zap, protobuf) even if they only need validation
- The `modules/` vs `services/` split adds indirection with no benefit; they're the same concern split across two packages for no reason
- `result` package has grpc and protobuf as direct dependencies, meaning REST-only services transitively import grpc
- Global mutable error listeners (`GlobalJsonWriterOnErrorListener`, `GlobalWriterOnErrorListener`) make tests non-deterministic and couple unrelated code
- Several packages add complexity instead of reducing it (see below)
- Critical bugs in existing implementations
- No tests except for validation

---

## What to Delete

These packages should be removed entirely. They either duplicate stdlib, are broken, or save no meaningful developer time.

### `modules/functional`
`Map`, `Filter`, `FirstWhere` are 3 trivial functions. Since Go 1.21, the stdlib `slices` package covers these. Any Go developer can write a 2-line for loop faster than looking up this package.

### `modules/collections`
- `GetValues` → `maps.Values()` (stdlib since Go 1.21)
- `MergeSlice` → `slices.Concat()` (stdlib since Go 1.22)
- `MergeMaps` → `maps.Copy()` (stdlib since Go 1.21)
- `StructToMap` → JSON marshal/unmarshal round-trip is an antipattern. Slow, lossy, and signals a design problem wherever it's used.

### `modules/dates`
Five functions wrapping `time.Format()` and `time.Date()`. `TimeToISO8601` is literally `t.Format(time.RFC3339)`. This hides trivial code behind unfamiliar names.

### `modules/env`
Parses `os.Args` for `flavor=` with string splitting — fragile and opinionated. Every project has its own config story. Calling `godotenv.Load()` directly is one line.

### `modules/locking/service.go` (InMemoryLocker only)
Fundamentally broken:
- `releaseAfter` spawns a goroutine calling `lock.Unlock()` after a TTL — if the lock was already released by the caller via `defer`, this double-unlocks and panics
- The `locks` map is accessed without synchronization — concurrent `AcquireLock` calls on different keys race

The `ILocker` interface is fine. Keep it. Delete the in-memory implementation.

### `modules/cache` + `services/cache`
An interface with no implementation and an empty file. Ship it when it's implemented.

### `services/email` + `services/logging`
Empty directories.

---

## Bugs to Fix Before or During Restructure

1. **OTP generator has swapped method names** — `generateAlphaNumeric()` only uses digits, `generateNumeric()` generates alphanumeric. They're backwards in `modules/otp/generator.go`.
2. **Ignored error in OTP service** — `services/otp/otp_service.go` calls `s.otpRepository.UpsertOtp(ctx, o)` and discards the error after incrementing retry count.
3. **Typos in exported function names** — `UseAuthenitcation` (×2) and `hanlder` (×2) in `services/rest/auth.middleware.go` and `services/rest/pagination.middleware.go`. Breaking change to fix later if not fixed now.

---

## New Structure: Multi-Module Monorepo

Each directory gets its own `go.mod`. Consumers import only the modules they need.

```
nayef_go_common/
├── result/                          # module: nayef_go_common/result
│   ├── error.go                     #   ResultError, ErrorDetails, error codes
│   └── factory.go                   #   BadRequestError(), NotFoundError(), etc.
│                                    #   ZERO external dependencies — pure Go
│
├── resultgrpc/                      # module: nayef_go_common/resultgrpc
│   ├── error.proto                  #   protobuf source (commit this)
│   ├── error.pb.go                  #   generated
│   └── convert.go                   #   ToGRPCError(), FromGRPCError(), code mappings
│                                    #   depends on: result, grpc, protobuf
│
├── validation/                      # module: nayef_go_common/validation
│   ├── validator.go
│   ├── string.go
│   ├── number.go
│   ├── date.go
│   ├── slice.go
│   └── validation_test.go           #   expand existing tests
│                                    #   depends on: result
│
├── pgutil/                          # module: nayef_go_common/pgutil
│   ├── connection.go                #   ConnectionConfig, Connect()
│   ├── errors.go                    #   MapPgError(), PG error code constants
│   └── tx.go                        #   Tx(), TxWithData()
│                                    #   depends on: result, pgx
│
├── auth/                            # module: nayef_go_common/auth
│   ├── jwt.go
│   ├── jwt_config.go
│   ├── jwt_signers.go
│   ├── jwt_parsers.go
│   ├── hash.go
│   ├── tokens.go
│   ├── dtos.go
│   ├── interfaces.go
│   ├── refresh_provider.go          #   merged from services/auth
│   ├── refresh_revoke_provider.go   #   merged from services/auth
│   └── reference_provider.go        #   merged from services/auth
│                                    #   depends on: result, jwt/v5, crypto
│
├── otp/                             # module: nayef_go_common/otp
│   ├── generator.go                 #   FIX: swap method names back
│   ├── model.go
│   ├── errors.go
│   ├── interfaces.go
│   ├── service.go                   #   FIX: handle upsert error
│   └── redis_repository.go          #   merged from services/otp
│                                    #   depends on: result, redis
│
├── pagination/                      # module: nayef_go_common/pagination
│   ├── offset.go
│   └── cursor.go
│                                    #   ZERO external dependencies — pure Go
│
├── httputil/                        # module: nayef_go_common/httputil
│   ├── writer.go                    #   JsonResponseWriter — no global listener
│   ├── parser.go                    #   ParseJsonBody
│   ├── recover.go                   #   Recover middleware
│   ├── utils.go                     #   GetBearerToken, GetIntQueryParam
│   ├── auth_middleware.go           #   merged from services/rest — FIX typos
│   └── pagination_middleware.go     #   merged from services/rest — FIX typos
│                                    #   depends on: result, auth, pagination
│
├── grpcutil/                        # module: nayef_go_common/grpcutil
│   ├── writer.go                    #   GrpcResponseWriter — no global listener
│   └── recover.go                   #   RecoverUnary interceptor
│                                    #   depends on: resultgrpc, grpc
│
├── redisutil/                       # module: nayef_go_common/redisutil
│   └── connection.go
│                                    #   depends on: go-redis
│
├── distlock/                        # module: nayef_go_common/distlock
│   ├── interface.go                 #   ILocker, LockParams — keep this
│   └── redsync.go                   #   DistributedLocker (redsync impl only)
│                                    #   depends on: result, redsync, go-redis
│
└── logging/                         # module: nayef_go_common/logging
    └── logger.go                    #   Zap + lumberjack setup
                                     #   depends on: zap, lumberjack
```

---

## Key Design Decisions

### `result` has zero external dependencies
Currently `result` imports grpc and protobuf because of `ToGRPCError()` and the `.pb.go` file. In the new structure, `result` is pure Go. REST-only services never touch protobuf. gRPC services import `resultgrpc`, which wraps `result` with the transport conversion.

### The modules/services split is gone
The current `modules/auth` + `services/auth` separation forces consumers to import two packages for one concern. The "service" is just a composition of the module types — it belongs in the same package. Same applies to rest/grpc/otp/locking.

### Global error listeners are removed
`GlobalJsonWriterOnErrorListener` and `GlobalWriterOnErrorListener` are mutable package-level state. They make tests non-deterministic and couple unrelated code. The `ErrorListener` field already exists on the writer structs — require explicit injection instead of relying on a global.

### `pagination` and `result` depend on nothing
Pure Go. No reason to pull any external dependency for models and context helpers.

---

## Dependency Graph

```
result (pure Go, zero deps)          pagination (pure Go, zero deps)
  |     \          \                       |
  |    validation  pgutil               httputil
  |                                    /    |
  |                   auth ───────────      |
  |                    |                    |
  |                   otp               grpcutil
  |                                        |
resultgrpc (grpc, protobuf) ───────────────┘

redisutil ← distlock
           ← otp
logging (standalone)
```

---

## Implementation Order

1. Create `result/` as standalone module (strip grpc out)
2. Create `resultgrpc/` with grpc conversion + commit the `.proto` source
3. Migrate `validation/` — easiest, already tested
4. Migrate `pgutil/` — solid, just needs new module
5. Migrate `auth/` — merge modules + services, no behavior change
6. Migrate `pagination/` — trivial, zero deps
7. Migrate `httputil/` — merge modules/rest + services/rest, fix typos
8. Migrate `grpcutil/` — merge modules/grpc, point at resultgrpc
9. Migrate `otp/` — merge modules + services, fix bugs
10. Migrate `distlock/` — drop InMemoryLocker, keep interface + redsync
11. Migrate `redisutil/` and `logging/`
12. Delete old structure
13. Add tests for all migrated modules

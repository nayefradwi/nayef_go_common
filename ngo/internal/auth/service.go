package auth

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/nayefradwi/nayef_go_common/errors"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/common"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/printer"
)

func scaffold(req CreateAuthRequest) error {
	runner := errors.ResultRunnerWithParam[CreateAuthRequest]{}
	runner.Do(req, checkPostgres)
	runner.Do(req, checkNotGenerated)
	runner.Do(req, checkGooseInstalled)
	runner.Do(req, checkBufInstalled)
	runner.Do(req, installGoPackages)
	runner.Do(req, generateCode)
	runner.Do(req, createMigration)
	runner.Do(req, appendEnv)
	runner.Do(req, writeProjectFile)
	runner.Do(req, runBufGenerate)
	runner.Do(req, printNextSteps)
	return runner.Error
}

func checkPostgres(req CreateAuthRequest) error {
	if !req.HasPostgres() {
		return fmt.Errorf("ngo auth needs postgres for the users table, and %s has none", common.PROJECT_FILE)
	}

	return nil
}

func checkNotGenerated(req CreateAuthRequest) error {
	for _, path := range []string{authDir(req), queriesPath(req)} {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("%s already exists, ngo auth only runs once", path)
		}
	}

	return nil
}

func checkGooseInstalled(req CreateAuthRequest) error {
	if _, err := exec.LookPath("goose"); err != nil {
		return fmt.Errorf("ngo auth needs the goose CLI on PATH, see https://github.com/pressly/goose#install: %w", err)
	}

	return nil
}

func checkBufInstalled(req CreateAuthRequest) error {
	if !req.IsGrpc() {
		return nil
	}

	return common.CheckBufInstalled()
}

func getPackages(req CreateAuthRequest) []string {
	packages := []string{AUTH_PKG, PGUTIL_PKG, VALIDATION_PKG, AUTHPG_PKG}
	if req.IsRedisStore() {
		packages = []string{AUTH_PKG, PGUTIL_PKG, VALIDATION_PKG, AUTHREDIS_PKG}
	}

	if req.IsGrpc() {
		return append(packages, AUTHCONNECT_PKG)
	}

	return append(packages, AUTHHTTP_PKG)
}

// No go mod tidy afterwards: the store packages are only imported by the di.go wiring
// the dev pastes later, so tidy would drop them.
func installGoPackages(req CreateAuthRequest) error {
	stop := printer.Spin("Installing auth packages")
	args := append([]string{"get"}, getPackages(req)...)
	cmd := exec.Command("go", args...)
	cmd.Dir = req.RootDirPath

	out, err := cmd.CombinedOutput()
	if err == nil {
		// go get bumps shared deps (pgx, x/sync) without recording their sums when only
		// authredis is fetched; download all fills go.sum so the project still builds
		download := exec.Command("go", "mod", "download", "all")
		download.Dir = req.RootDirPath
		out, err = download.CombinedOutput()
	}

	stop(err)
	if err != nil {
		return fmt.Errorf("failed to get auth packages %w: %s", err, out)
	}

	return nil
}

func generateCode(req CreateAuthRequest) error {
	runner := errors.ResultRunnerWithParam[CreateAuthRequest]{}
	runner.Do(req, renderConfig)
	runner.Do(req, renderModule)
	runner.Do(req, renderQueries)
	runner.Do(req, renderProto)
	return runner.Error
}

func writeProjectFile(req CreateAuthRequest) error {
	project := req.Project
	project.Auth = &req.Auth
	return common.WriteProject(req.RootDirPath, project)
}

func runBufGenerate(req CreateAuthRequest) error {
	if !req.IsGrpc() {
		return nil
	}

	return common.RunBufGenerate(req.RootDirPath)
}

func printNextSteps(req CreateAuthRequest) error {
	steps, err := nextSteps(req)
	if err != nil {
		return err
	}

	printer.Success("added auth")
	for _, line := range steps {
		printer.Info(line)
	}

	return nil
}

func nextSteps(req CreateAuthRequest) ([]string, error) {
	view := newAuthView(req)
	imports, err := renderNamed(TMPL_DI, "imports", view)
	if err != nil {
		return nil, err
	}

	wiring, err := renderNamed(TMPL_DI, "wiring", view)
	if err != nil {
		return nil, err
	}

	n := 0
	step := func(title string) string {
		n++
		return fmt.Sprintf("%d. %s", n, title)
	}

	steps := []string{"", step("generate the user queries:"), "   sqlc generate", ""}
	if len(envKeys(req)) > 0 {
		steps = append(steps,
			step("internal/config/config.go (import "+req.GoModule+"/internal/auth):"),
			"   add to Config:       Auth auth.AuthConfig",
			"   add to NewConfig():  Auth: auth.NewAuthConfig(),",
			"",
		)
	}

	middleware := "func(http.Handler) http.Handler"
	if req.IsGrpc() {
		middleware = "connect.Interceptor"
	}

	steps = append(steps,
		step("internal/di/di.go, add to Di:"),
		"   Auth          appauth.AuthService",
		"   Authenticate  "+middleware+" // use it on your own routes too",
		"   AuthRateLimit "+middleware,
		"   imports:",
	)
	steps = append(steps, indent(imports, "   ")...)
	steps = append(steps, "   add to RegisterServices() after the connects:")
	steps = append(steps, indent(wiring, "   ")...)
	steps = append(steps, "")

	if req.IsGrpc() {
		steps = append(steps,
			step("cmd/api/router.go (import appauth \""+req.GoModule+"/internal/auth\" and "+req.GoModule+"/gen/auth/v1/authv1connect), add to setup():"),
			"   authHandler := appauth.NewAuthHandler(ar.di.Auth)",
			"   ar.mux.Handle(authv1connect.NewAuthServiceHandler(authHandler, opts))",
			"   limited := connect.WithHandlerOptions(opts, connect.WithInterceptors(ar.di.AuthRateLimit))",
			"   ar.mux.Handle(authv1connect.NewCodeServiceHandler(authHandler, limited))",
			"   authed := connect.WithHandlerOptions(opts, connect.WithInterceptors(ar.di.Authenticate))",
			"   ar.mux.Handle(authv1connect.NewAccountServiceHandler(authHandler, authed))",
			"   use authed on your own services; narrow with",
			"   connect.WithInterceptors(authconnect.VerifyClaim(\"role\", \"admin\"))",
			"   internal/health/handler.go, add to NewStaticChecker:",
			"   authv1connect.AuthServiceName, authv1connect.CodeServiceName, authv1connect.AccountServiceName",
			"",
		)
	} else {
		steps = append(steps,
			step("cmd/api/router.go (import appauth \""+req.GoModule+"/internal/auth\"), add to setup():"),
			"   ar.mux.Mount(\"/auth\", appauth.NewAuthHandler(ar.di.Auth).Routes(ar.di.Authenticate, ar.di.AuthRateLimit))",
			"   protect your own routes:",
			"   ar.mux.Group(func(r chi.Router) {",
			"       r.Use(ar.di.Authenticate)",
			"       r.Use(authhttp.VerifyClaim(\"role\", \"admin\")) // optional, same as authhttp.Has(authhttp.HasClaim(\"role\", \"admin\"))",
			"   })",
			"",
		)
	}

	keys := envKeys(req)
	if len(keys) > 0 {
		steps = append(steps, step("deployments/{staging,production}/docker-compose.yaml, add to environment:"))
		for _, key := range keys {
			steps = append(steps, "      - "+key+"=${"+key+"}")
		}
		steps = append(steps, "   .env and deployments/local/.env already have generated values", "")
	}

	steps = append(steps, step("apply the migration:"), "   goose up", "")
	if req.IsRedisStore() {
		steps = append(steps,
			step("redis holds sessions, codes, attempts and TOTP secrets: it needs Redis 7+ (no Cluster),"),
			"   persistence on, and an eviction policy that never drops keys (noeviction), or users lose MFA",
			"",
		)
	} else {
		steps = append(steps,
			step("schedule cleanup of expired rows: DeleteExpired on the session, code and attempt stores"),
			"",
		)
	}

	steps = append(steps, step("fill the TODOs: grep -rn TODO internal/auth ("+strings.Join(todos(req), ", ")+")"))
	return steps, nil
}

func indent(block, prefix string) []string {
	lines := strings.Split(strings.TrimSpace(block), "\n")
	for i, line := range lines {
		lines[i] = prefix + line
	}

	return lines
}

func todos(req CreateAuthRequest) []string {
	if req.IsPassword() {
		return []string{"send reset tokens", "claims"}
	}

	return []string{"send login codes", "claims"}
}

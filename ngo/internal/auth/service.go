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
	runner.Do(req, checkSqlcInstalled)
	runner.Do(req, checkBufInstalled)
	runner.Do(req, installGoPackages)
	runner.Do(req, generateCode)
	runner.Do(req, createMigration)
	runner.Do(req, runSqlcGenerate)
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

func checkSqlcInstalled(req CreateAuthRequest) error {
	if _, err := exec.LookPath("sqlc"); err != nil {
		return fmt.Errorf("ngo auth needs the sqlc CLI on PATH, see https://docs.sqlc.dev/en/latest/overview/install.html: %w", err)
	}

	return nil
}

// runs after createMigration: sqlc reads the schema from the migrations
func runSqlcGenerate(req CreateAuthRequest) error {
	stop := printer.Spin("Generating user queries")
	cmd := exec.Command("sqlc", "generate")
	cmd.Dir = req.RootDirPath

	out, err := cmd.CombinedOutput()
	stop(err)
	if err != nil {
		return fmt.Errorf("failed to run sqlc generate %w: %s", err, out)
	}

	return nil
}

func printNextSteps(req CreateAuthRequest) error {
	printer.Success("added auth, finish the steps in internal/auth/TODO.md:")
	steps := []string{"wire config, di and router"}
	if len(envKeys(req)) > 0 {
		steps = append(steps, "add "+strings.Join(envKeys(req), ", ")+" to the staging and production compose files")
	}

	steps = append(steps, "goose up")
	if req.IsRedisStore() {
		steps = append(steps, "configure redis durability")
	} else {
		steps = append(steps, "schedule DeleteExpired on the auth stores")
	}

	steps = append(steps, "fill the TODOs in internal/auth")
	for i, step := range steps {
		printer.Info(fmt.Sprintf("%d. %s", i+1, step))
	}

	return nil
}

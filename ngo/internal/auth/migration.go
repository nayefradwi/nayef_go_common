package auth

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/nayefradwi/nayef_go_common/auth/authpg"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/common"
)

func migrationsDir(req CreateAuthRequest) string {
	return filepath.Join(req.RootDirPath, INTERNAL, INFRA, MIGRATIONS)
}

func authpgMigrationFiles(req CreateAuthRequest) []string {
	if req.IsRedisStore() {
		return nil
	}

	files := slices.Clone(authpgMigrations)
	if req.Mfa {
		files = append(files, AUTHPG_TOTP_MIGRATION)
	}

	return files
}

// users first, then authpg's files in order; Down undoes them in reverse.
func migrationSQL(req CreateAuthRequest) (string, error) {
	users, err := common.RenderToString(templatesFs, TMPL_USERS_TABLE, newAuthView(req))
	if err != nil {
		return "", err
	}

	ups := []string{strings.TrimSpace(users)}
	downs := []string{"DROP TABLE users;"}
	for _, name := range authpgMigrationFiles(req) {
		content, err := authpg.Migrations.ReadFile("migrations/" + name)
		if err != nil {
			return "", err
		}

		up, down, found := strings.Cut(string(content), "-- +goose Down")
		if !found {
			return "", fmt.Errorf("%s has no goose Down section", name)
		}

		ups = append(ups, strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(up), "-- +goose Up")))
		downs = append([]string{strings.TrimSpace(down)}, downs...)
	}

	return "-- +goose Up\n" + strings.Join(ups, "\n\n") + "\n\n-- +goose Down\n" + strings.Join(downs, "\n") + "\n", nil
}

// One file for every table: goose create stamps to the second, so two files made
// in the same second would share a version and goose refuses duplicates.
func createMigration(req CreateAuthRequest) error {
	sql, err := migrationSQL(req)
	if err != nil {
		return err
	}

	dir := migrationsDir(req)
	before, err := listFiles(dir)
	if err != nil {
		return err
	}

	cmd := exec.Command("goose", "-dir", dir, "create", AUTH, "sql")
	cmd.Dir = req.RootDirPath
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to run goose create %w: %s", err, out)
	}

	after, err := listFiles(dir)
	if err != nil {
		return err
	}

	created := []string{}
	for _, name := range after {
		if !slices.Contains(before, name) {
			created = append(created, name)
		}
	}

	if len(created) != 1 {
		return fmt.Errorf("expected goose to create one migration in %s, found %v", dir, created)
	}

	return os.WriteFile(filepath.Join(dir, created[0]), []byte(sql), 0644)
}

func listFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}

	return names, nil
}

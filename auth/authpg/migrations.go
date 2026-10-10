package authpg

import (
	"embed"
	"strings"
)

//go:embed migrations/*.sql
var Migrations embed.FS

func SessionMigration(table string) (string, error) {
	if err := validTable(table); err != nil {
		return "", err
	}
	b, err := Migrations.ReadFile("migrations/authpg_0003_sessions.sql")
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(string(b), defaultSessionTable, table), nil
}

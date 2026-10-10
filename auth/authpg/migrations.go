package authpg

import (
	"embed"
	"strings"
)

//go:embed migrations/*.sql
var Migrations embed.FS

func SessionMigration(table string) (string, error) {
	return renderMigration("migrations/authpg_0003_sessions.sql", defaultSessionTable, table)
}

func OtpMigration(table string) (string, error) {
	return renderMigration("migrations/authpg_0004_otps.sql", defaultOtpTable, table)
}

func renderMigration(file, defaultTable, table string) (string, error) {
	if err := validTable(table); err != nil {
		return "", err
	}
	b, err := Migrations.ReadFile(file)
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(string(b), defaultTable, table), nil
}

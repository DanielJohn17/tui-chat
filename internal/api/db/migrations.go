package db

import (
	"database/sql"
	"embed"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embendMigrations embed.FS

func Migrate(db *sql.DB) error {
	goose.SetBaseFS(embendMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	return goose.Up(db, "migrations")
}

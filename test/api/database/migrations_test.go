package database_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
)

// TEST_DATABASE_URL must point to a disposable PostgreSQL database. These tests
// apply the real migrations; query fixtures are rolled back or removed.
func TestMigrations(t *testing.T) {
	migrate(t)
}

func migrate(t *testing.T) string {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	db, err := sql.Open("pgx", url)
	require.NoError(t, err)
	defer db.Close()
	require.NoError(t, goose.SetDialect("postgres"))
	require.NoError(t, goose.UpContext(context.Background(), db, "../../../internal/api/db/migrations"))
	return url
}

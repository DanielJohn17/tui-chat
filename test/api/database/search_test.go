package database_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/DanielJohn17/tui-chat/internal/api/database"
	"github.com/DanielJohn17/tui-chat/internal/api/users"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestSearchSQL(t *testing.T) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, migrate(t))
	require.NoError(t, err)
	defer conn.Close(ctx)
	tx, err := conn.Begin(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)
	q := database.New(tx)
	add := func(username string) int64 {
		row, err := q.CreateUser(ctx, database.CreateUserParams{Name: username, Username: username, Password: "never selected"})
		require.NoError(t, err)
		return row.ID
	}
	self := add("zzdiscovery-self")
	add("zzdiscovery-b")
	add("zzdiscovery-a")
	add("zzdiscovery")
	add("ZZdiscovery")
	add("not-zzdiscovery")
	for i := 0; i < 25; i++ {
		add(fmt.Sprintf("zzdiscovery-%02d", i))
	}
	r := users.NewUserRepository(q)
	result, err := r.SearchUsers(ctx, self, "zzdiscovery")
	require.NoError(t, err)
	require.Len(t, result, 20)
	require.Equal(t, "ZZdiscovery", result[0].Username)
	require.Equal(t, "zzdiscovery", result[1].Username)
	for i, row := range result {
		require.NotEqual(t, self, row.ID)
		if i >= 2 {
			require.Equal(t, fmt.Sprintf("zzdiscovery-%02d", i-2), row.Username)
		}
	}
	for _, prefix := range []string{"zz%", "zz_", `zz\`} {
		id := add(prefix + "literal")
		rows, err := r.SearchUsers(ctx, self, prefix)
		require.NoError(t, err)
		require.Equal(t, []users.PublicUser{{ID: id, Name: prefix + "literal", Username: prefix + "literal"}}, rows)
	}
	rows, err := r.SearchUsers(ctx, self, "no-matches")
	require.NoError(t, err)
	require.NotNil(t, rows)
	require.Empty(t, rows)
	// If the query selected password, PostgreSQL would reject this restricted role.
	_, err = tx.Exec(ctx, `CREATE ROLE discovery_query_test; GRANT SELECT (id, name, username) ON users TO discovery_query_test; SET LOCAL ROLE discovery_query_test`)
	require.NoError(t, err)
	_, err = r.SearchUsers(ctx, self, "zzdiscovery")
	require.NoError(t, err)
}

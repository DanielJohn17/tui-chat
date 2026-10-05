package database_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/DanielJohn17/tui-chat/internal/api/database"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestDirectConversationSQL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, migrate(t))
	require.NoError(t, err)
	defer pool.Close()
	q := database.New(pool)
	ids := make([]int64, 3)
	prefix := fmt.Sprintf("discovery-conv-%d-", time.Now().UnixNano())
	for i := range ids {
		row, err := q.CreateUser(ctx, database.CreateUserParams{Name: "Participant", Username: fmt.Sprintf("%s%d", prefix, i), Password: "unused"})
		require.NoError(t, err)
		ids[i] = row.ID
	}
	defer func() {
		_, err := pool.Exec(context.Background(), "DELETE FROM users WHERE id = ANY($1)", ids)
		require.NoError(t, err)
		_, err = pool.Exec(context.Background(), "DELETE FROM users_archive WHERE id = ANY($1)", ids)
		require.NoError(t, err)
	}()
	params := database.GetOrCreateDirectConversationParams{UserID: ids[0], UserID2: ids[1]}
	first, err := q.GetOrCreateDirectConversation(ctx, params)
	require.NoError(t, err)
	require.Len(t, first, 2)
	require.Equal(t, ids[0], first[0].UserID)
	require.Equal(t, ids[1], first[1].UserID)
	again, err := q.GetOrCreateDirectConversation(ctx, database.GetOrCreateDirectConversationParams{UserID: ids[1], UserID2: ids[0]})
	require.NoError(t, err)
	require.Equal(t, first, again)
	var count int
	require.NoError(t, pool.QueryRow(ctx, "SELECT count(*) FROM participants WHERE conv_id = $1", first[0].ConvID).Scan(&count))
	require.Equal(t, 2, count)
	// Race on a fresh pair, not an already committed conversation.
	params.UserID2 = ids[2]
	type outcome struct {
		rows []database.GetOrCreateDirectConversationRow
		err  error
	}
	start := make(chan struct{})
	results := make(chan outcome, 12)
	for i := 0; i < cap(results); i++ {
		go func() {
			<-start
			rows, err := q.GetOrCreateDirectConversation(ctx, params)
			results <- outcome{rows, err}
		}()
	}
	close(start)
	var convID int64
	for i := 0; i < cap(results); i++ {
		out := <-results
		require.NoError(t, out.err)
		require.Len(t, out.rows, 2)
		if convID == 0 {
			convID = out.rows[0].ConvID
		}
		for _, row := range out.rows {
			require.Equal(t, convID, row.ConvID)
			require.True(t, strings.HasPrefix(row.Username, prefix))
		}
	}
	require.NoError(t, pool.QueryRow(ctx, "SELECT count(*) FROM participants WHERE conv_id = $1", convID).Scan(&count))
	require.Equal(t, 2, count)
	require.NoError(t, pool.QueryRow(ctx, "SELECT count(*) FROM conversations WHERE user_min_id = $1 AND user_max_id = $2 AND NOT is_deleting", ids[0], ids[2]).Scan(&count))
	require.Equal(t, 1, count)
}

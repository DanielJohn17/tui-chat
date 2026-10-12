package test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DanielJohn17/tui-chat/internal/tui"
	"github.com/DanielJohn17/tui-chat/internal/tui/client"
	"github.com/DanielJohn17/tui-chat/internal/tui/views/modals"
	"github.com/DanielJohn17/tui-chat/internal/tui/views/profile"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type discoveryClient struct {
	*TestClient
	search func(context.Context, string) ([]client.UserSummary, error)
	open   func(context.Context, int64) (client.Chat, error)
}

func (c *discoveryClient) SearchUsers(ctx context.Context, query string) ([]client.UserSummary, error) {
	return c.search(ctx, query)
}

func (c *discoveryClient) OpenDirectConversation(ctx context.Context, id int64) (client.Chat, error) {
	return c.open(ctx, id)
}

// Text input updates batch the debounce with cursor commands. Execute only the
// debounce here, never the unrelated long-lived application commands.
func discoveryDebounce(t *testing.T, cmd tea.Cmd) tui.DiscoveryDebounceMsg {
	t.Helper()
	require.NotNil(t, cmd)
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, child := range batch {
			if tick, ok := child().(tui.DiscoveryDebounceMsg); ok {
				return tick
			}
		}
	}
	tick, ok := msg.(tui.DiscoveryDebounceMsg)
	require.True(t, ok, "expected discovery debounce, got %T", msg)
	return tick
}

func discoveryApp(c client.Client) tea.Model {
	p := c.Profile()
	p.Token = "discovery-test-token"
	c.SetProfile(p)
	m := tea.Model(tui.NewApp(c))
	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 25})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	return m
}

func discoveryResults(t *testing.T, m tea.Model, query string) (tea.Model, tui.DiscoveryDebounceMsg, tea.Msg) {
	t.Helper()
	m, debounce := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(query)})
	tick := discoveryDebounce(t, debounce)
	m, search := m.Update(tick)
	require.NotNil(t, search)
	result := search()
	m, _ = m.Update(result)
	return m, tick, result
}

func TestDiscoveryAsyncSelectionAndUpsert(t *testing.T) {
	c := &discoveryClient{TestClient: newTestClient()}
	original := c.Chats()[2]
	c.search = func(ctx context.Context, query string) ([]client.UserSummary, error) {
		assert.Equal(t, "ch", query)
		return []client.UserSummary{{ID: 102, Name: "Bob Martin", Username: "bob"}, {ID: 104, Name: "Charlie Zhang", Username: "charlie"}}, nil
	}
	opened := 0
	c.open = func(ctx context.Context, id int64) (client.Chat, error) {
		opened++
		assert.Equal(t, int64(104), id)
		return client.Chat{ID: 3, RecipientID: 104, Name: "Charlie Updated", Username: "charlie"}, nil
	}
	m := discoveryApp(c)
	m, _, _ = discoveryResults(t, m, "ch")
	assert.Contains(t, m.View(), "Bob Martin @bob")
	assert.Contains(t, m.View(), "Charlie Zhang @charlie")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Zero(t, opened, "Update must not execute network calls")
	assert.Contains(t, m.View(), "Opening conversation")
	m, duplicate := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Nil(t, duplicate)
	m, _ = m.Update(cmd())
	assert.Equal(t, 1, opened)
	assert.NotContains(t, m.View(), "NEW DIRECT MESSAGE")
	assert.NotContains(t, m.View(), "INPUT MODE")
	assert.Contains(t, m.View(), "Charlie Updated")
	ch := c.Chats()[2]
	assert.Equal(t, original.LastMessage, ch.LastMessage)
	assert.Equal(t, original.Unread, ch.Unread)
	assert.Equal(t, original.Time, ch.Time)
	assert.Equal(t, original.Online, ch.Online)
	assert.Len(t, c.Chats(), 12)
}

func TestDiscoveryStaleResultsAndCancellation(t *testing.T) {
	c := &discoveryClient{TestClient: newTestClient()}
	c.search = func(ctx context.Context, query string) ([]client.UserSummary, error) {
		return []client.UserSummary{{ID: 888, Name: "Stale Name", Username: "stale"}}, nil
	}
	c.open = func(ctx context.Context, id int64) (client.Chat, error) {
		assert.ErrorIs(t, ctx.Err(), context.Canceled)
		return client.Chat{ID: 999, Name: "Stale Chat"}, nil
	}
	m := discoveryApp(c)
	m, tick, stale := discoveryResults(t, m, "st")
	m, pending := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.ErrorIs(t, tick.Context.Err(), context.Canceled)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m, _ = m.Update(stale)
	m, _ = m.Update(pending())
	assert.Contains(t, m.View(), "NEW DIRECT MESSAGE")
	assert.NotContains(t, m.View(), "Stale Name")
	assert.Len(t, c.Chats(), 12)
	// A new query invalidates both a running search and its debounce message.
	m, tick, stale = discoveryResults(t, m, "st")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	assert.ErrorIs(t, tick.Context.Err(), context.Canceled)
	m, cmd := m.Update(tick)
	assert.Nil(t, cmd)
	m, _ = m.Update(stale)
	assert.NotContains(t, m.View(), "Stale Name")
	// Logout invalidates results even if the transport ignores cancellation.
	m, tick, stale = discoveryResults(t, m, "t")
	m, _ = m.Update(profile.LogoutMsg{})
	assert.ErrorIs(t, tick.Context.Err(), context.Canceled)
	m, _ = m.Update(stale)
	assert.NotContains(t, m.View(), "NEW DIRECT MESSAGE")
}

func TestDiscoveryErrorsEmptyAndRetry(t *testing.T) {
	c := &discoveryClient{TestClient: newTestClient()}
	c.search = func(context.Context, string) ([]client.UserSummary, error) { return nil, errors.New("search failed") }
	c.open = func(context.Context, int64) (client.Chat, error) { return client.Chat{}, errors.New("open failed") }
	m := discoveryApp(c)
	m, _, _ = discoveryResults(t, m, "bo")
	assert.Contains(t, m.View(), "search failed")
	c.search = func(context.Context, string) ([]client.UserSummary, error) { return nil, nil }
	m, _, _ = discoveryResults(t, m, "b")
	assert.Contains(t, m.View(), "No results")
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Nil(t, cmd)
	c.search = func(context.Context, string) ([]client.UserSummary, error) {
		return []client.UserSummary{{ID: 102, Name: "Bob", Username: "bob"}}, nil
	}
	m, _, _ = discoveryResults(t, m, "y")
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m, _ = m.Update(cmd())
	assert.Contains(t, m.View(), "open failed")
	assert.Contains(t, m.View(), "NEW DIRECT MESSAGE")
	c.open = func(context.Context, int64) (client.Chat, error) {
		return client.Chat{ID: 999, RecipientID: 102, Name: "Bob", Username: "bob"}, nil
	}
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m, followup := m.Update(cmd())
	assert.NotNil(t, followup, "new conversation should fetch history and focus asynchronously")
	assert.Equal(t, int64(999), c.Chats()[0].ID)
	assert.Contains(t, m.View(), "Loading conversation messages")
}

func TestDiscoveryDebounceMinimumAndLimit(t *testing.T) {
	c := &discoveryClient{TestClient: newTestClient()}
	c.search = func(context.Context, string) ([]client.UserSummary, error) {
		t.Fatal("search ran inside Update")
		return nil, nil
	}
	m := discoveryApp(c)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	assert.Contains(t, m.View(), "Type at least 2 characters")
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(strings.Repeat("b", 25))})
	assert.Contains(t, m.View(), "Searching...")
	start := time.Now()
	tick := discoveryDebounce(t, cmd)
	assert.GreaterOrEqual(t, time.Since(start), 250*time.Millisecond)
	assert.Len(t, []rune(tick.Query), 20)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.ErrorIs(t, tick.Context.Err(), context.Canceled)
}

func TestDiscoveryResponsiveLayout(t *testing.T) {
	m := modals.NewDM()
	m.Reset()
	m.SetResults([]client.UserSummary{{ID: 1, Name: "Very long name 日本語 👍 with extra text", Username: "long_username"}}, nil)
	for _, size := range [][2]int{{120, 35}, {48, 14}, {30, 10}, {12, 7}, {5, 4}, {1, 1}} {
		view := m.View(size[0], size[1])
		assert.LessOrEqual(t, lipgloss.Width(view), size[0], "width at %v", size)
		assert.Equal(t, size[1], lipgloss.Height(view), "height at %v", size)
		assert.NotContains(t, view, "\uFFFD")
	}
}

func TestDiscoveryLogoutCancelsOpen(t *testing.T) {
	c := &discoveryClient{TestClient: newTestClient()}
	c.search = func(context.Context, string) ([]client.UserSummary, error) {
		return []client.UserSummary{{ID: 888, Name: "Pending User", Username: "pending"}}, nil
	}
	c.open = func(ctx context.Context, id int64) (client.Chat, error) {
		assert.ErrorIs(t, ctx.Err(), context.Canceled)
		return client.Chat{ID: 999, RecipientID: id, Name: "Pending User", Username: "pending"}, nil
	}
	m := discoveryApp(c)
	m, _, _ = discoveryResults(t, m, "pe")
	m, open := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m, _ = m.Update(profile.LogoutMsg{})
	m, cmd := m.Update(open())
	assert.Nil(t, cmd)
	assert.Len(t, c.Chats(), 12)
	assert.NotContains(t, m.View(), "NEW DIRECT MESSAGE")
}

func TestDiscoveryCancelledDebounceDoesNotSearch(t *testing.T) {
	c := &discoveryClient{TestClient: newTestClient()}
	c.search = func(context.Context, string) ([]client.UserSummary, error) {
		t.Fatal("cancelled debounce must not search")
		return nil, nil
	}
	m := discoveryApp(c)
	m, debounce := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("bo")})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	msg := debounce()
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, child := range batch {
			result := child()
			_, isDebounce := result.(tui.DiscoveryDebounceMsg)
			assert.False(t, isDebounce)
		}
	} else {
		assert.Nil(t, msg)
	}
	assert.NotContains(t, m.View(), "NEW DIRECT MESSAGE")
}

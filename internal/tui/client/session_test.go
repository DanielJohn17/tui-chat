package client

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveSessionPath(t *testing.T) {
	// Test default production path
	prodPath := ResolveSessionPath("production", "")
	assert.Contains(t, prodPath, "session.json")
	assert.NotContains(t, prodPath, "dev_session")

	// Test production with named session
	prodNamedPath := ResolveSessionPath("production", "user1")
	assert.Contains(t, prodNamedPath, "session_user1.json")

	// Test development default path
	devPath := ResolveSessionPath("development", "")
	assert.Contains(t, devPath, "dev_session.json")

	// Test development with named session
	devAlicePath := ResolveSessionPath("development", "alice")
	assert.Contains(t, devAlicePath, "dev_session_alice.json")

	// Test environment variable override
	t.Setenv("TUI_CHAT_SESSION_FILE", "/tmp/custom_session.json")
	customPath := ResolveSessionPath("production", "")
	assert.Equal(t, "/tmp/custom_session.json", customPath)
}

func TestSessionSaveLoadClear(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tui-chat-session-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	sessionPath := filepath.Join(tempDir, "test_session.json")

	// 1. Loading non-existent session returns error
	prof, err := LoadSession(sessionPath)
	assert.Error(t, err)
	assert.Nil(t, prof)

	// 2. Saving session with empty token returns error
	err = SaveSession(sessionPath, Profile{ID: 1, Name: "Test"})
	assert.Error(t, err)

	// 3. Saving valid session
	origProf := Profile{
		ID:        42,
		Name:      "Alice Walker",
		Username:  "alice",
		Token:     "jwt-token-12345",
		CreatedAt: "2026-09-21T12:00:00Z",
	}
	err = SaveSession(sessionPath, origProf)
	require.NoError(t, err)

	// Verify file permissions (0600)
	info, err := os.Stat(sessionPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

	// 4. Loading saved session
	loadedProf, err := LoadSession(sessionPath)
	require.NoError(t, err)
	require.NotNil(t, loadedProf)
	assert.Equal(t, origProf.ID, loadedProf.ID)
	assert.Equal(t, origProf.Name, loadedProf.Name)
	assert.Equal(t, origProf.Username, loadedProf.Username)
	assert.Equal(t, origProf.Token, loadedProf.Token)

	// 5. Updating / overwriting session
	origProf.Name = "Alice In Wonderland"
	err = SaveSession(sessionPath, origProf)
	require.NoError(t, err)

	reloadedProf, err := LoadSession(sessionPath)
	require.NoError(t, err)
	assert.Equal(t, "Alice In Wonderland", reloadedProf.Name)

	// 6. Clearing session
	err = ClearSession(sessionPath)
	require.NoError(t, err)

	// Verify file is gone
	_, err = os.Stat(sessionPath)
	assert.True(t, os.IsNotExist(err))

	// Clearing already removed file should not error
	err = ClearSession(sessionPath)
	assert.NoError(t, err)
}

func TestCorruptedSession(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tui-chat-session-corrupt-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	sessionPath := filepath.Join(tempDir, "corrupt.json")
	err = os.WriteFile(sessionPath, []byte("{invalid json"), 0600)
	require.NoError(t, err)

	prof, err := LoadSession(sessionPath)
	assert.Error(t, err)
	assert.Nil(t, prof)
}

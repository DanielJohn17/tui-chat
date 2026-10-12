package crypto

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateIdentityKey(t *testing.T) {
	key, err := GenerateIdentityKey()
	require.NoError(t, err)
	require.NotNil(t, key)

	// Private and public keys must not be zero-valued
	var zero [32]byte
	assert.NotEqual(t, zero, key.PrivateKey)
	assert.NotEqual(t, zero, key.PublicKey)

	// Two generated keys must differ
	key2, err := GenerateIdentityKey()
	require.NoError(t, err)
	assert.NotEqual(t, key.PrivateKey, key2.PrivateKey)
	assert.NotEqual(t, key.PublicKey, key2.PublicKey)
}

func TestSaveAndLoadIdentityKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "identity.key")

	key, err := GenerateIdentityKey()
	require.NoError(t, err)

	err = SaveIdentityKey(path, key)
	require.NoError(t, err)

	// File should exist with restricted permissions
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

	// Load and compare
	loaded, err := LoadOrCreateIdentityKey(path)
	require.NoError(t, err)
	assert.Equal(t, key.PrivateKey, loaded.PrivateKey)
	assert.Equal(t, key.PublicKey, loaded.PublicKey)
}

func TestLoadOrCreateIdentityKey_CreatesWhenMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "new_identity.key")

	// File doesn't exist yet — should be created
	key, err := LoadOrCreateIdentityKey(path)
	require.NoError(t, err)
	require.NotNil(t, key)

	// File must now exist
	_, err = os.Stat(path)
	assert.NoError(t, err)

	// Loading again must return the same key
	key2, err := LoadOrCreateIdentityKey(path)
	require.NoError(t, err)
	assert.Equal(t, key.PrivateKey, key2.PrivateKey)
	assert.Equal(t, key.PublicKey, key2.PublicKey)
}

func TestSaveIdentityKey_EmptyPath(t *testing.T) {
	key, _ := GenerateIdentityKey()
	err := SaveIdentityKey("", key)
	assert.Error(t, err)
}

func TestSaveIdentityKey_NilKey(t *testing.T) {
	dir := t.TempDir()
	err := SaveIdentityKey(filepath.Join(dir, "key"), nil)
	assert.Error(t, err)
}

func TestResolveIdentityKeyPath(t *testing.T) {
	t.Run("production path", func(t *testing.T) {
		path := ResolveIdentityKeyPath("production", "")
		assert.Contains(t, path, "tui-chat")
		assert.Contains(t, path, "identity.key")
	})

	t.Run("development path", func(t *testing.T) {
		path := ResolveIdentityKeyPath("development", "")
		assert.Contains(t, path, "dev_identity.key")
	})

	t.Run("development with session name", func(t *testing.T) {
		path := ResolveIdentityKeyPath("dev", "alice")
		assert.Contains(t, path, "dev_identity_alice.key")
	})

	t.Run("production with user ID", func(t *testing.T) {
		path := ResolveIdentityKeyPath("production", "", 42)
		assert.Contains(t, path, "identity_42.key")
	})

	t.Run("dev with session name and user ID", func(t *testing.T) {
		path := ResolveIdentityKeyPath("dev", "alice", 42)
		assert.Contains(t, path, "dev_identity_alice_42.key")
	})

	t.Run("env var override", func(t *testing.T) {
		t.Setenv("TUI_CHAT_IDENTITY_KEY_FILE", "/tmp/custom.key")
		path := ResolveIdentityKeyPath("production", "")
		assert.Equal(t, "/tmp/custom.key", path)
	})
}

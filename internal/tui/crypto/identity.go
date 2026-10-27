// Package crypto provides end-to-end encryption primitives for the TUI client.
// It implements X25519-ECDH key exchange and AES-256-GCM authenticated encryption
// so that the server stores only opaque ciphertext and never sees plaintext messages.
package crypto

import (
	"crypto/ecdh"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// IdentityKey holds an X25519 keypair for the local user.
// The private key must never leave the device.
type IdentityKey struct {
	PrivateKey [32]byte
	PublicKey  [32]byte
}

// GenerateIdentityKey creates a fresh X25519 keypair using a secure random source.
func GenerateIdentityKey() (*IdentityKey, error) {
	priv, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("crypto/identity: failed to generate keypair: %w", err)
	}

	var privArr, pubArr [32]byte
	copy(privArr[:], priv.Bytes())
	copy(pubArr[:], priv.PublicKey().Bytes())

	return &IdentityKey{PrivateKey: privArr, PublicKey: pubArr}, nil
}

// LoadOrCreateIdentityKey loads an existing identity key from path, or generates
// and saves a new one if the file does not exist.
// The key file is stored in raw binary (64 bytes: 32 priv || 32 pub) with mode 0600.
func LoadOrCreateIdentityKey(path string) (*IdentityKey, error) {
	if path == "" {
		return nil, errors.New("crypto/identity: empty key path")
	}

	data, err := os.ReadFile(path)
	if err == nil {
		return parseIdentityKey(data)
	}
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("crypto/identity: failed to read key file: %w", err)
	}

	// Check if a base identity_<uid>.key exists in the same directory to avoid regenerating
	// and invalidating existing message history when launching under a session flag
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	nameWithoutExt := strings.TrimSuffix(base, ext)
	parts := strings.Split(nameWithoutExt, "_")
	if len(parts) >= 3 {
		fallbackName := fmt.Sprintf("%s_%s%s", parts[0], parts[len(parts)-1], ext)
		fallbackPath := filepath.Join(dir, fallbackName)
		if fallbackData, fErr := os.ReadFile(fallbackPath); fErr == nil {
			if k, pErr := parseIdentityKey(fallbackData); pErr == nil {
				_ = SaveIdentityKey(path, k)
				return k, nil
			}
		}
	}

	// Generate a new key and persist it
	key, err := GenerateIdentityKey()
	if err != nil {
		return nil, err
	}
	if err := SaveIdentityKey(path, key); err != nil {
		return nil, err
	}
	return key, nil
}

// SaveIdentityKey persists key to path atomically with mode 0600.
// The parent directory is created if it does not exist (mode 0700).
func SaveIdentityKey(path string, key *IdentityKey) error {
	if path == "" {
		return errors.New("crypto/identity: empty key path")
	}
	if key == nil {
		return errors.New("crypto/identity: nil identity key")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("crypto/identity: failed to create key directory: %w", err)
	}

	// 64 bytes: private (32) || public (32)
	data := make([]byte, 64)
	copy(data[:32], key.PrivateKey[:])
	copy(data[32:], key.PublicKey[:])

	tmpPath := fmt.Sprintf("%s.tmp.%d", path, os.Getpid())
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return fmt.Errorf("crypto/identity: failed to write temp key file: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("crypto/identity: failed to persist key file: %w", err)
	}
	return nil
}

func parseIdentityKey(data []byte) (*IdentityKey, error) {
	if len(data) != 64 {
		return nil, fmt.Errorf("crypto/identity: corrupted key file: expected 64 bytes, got %d", len(data))
	}
	var key IdentityKey
	copy(key.PrivateKey[:], data[:32])
	copy(key.PublicKey[:], data[32:])
	return &key, nil
}

// ResolveIdentityKeyPath returns the filesystem path to the identity key file,
// mirroring the session path resolution logic in internal/tui/client/session.go.
//
//   - Production: ~/.config/tui-chat/identity.key (or identity_<userID>.key)
//   - Development: ~/.config/tui-chat/dev_identity.key (or dev_identity_<name>_<userID>.key)
func ResolveIdentityKeyPath(goEnv, sessionName string, optionalUserID ...int64) string {
	if customFile := os.Getenv("TUI_CHAT_IDENTITY_KEY_FILE"); customFile != "" {
		return customFile
	}

	baseDir, err := os.UserConfigDir()
	if err != nil || baseDir == "" {
		homeDir, hErr := os.UserHomeDir()
		if hErr == nil && homeDir != "" {
			baseDir = filepath.Join(homeDir, ".config")
		} else {
			baseDir = "."
		}
	}

	appDir := filepath.Join(baseDir, "tui-chat")

	var uid int64
	if len(optionalUserID) > 0 {
		uid = optionalUserID[0]
	}

	sessionName = strings.TrimSpace(sessionName)
	cleanEnv := strings.ToLower(strings.TrimSpace(goEnv))
	isDev := cleanEnv == "development" || cleanEnv == "dev" || cleanEnv == "local"

	prefix := "identity"
	if isDev {
		prefix = "dev_identity"
	}

	parts := []string{prefix}
	if sessionName != "" {
		parts = append(parts, sanitize(sessionName))
	}
	if uid > 0 {
		parts = append(parts, fmt.Sprintf("%d", uid))
	}

	fileName := strings.Join(parts, "_") + ".key"
	return filepath.Join(appDir, fileName)
}

func sanitize(name string) string {
	clean := strings.ToLower(name)
	var sb strings.Builder
	for _, r := range clean {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			sb.WriteRune(r)
		}
	}
	res := sb.String()
	if res == "" {
		return "default"
	}
	return res
}

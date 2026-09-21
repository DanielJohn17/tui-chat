package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResolveSessionPath determines the session file location based on the environment and session identifier.
// In production mode, it defaults to the OS user config directory:
//   - Linux/macOS: ~/.config/tui-chat/session.json
//   - Windows: %APPDATA%\tui-chat\session.json
//
// In development mode (or if a specific session name is given), it scopes the file name:
//   - e.g. dev_session.json or dev_session_alice.json
func ResolveSessionPath(goEnv, sessionName string) string {
	// If explicit env var TUI_CHAT_SESSION_FILE is set, use it directly (helpful for tests)
	if customFile := os.Getenv("TUI_CHAT_SESSION_FILE"); customFile != "" {
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

	sessionName = strings.TrimSpace(sessionName)
	cleanEnv := strings.ToLower(strings.TrimSpace(goEnv))
	isDev := cleanEnv == "development" || cleanEnv == "dev" || cleanEnv == "local"

	var fileName string
	if isDev {
		if sessionName != "" {
			fileName = fmt.Sprintf("dev_session_%s.json", sanitizeSessionName(sessionName))
		} else {
			fileName = "dev_session.json"
		}
	} else {
		if sessionName != "" {
			fileName = fmt.Sprintf("session_%s.json", sanitizeSessionName(sessionName))
		} else {
			fileName = "session.json"
		}
	}

	return filepath.Join(appDir, fileName)
}

func sanitizeSessionName(name string) string {
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

// LoadSession reads and unmarshals the persisted session from the given file path.
// If the file does not exist, it returns (nil, os.ErrNotExist).
func LoadSession(sessionPath string) (*Profile, error) {
	if sessionPath == "" {
		return nil, errors.New("empty session path")
	}

	data, err := os.ReadFile(sessionPath)
	if err != nil {
		return nil, err
	}

	var prof Profile
	if err := json.Unmarshal(data, &prof); err != nil {
		return nil, fmt.Errorf("corrupted session data: %w", err)
	}

	if prof.Token == "" {
		return nil, errors.New("empty session token")
	}

	return &prof, nil
}

// SaveSession securely writes the profile session to the given file path.
// It creates the parent directory if needed and performs an atomic write with 0600 permissions.
func SaveSession(sessionPath string, prof Profile) error {
	if sessionPath == "" {
		return errors.New("empty session path")
	}

	if prof.Token == "" {
		return errors.New("cannot save session with empty token")
	}

	dir := filepath.Dir(sessionPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create session directory: %w", err)
	}

	data, err := json.MarshalIndent(prof, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode session: %w", err)
	}

	// Write atomically to temporary file first
	tmpFile := fmt.Sprintf("%s.tmp.%d", sessionPath, os.Getpid())
	if err := os.WriteFile(tmpFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write temporary session file: %w", err)
	}

	// Rename temp file to target path
	if err := os.Rename(tmpFile, sessionPath); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to persist session file: %w", err)
	}

	return nil
}

// ClearSession removes the session file if it exists.
func ClearSession(sessionPath string) error {
	if sessionPath == "" {
		return nil
	}
	if err := os.Remove(sessionPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

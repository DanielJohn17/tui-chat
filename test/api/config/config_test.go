package config_test

import (
	"strings"
	"testing"

	"github.com/DanielJohn17/tui-chat/internal/api/config"
)

func TestValidateJWTSecret(t *testing.T) {
	tests := []struct {
		name      string
		secret    string
		goEnv     string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "production with empty secret",
			secret:    "",
			goEnv:     "production",
			expectErr: true,
			errMsg:    "required in production",
		},
		{
			name:      "production with public default fallback",
			secret:    config.DefaultInsecureSecretKey,
			goEnv:     "production",
			expectErr: true,
			errMsg:    "cannot use default fallback",
		},
		{
			name:      "production with short secret",
			secret:    "too-short-secret",
			goEnv:     "production",
			expectErr: true,
			errMsg:    "must be at least 32 characters long",
		},
		{
			name:      "production with strong secret (>= 32 chars)",
			secret:    "super-strong-jwt-secret-key-that-is-at-least-32-chars-long",
			goEnv:     "production",
			expectErr: false,
		},
		{
			name:      "development with empty secret allowed",
			secret:    "",
			goEnv:     "development",
			expectErr: false,
		},
		{
			name:      "development with default secret allowed",
			secret:    config.DefaultInsecureSecretKey,
			goEnv:     "development",
			expectErr: false,
		},
		{
			name:      "development with short secret allowed",
			secret:    "dev-secret",
			goEnv:     "development",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.ValidateJWTSecret(tt.secret, tt.goEnv)
			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Fatalf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
			}
		})
	}
}

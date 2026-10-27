package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DanielJohn17/tui-chat/internal/api/config"
	"github.com/DanielJohn17/tui-chat/internal/api/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestServer(expectedSecret string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ClientSecretAuth(expectedSecret))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/api/v1/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": "secret_data"})
	})

	return r
}

func TestClientSecretAuth(t *testing.T) {
	const secret = "test-secret-min-32-characters-for-testing"
	r := setupTestServer(secret)

	tests := []struct {
		name         string
		path         string
		headerKey    string
		headerVal    string
		expectedCode int
		expectedErr  string
	}{
		{
			name:         "missing header returns 403 Forbidden",
			path:         "/api/v1/protected",
			expectedCode: http.StatusForbidden,
			expectedErr:  "Forbidden: missing client secret header",
		},
		{
			name:         "invalid secret returns 403 Forbidden",
			path:         "/api/v1/protected",
			headerKey:    config.ClientSecretHeader,
			headerVal:    "wrong-client-secret-value",
			expectedCode: http.StatusForbidden,
			expectedErr:  "Forbidden: invalid client secret",
		},
		{
			name:         "valid secret with primary header returns 200 OK",
			path:         "/api/v1/protected",
			headerKey:    config.ClientSecretHeader,
			headerVal:    secret,
			expectedCode: http.StatusOK,
		},
		{
			name:         "valid secret with alternative header returns 200 OK",
			path:         "/api/v1/protected",
			headerKey:    config.AltClientSecretHeader,
			headerVal:    secret,
			expectedCode: http.StatusOK,
		},
		{
			name:         "health probe bypasses auth",
			path:         "/health",
			expectedCode: http.StatusOK,
		},
		{
			name:         "healthz probe bypasses auth",
			path:         "/healthz",
			expectedCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, tt.path, nil)
			require.NoError(t, err)

			if tt.headerKey != "" {
				req.Header.Set(tt.headerKey, tt.headerVal)
			}
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedErr != "" {
				var resp map[string]any
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, false, resp["success"])
				assert.Equal(t, tt.expectedErr, resp["error"])
			}
		})
	}

	t.Run("panic on empty expected secret", func(t *testing.T) {
		assert.Panics(t, func() {
			middleware.ClientSecretAuth("")
		})
	})
}

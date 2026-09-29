package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DanielJohn17/tui-chat/internal/api/config"
	"github.com/DanielJohn17/tui-chat/internal/api/middleware"
	"github.com/gin-gonic/gin"
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

func TestClientSecretAuth_MissingHeader(t *testing.T) {
	const secret = "test-secret-min-32-characters-for-testing"
	r := setupTestServer(secret)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/protected", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403 Forbidden, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["success"] != false || resp["error"] != "Forbidden: missing client secret header" {
		t.Fatalf("unexpected error payload: %v", resp)
	}
}

func TestClientSecretAuth_InvalidSecret(t *testing.T) {
	const secret = "test-secret-min-32-characters-for-testing"
	r := setupTestServer(secret)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/protected", nil)
	req.Header.Set(config.ClientSecretHeader, "wrong-client-secret-value")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403 Forbidden, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["success"] != false || resp["error"] != "Forbidden: invalid client secret" {
		t.Fatalf("unexpected error payload: %v", resp)
	}
}

func TestClientSecretAuth_ValidSecret_PrimaryHeader(t *testing.T) {
	const secret = "test-secret-min-32-characters-for-testing"
	r := setupTestServer(secret)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/protected", nil)
	req.Header.Set(config.ClientSecretHeader, secret)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d (body: %s)", w.Code, w.Body.String())
	}
}

func TestClientSecretAuth_ValidSecret_AltHeader(t *testing.T) {
	const secret = "test-secret-min-32-characters-for-testing"
	r := setupTestServer(secret)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/protected", nil)
	req.Header.Set(config.AltClientSecretHeader, secret)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d (body: %s)", w.Code, w.Body.String())
	}
}

func TestClientSecretAuth_HealthProbesBypass(t *testing.T) {
	const secret = "test-secret-min-32-characters-for-testing"
	r := setupTestServer(secret)

	for _, probePath := range []string{"/health", "/healthz"} {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, probePath, nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK for %s without secret, got %d", probePath, w.Code)
		}
	}
}

func TestClientSecretAuth_PanicOnEmptyExpectedSecret(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic on empty expectedSecret, got none")
		}
	}()

	_ = middleware.ClientSecretAuth("")
}

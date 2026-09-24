package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DanielJohn17/tui-chat/internal/api/helpers"
	"github.com/DanielJohn17/tui-chat/internal/api/middleware"
	"github.com/DanielJohn17/tui-chat/internal/api/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCursor(t *testing.T) {
	gin.SetMode(gin.TestMode)

	base64Cursor := helpers.EncodeCursor("2026-09-07T18:49:56+03:00", 3)

	tests := []struct {
		name     string
		cursor   string
		wantTime string
		wantID   int64
		wantErr  bool
	}{
		{
			name:     "valid Base64 URL-safe cursor",
			cursor:   base64Cursor,
			wantTime: "2026-09-07T18:49:56+03:00",
			wantID:   3,
			wantErr:  false,
		},
		{
			name:     "valid UTC RFC3339",
			cursor:   "2026-09-08T12:00:00Z+15",
			wantTime: "2026-09-08T12:00:00Z",
			wantID:   15,
			wantErr:  false,
		},
		{
			name:     "valid timezone offset +03:00",
			cursor:   "2026-09-08T15:00:00+03:00+42",
			wantTime: "2026-09-08T15:00:00+03:00",
			wantID:   42,
			wantErr:  false,
		},
		{
			name:     "valid timezone offset -05:00",
			cursor:   "2026-09-08T07:00:00-05:00+100",
			wantTime: "2026-09-08T07:00:00-05:00",
			wantID:   100,
			wantErr:  false,
		},
		{
			name:     "valid with space as plus",
			cursor:   "2026-09-08T12:00:00Z 15",
			wantTime: "2026-09-08T12:00:00Z",
			wantID:   15,
			wantErr:  false,
		},
		{
			name:     "valid timezone offset decoded with spaces in query string",
			cursor:   "2026-09-07T18:49:56 03:00 3",
			wantTime: "2026-09-07T18:49:56+03:00",
			wantID:   3,
			wantErr:  false,
		},
		{
			name:     "valid timezone offset with literal plus",
			cursor:   "2026-09-07T18:49:56+03:00+3",
			wantTime: "2026-09-07T18:49:56+03:00",
			wantID:   3,
			wantErr:  false,
		},
		{
			name:    "invalid format without separator",
			cursor:  "2026-09-08T12:00:00Z",
			wantErr: true,
		},
		{
			name:    "invalid time format",
			cursor:  "invalid-time+10",
			wantErr: true,
		},
		{
			name:    "invalid id format",
			cursor:  "2026-09-08T12:00:00Z+abc",
			wantErr: true,
		},
		{
			name:    "negative id",
			cursor:  "2026-09-08T12:00:00Z+-5",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTime, gotID, err := middleware.ParseCursor(tt.cursor)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				parsedWantTime, _ := time.Parse(time.RFC3339, tt.wantTime)
				assert.True(t, gotTime.Equal(parsedWantTime))
				assert.Equal(t, tt.wantID, gotID)
			}
		})
	}
}

func TestValidateQueryParamsMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	var extractedQuery types.URLQueryParams

	r.GET("/test", middleware.ValidateQueryParams(), func(c *gin.Context) {
		if q, exists := c.Get("queries"); exists {
			extractedQuery = q.(types.URLQueryParams)
		}
		c.Status(http.StatusOK)
	})

	// Test 1: With cursor
	req := httptest.NewRequest(http.MethodGet, "/test?cursor=2026-09-08T12:00:00Z%2B25&limit=50", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	expectedTime, _ := time.Parse(time.RFC3339, "2026-09-08T12:00:00Z")
	assert.True(t, extractedQuery.CursorTime.Equal(expectedTime))
	assert.Equal(t, int64(25), extractedQuery.CursorID)
	assert.Equal(t, int64(50), extractedQuery.Limit)

	// Test 2: Invalid cursor returns 400
	reqBad := httptest.NewRequest(http.MethodGet, "/test?cursor=badcursor", nil)
	wBad := httptest.NewRecorder()
	r.ServeHTTP(wBad, reqBad)

	assert.Equal(t, http.StatusBadRequest, wBad.Code)

	// Test 3: Raw query with plus signs in timezone offset and ID separator
	reqRawPlus := httptest.NewRequest(http.MethodGet, "/test?limit=2&cursor=2026-09-07T18:49:56+03:00+3", nil)
	wRawPlus := httptest.NewRecorder()
	r.ServeHTTP(wRawPlus, reqRawPlus)

	assert.Equal(t, http.StatusOK, wRawPlus.Code)
	expectedTimeRaw, _ := time.Parse(time.RFC3339, "2026-09-07T18:49:56+03:00")
	assert.True(t, extractedQuery.CursorTime.Equal(expectedTimeRaw))
	assert.Equal(t, int64(3), extractedQuery.CursorID)

	// Test 4: Base64 URL-safe cursor query parameter
	b64Cursor := helpers.EncodeCursor("2026-09-07T18:49:56+03:00", 3)
	reqB64 := httptest.NewRequest(http.MethodGet, "/test?limit=10&cursor="+b64Cursor, nil)
	wB64 := httptest.NewRecorder()
	r.ServeHTTP(wB64, reqB64)

	assert.Equal(t, http.StatusOK, wB64.Code)
	assert.True(t, extractedQuery.CursorTime.Equal(expectedTimeRaw))
	assert.Equal(t, int64(3), extractedQuery.CursorID)
	assert.Equal(t, int64(10), extractedQuery.Limit)
}

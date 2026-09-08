package test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DanielJohn17/tui-chat/app/internal/api/helpers"
	"github.com/DanielJohn17/tui-chat/app/internal/api/middleware"
	"github.com/DanielJohn17/tui-chat/app/internal/api/types"
	"github.com/gin-gonic/gin"
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
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseCursor() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				parsedWantTime, _ := time.Parse(time.RFC3339, tt.wantTime)
				if !gotTime.Equal(parsedWantTime) {
					t.Errorf("gotTime = %v, want %v", gotTime, parsedWantTime)
				}
				if gotID != tt.wantID {
					t.Errorf("gotID = %v, want %v", gotID, tt.wantID)
				}
			}
		})
	}
}

func TestValidateQueryParamsMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	var extractedQuery types.URLQueryParams

	r.GET("/test", middleware.ValidateQueryParams(), func(c *gin.Context) {
		extractedQuery = c.MustGet("queries").(types.URLQueryParams)
		c.Status(http.StatusOK)
	})

	// Test 1: With cursor
	req := httptest.NewRequest(http.MethodGet, "/test?cursor=2026-09-08T12:00:00Z%2B25&limit=50", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	expectedTime, _ := time.Parse(time.RFC3339, "2026-09-08T12:00:00Z")
	if !extractedQuery.CursorTime.Equal(expectedTime) {
		t.Errorf("expected CursorTime %v, got %v", expectedTime, extractedQuery.CursorTime)
	}
	if extractedQuery.CursorID != 25 {
		t.Errorf("expected CursorID 25, got %d", extractedQuery.CursorID)
	}
	if extractedQuery.Limit != 50 {
		t.Errorf("expected Limit 50, got %d", extractedQuery.Limit)
	}

	// Test 2: Invalid cursor returns 400
	reqBad := httptest.NewRequest(http.MethodGet, "/test?cursor=badcursor", nil)
	wBad := httptest.NewRecorder()
	r.ServeHTTP(wBad, reqBad)

	if wBad.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad cursor, got %d", wBad.Code)
	}

	// Test 3: Raw query with plus signs in timezone offset and ID separator
	reqRawPlus := httptest.NewRequest(http.MethodGet, "/test?limit=2&cursor=2026-09-07T18:49:56+03:00+3", nil)
	wRawPlus := httptest.NewRecorder()
	r.ServeHTTP(wRawPlus, reqRawPlus)

	if wRawPlus.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", wRawPlus.Code, wRawPlus.Body.String())
	}
	expectedTimeRaw, _ := time.Parse(time.RFC3339, "2026-09-07T18:49:56+03:00")
	if !extractedQuery.CursorTime.Equal(expectedTimeRaw) {
		t.Errorf("expected CursorTime %v, got %v", expectedTimeRaw, extractedQuery.CursorTime)
	}
	if extractedQuery.CursorID != 3 {
		t.Errorf("expected CursorID 3, got %d", extractedQuery.CursorID)
	}

	// Test 4: Base64 URL-safe cursor query parameter
	b64Cursor := helpers.EncodeCursor("2026-09-07T18:49:56+03:00", 3)
	reqB64 := httptest.NewRequest(http.MethodGet, "/test?limit=10&cursor="+b64Cursor, nil)
	wB64 := httptest.NewRecorder()
	r.ServeHTTP(wB64, reqB64)

	if wB64.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", wB64.Code, wB64.Body.String())
	}
	if !extractedQuery.CursorTime.Equal(expectedTimeRaw) {
		t.Errorf("expected CursorTime %v, got %v", expectedTimeRaw, extractedQuery.CursorTime)
	}
	if extractedQuery.CursorID != 3 {
		t.Errorf("expected CursorID 3, got %d", extractedQuery.CursorID)
	}
	if extractedQuery.Limit != 10 {
		t.Errorf("expected Limit 10, got %d", extractedQuery.Limit)
	}
}

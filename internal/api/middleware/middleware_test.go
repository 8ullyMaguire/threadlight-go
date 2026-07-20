package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func TestCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORS("*"))
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
		t.Errorf("expected 204 or 200, got %d", w.Code)
	}
	if v := w.Header().Get("Access-Control-Allow-Origin"); v != "http://example.com" {
		t.Errorf("expected http://example.com, got %s", v)
	}
}

func TestCORSSpecificOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORS("http://example.com"))
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCORSGET(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORS("*"))
	r.GET("/test", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if v := w.Header().Get("Access-Control-Allow-Origin"); v != "http://example.com" {
		t.Errorf("expected http://example.com, got %s", v)
	}
}

func TestLogging(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Logging(zerolog.Nop()))
	r.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "pong") })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestLoggingMultipleMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Logging(zerolog.Nop()))
	r.POST("/data", func(c *gin.Context) { c.String(http.StatusCreated, "created") })
	r.PUT("/data/1", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	r.DELETE("/data/1", func(c *gin.Context) { c.String(http.StatusNoContent, "") })

	for _, tt := range []struct{ method, path string; expected int }{
		{"POST", "/data", 201},
		{"PUT", "/data/1", 200},
		{"DELETE", "/data/1", 204},
	} {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(tt.method, tt.path, nil)
		r.ServeHTTP(w, req)
		if w.Code != tt.expected {
			t.Errorf("%s %s: expected %d, got %d", tt.method, tt.path, tt.expected, w.Code)
		}
	}
}

func TestCORSHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORS("*"))
	r.GET("/test", func(c *gin.Context) {
		c.Header("X-Custom", "value")
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	r.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("expected Allow-Origin header")
	}
	if w.Header().Get("Access-Control-Allow-Credentials") == "" {
		t.Error("expected Allow-Credentials header")
	}
}

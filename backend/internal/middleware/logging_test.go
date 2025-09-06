package middleware_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/middleware"
)

func TestRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Initialize logger
	err := logger.Init()
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Shutdown()

	r := gin.New()
	r.Use(middleware.RequestID())

	r.GET("/test", func(c *gin.Context) {
		requestID, exists := c.Get("request_id")
		if !exists {
			t.Error("Request ID should be set in context")
		}
		if requestID == "" {
			t.Error("Request ID should not be empty")
		}
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Test without existing request ID
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	requestIDHeader := w.Header().Get("X-Request-ID")
	if requestIDHeader == "" {
		t.Error("X-Request-ID header should be set")
	}

	// Test with existing request ID
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Request-ID", "custom-request-id")
	r.ServeHTTP(w2, req2)

	if w2.Header().Get("X-Request-ID") != "custom-request-id" {
		t.Error("Should preserve existing X-Request-ID header")
	}
}

func TestZapLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Initialize logger
	err := logger.Init()
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Shutdown()

	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.ZapLogger())

	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.GET("/error", func(c *gin.Context) {
		c.JSON(500, gin.H{"error": "test error"})
	})

	// Test successful request
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test error request
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/error", nil)
	r.ServeHTTP(w2, req2)

	if w2.Code != 500 {
		t.Errorf("Expected status 500, got %d", w2.Code)
	}
}

func TestZapRecovery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Initialize logger
	err := logger.Init()
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Shutdown()

	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.ZapRecovery())

	r.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	// Test panic recovery
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/panic", nil)
	r.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("Expected status 500 after panic recovery, got %d", w.Code)
	}
}

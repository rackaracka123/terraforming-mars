package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"terraforming-mars-backend/internal/logger"
)

// RequestID middleware adds a request ID to the context
func RequestID() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	})
}

// ZapLogger middleware logs HTTP requests using Zap
func ZapLogger() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		
		// Process request
		c.Next()
		
		// Calculate request duration
		duration := time.Since(start)
		
		// Get request ID from context
		requestID, _ := c.Get("request_id")
		
		// Build log fields
		fields := []zap.Field{
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Duration("duration", duration),
			zap.Int("size", c.Writer.Size()),
		}
		
		// Add request ID if present
		if requestID != nil {
			fields = append(fields, zap.String("request_id", requestID.(string)))
		}
		
		// Add query parameters if present
		if raw != "" {
			fields = append(fields, zap.String("query", raw))
		}
		
		// Log based on status code
		status := c.Writer.Status()
		msg := "HTTP Request"
		
		if len(c.Errors) > 0 {
			// Log errors
			for _, err := range c.Errors {
				logger.Get().Error("HTTP Request Error",
					append(fields, zap.String("error", err.Error()))...)
			}
		} else if status >= 500 {
			logger.Get().Error(msg, fields...)
		} else if status >= 400 {
			logger.Get().Warn(msg, fields...)
		} else {
			logger.Get().Info(msg, fields...)
		}
	})
}

// Recovery middleware with Zap logging
func ZapRecovery() gin.HandlerFunc {
	return gin.RecoveryWithWriter(gin.DefaultWriter, func(c *gin.Context, err interface{}) {
		requestID, _ := c.Get("request_id")
		
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
			zap.Any("error", err),
		}
		
		if requestID != nil {
			fields = append(fields, zap.String("request_id", requestID.(string)))
		}
		
		logger.Get().Error("Panic recovered", fields...)
		c.AbortWithStatus(500)
	})
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + time.Now().Format("000")
}
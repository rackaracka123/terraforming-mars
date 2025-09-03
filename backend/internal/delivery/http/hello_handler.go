package http

import (
	"net/http"
	"terraforming-mars-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// HelloHandler handles HTTP requests for hello endpoints
type HelloHandler struct {
	helloService *service.HelloService
}

// NewHelloHandler creates a new hello handler
func NewHelloHandler(helloService *service.HelloService) *HelloHandler {
	return &HelloHandler{
		helloService: helloService,
	}
}

// GetHello handles GET /hello endpoint
func (h *HelloHandler) GetHello(c *gin.Context) {
	message := h.helloService.GetMessage()
	
	c.JSON(http.StatusOK, gin.H{
		"message": message,
	})
}

// HealthCheck handles health check endpoint
func (h *HelloHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"service": "terraforming-mars-backend",
	})
}
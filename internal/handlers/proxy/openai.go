package proxyhandler

import (
	"net/http"

	"apihub/internal/models"
	"apihub/internal/proxy"

	"github.com/gin-gonic/gin"
)

type OpenAIHandler struct {
	proxy *proxy.OpenAIProxy
}

func NewOpenAIHandler(p *proxy.OpenAIProxy) *OpenAIHandler {
	return &OpenAIHandler{proxy: p}
}

func (h *OpenAIHandler) ChatCompletions(c *gin.Context) {
	apiKey := c.MustGet("api_key").(*models.APIKey)
	h.proxy.HandleChatCompletions(c.Writer, c.Request, apiKey)
}

func (h *OpenAIHandler) Completions(c *gin.Context) {
	apiKey := c.MustGet("api_key").(*models.APIKey)
	h.proxy.HandleCompletions(c.Writer, c.Request, apiKey)
}

func (h *OpenAIHandler) ListModels(c *gin.Context) {
	apiKey := c.MustGet("api_key").(*models.APIKey)
	h.proxy.HandleListModels(c.Writer, c.Request, apiKey)
}

func (h *OpenAIHandler) NotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"error": gin.H{
			"message": "unknown endpoint",
			"type":    "invalid_request_error",
		},
	})
}

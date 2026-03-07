package proxyhandler

import (
	"apihub/internal/models"
	"apihub/internal/proxy"

	"github.com/gin-gonic/gin"
)

type AnthropicHandler struct {
	proxy *proxy.AnthropicProxy
}

func NewAnthropicHandler(p *proxy.AnthropicProxy) *AnthropicHandler {
	return &AnthropicHandler{proxy: p}
}

func (h *AnthropicHandler) Messages(c *gin.Context) {
	apiKey := c.MustGet("api_key").(*models.APIKey)
	h.proxy.HandleMessages(c.Writer, c.Request, apiKey)
}

func (h *AnthropicHandler) ListModels(c *gin.Context) {
	apiKey := c.MustGet("api_key").(*models.APIKey)
	h.proxy.HandleListModels(c.Writer, c.Request, apiKey)
}

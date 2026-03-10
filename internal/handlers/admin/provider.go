package admin

import (
	"net/http"
	"strconv"
	"strings"

	"apihub/internal/models"
	"apihub/internal/services"

	"github.com/gin-gonic/gin"
)

type ProviderHandler struct {
	providerService *services.ProviderService
	syncService     *services.SyncService
}

func NewProviderHandler(providerService *services.ProviderService, syncService *services.SyncService) *ProviderHandler {
	return &ProviderHandler{providerService: providerService, syncService: syncService}
}

type createProviderRequest struct {
	Name     string   `json:"name" binding:"required"`
	Protocol string   `json:"protocol" binding:"required,oneof=openai anthropic"`
	BaseURL  string   `json:"base_url" binding:"required,url"`
	APIKey   string   `json:"api_key" binding:"required"`
	Tags     []string `json:"tags"`
}

func (h *ProviderHandler) Create(c *gin.Context) {
	var req createProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	provider := &models.Provider{
		Name:       req.Name,
		Protocol:   models.ProviderProtocol(req.Protocol),
		BaseURL:    req.BaseURL,
		APIKey:     req.APIKey,
		Enabled:    true,
		SyncStatus: "syncing",
		SyncError:  "",
		Tags:       strings.Join(req.Tags, ","),
	}

	if err := h.providerService.Create(provider); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Trigger async model sync
	go h.syncService.SyncProvider(provider.ID)

	c.JSON(http.StatusCreated, provider)
}

func (h *ProviderHandler) List(c *gin.Context) {
	providers, err := h.providerService.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, providers)
}

func (h *ProviderHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	provider, err := h.providerService.FindByIDWithModels(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
		return
	}
	c.JSON(http.StatusOK, provider)
}

type updateProviderRequest struct {
	Name         string  `json:"name"`
	BaseURL      string  `json:"base_url"`
	APIKey       string  `json:"api_key"`
	SyncInterval *string `json:"sync_interval"`
}

func (h *ProviderHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	provider, err := h.providerService.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
		return
	}

	var req updateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != "" {
		provider.Name = req.Name
	}
	if req.BaseURL != "" {
		provider.BaseURL = req.BaseURL
	}
	if req.APIKey != "" {
		provider.APIKey = req.APIKey
	}
	if req.Tags != nil {
		provider.Tags = strings.Join(req.Tags, ",")
	}

	// 同步频率变更
	syncIntervalChanged := false
	if req.SyncInterval != nil {
		provider.SyncInterval = *req.SyncInterval
		syncIntervalChanged = true
	}

	if err := h.providerService.Update(provider); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 同步频率变更时刷新调度器
	if syncIntervalChanged {
		h.syncService.RefreshScheduler()
	}

	c.JSON(http.StatusOK, provider)
}

func (h *ProviderHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.providerService.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *ProviderHandler) Toggle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	provider, err := h.providerService.Toggle(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, provider)
}

func (h *ProviderHandler) Sync(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.syncService.SyncProvider(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "sync completed"})
}

func (h *ProviderHandler) SyncAll(c *gin.Context) {
	if err := h.syncService.SyncAllProviders(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "sync completed"})
}

func (h *ProviderHandler) ListModels(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	models, err := h.providerService.GetModels(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, models)
}

func (h *ProviderHandler) ToggleModel(c *gin.Context) {
	mid, err := strconv.ParseUint(c.Param("mid"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model id"})
		return
	}

	pm, err := h.providerService.ToggleModel(uint(mid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pm)
}

// SyncAll 手动触发同步所有 Provider
func (h *ProviderHandler) SyncAll(c *gin.Context) {
	go h.syncService.SyncAllProviders()
	c.JSON(http.StatusOK, gin.H{"message": "sync all started"})
}

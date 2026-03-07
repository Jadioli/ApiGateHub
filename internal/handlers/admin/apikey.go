package admin

import (
	"net/http"
	"strconv"
	"time"

	"apihub/internal/models"
	"apihub/internal/services"

	"github.com/gin-gonic/gin"
)

type APIKeyHandler struct {
	apiKeyService *services.APIKeyService
}

func NewAPIKeyHandler(apiKeyService *services.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{apiKeyService: apiKeyService}
}

type createAPIKeyRequest struct {
	Name          string  `json:"name" binding:"required"`
	ModelConfigID *uint   `json:"model_config_id"`
	ExpiresAt     *string `json:"expires_at"`
}

func (h *APIKeyHandler) Create(c *gin.Context) {
	var req createAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	keyTemplate := &models.APIKey{ModelConfigID: req.ModelConfigID}
	if req.ExpiresAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expires_at format, use RFC3339"})
			return
		}
		keyTemplate.ExpiresAt = &t
	}

	result, err := h.apiKeyService.Create(req.Name, keyTemplate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result.APIKey)
}

func (h *APIKeyHandler) List(c *gin.Context) {
	keys, err := h.apiKeyService.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, keys)
}

func (h *APIKeyHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	key, err := h.apiKeyService.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, key)
}

func (h *APIKeyHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	key, err := h.apiKeyService.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Name != "" {
		key.Name = req.Name
	}
	if err := h.apiKeyService.Update(key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, key)
}

func (h *APIKeyHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.apiKeyService.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *APIKeyHandler) Toggle(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	key, err := h.apiKeyService.Toggle(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, key)
}

// --- Model mapping endpoints ---

func (h *APIKeyHandler) ListModels(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	ms, err := h.apiKeyService.GetModels(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ms)
}

func (h *APIKeyHandler) AddModel(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req struct {
		ProviderID      uint   `json:"provider_id" binding:"required"`
		ProviderModelID uint   `json:"provider_model_id" binding:"required"`
		MappedName      string `json:"mapped_name"`
		Priority        int    `json:"priority"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	m := &models.APIKeyModel{
		APIKeyID:        uint(id),
		ProviderID:      req.ProviderID,
		ProviderModelID: req.ProviderModelID,
		MappedName:      req.MappedName,
		Priority:        req.Priority,
		Enabled:         true,
	}

	if err := h.apiKeyService.AddModel(m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, m)
}

func (h *APIKeyHandler) UpdateModel(c *gin.Context) {
	mid, _ := strconv.ParseUint(c.Param("mid"), 10, 32)

	var req struct {
		MappedName *string `json:"mapped_name"`
		Priority   *int    `json:"priority"`
		Enabled    *bool   `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.MappedName != nil {
		updates["mapped_name"] = *req.MappedName
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	m := &models.APIKeyModel{}
	m.ID = uint(mid)
	if err := h.apiKeyService.UpdateModel(m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *APIKeyHandler) DeleteModel(c *gin.Context) {
	mid, _ := strconv.ParseUint(c.Param("mid"), 10, 32)
	if err := h.apiKeyService.DeleteModel(uint(mid)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *APIKeyHandler) ListModelsGrouped(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	grouped, err := h.apiKeyService.GetModelsGrouped(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, grouped)
}

// BatchUpdate replaces all model mappings for an API Key in one call.
func (h *APIKeyHandler) BatchUpdate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req []struct {
		ProviderID      uint   `json:"provider_id"`
		ProviderModelID uint   `json:"provider_model_id"`
		MappedName      string `json:"mapped_name"`
		Priority        int    `json:"priority"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ms := make([]models.APIKeyModel, len(req))
	for i, r := range req {
		ms[i] = models.APIKeyModel{
			ProviderID:      r.ProviderID,
			ProviderModelID: r.ProviderModelID,
			MappedName:      r.MappedName,
			Priority:        r.Priority,
		}
	}

	if err := h.apiKeyService.BatchReplace(uint(id), ms); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// ModelConfig operations

func (h *APIKeyHandler) GetModelConfig(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	config, err := h.apiKeyService.GetModelConfig(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if config == nil {
		c.JSON(http.StatusOK, gin.H{"model_config": nil})
		return
	}
	c.JSON(http.StatusOK, config)
}

func (h *APIKeyHandler) SetModelConfig(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req struct {
		ModelConfigID *uint `json:"model_config_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.apiKeyService.SetModelConfig(uint(id), req.ModelConfigID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

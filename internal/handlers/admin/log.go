package admin

import (
	"net/http"
	"strconv"

	"apihub/internal/repository"

	"github.com/gin-gonic/gin"
)

type LogHandler struct {
	logRepo *repository.LogRepo
}

func NewLogHandler(logRepo *repository.LogRepo) *LogHandler {
	return &LogHandler{logRepo: logRepo}
}

func (h *LogHandler) List(c *gin.Context) {
	q := repository.LogQuery{
		Page:     1,
		PageSize: 20,
	}

	if p, err := strconv.Atoi(c.Query("page")); err == nil {
		q.Page = p
	}
	if ps, err := strconv.Atoi(c.Query("page_size")); err == nil {
		q.PageSize = ps
	}
	if keyID, err := strconv.ParseUint(c.Query("api_key_id"), 10, 32); err == nil {
		id := uint(keyID)
		q.APIKeyID = &id
	}
	if model := c.Query("model"); model != "" {
		q.Model = model
	}
	if status, err := strconv.Atoi(c.Query("status")); err == nil {
		q.Status = &status
	}

	result, err := h.logRepo.Query(q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *LogHandler) Dashboard(c *gin.Context) {
	stats, err := h.logRepo.GetDashboardStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

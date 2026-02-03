package handlers

import (
	"gibraltar/config"
	"gibraltar/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ConfigsHandler struct {
	Cache *services.Cache
}

func NewConfigsHandler(cache *services.Cache) *ConfigsHandler {
	return &ConfigsHandler{
		Cache: cache,
	}
}

func (h *ConfigsHandler) CurrentAvailableConfigs(c *gin.Context) {
	configs, ok := h.Cache.Get(config.AvailableKey)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "configs unavailable retry later",
		})
		return
	}
	resultString := ""
	for _, v := range configs {
		resultString += v.URL + "\n"
	}
	c.String(http.StatusOK, resultString)
}

package handlers

import (
	"gibraltar/config"
	"gibraltar/internal/services"
	"net/http"
	"strconv"

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
	configs, ok := h.Cache.GetString(config.AvailableKey)
	if !ok {
		c.JSON(http.StatusNoContent, gin.H{
			"error": "configs unavailable retry later",
		})
		return
	}
	showTags, err := strconv.ParseBool(c.Query("tags"))
	if err != nil {
		showTags = true
	}
	var resultString = configs
	if showTags {
		resultString = config.Tags + resultString
	}
	c.String(http.StatusOK, resultString)
}

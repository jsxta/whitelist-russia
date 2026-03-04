package services

import (
	"gibraltar/config"
	"gibraltar/internal/models"
	"strings"
	"sync"
)

type Cache struct {
	mu    sync.RWMutex
	cache map[string][]models.AnyConfig
	str   map[string]string
}

func NewCache() *Cache {
	cacheMap := make(map[string][]models.AnyConfig)
	strMap := make(map[string]string)
	return &Cache{
		cache: cacheMap,
		str:   strMap,
	}
}

func (c *Cache) Set(id string, data []models.AnyConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[id] = data
	if id == config.AllKey {
		return
	}
	var str strings.Builder
	for _, v := range data {
		str.WriteString(v.GetURL())
		str.WriteString("\n")
	}
	c.str[id] = str.String()
}

func (c *Cache) Get(id string) ([]models.AnyConfig, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	res, ok := c.cache[id]
	return res, ok
}

func (c *Cache) GetString(id string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	res, ok := c.str[id]
	return res, ok
}

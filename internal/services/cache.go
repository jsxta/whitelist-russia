package services

import (
	"gibraltar/internal/models"
	"sync"
)

type Cache struct {
	mu    sync.RWMutex
	cache map[string][]models.AnyConfig
}

func NewCache() *Cache {
	cacheMap := make(map[string][]models.AnyConfig)
	return &Cache{
		cache: cacheMap,
	}
}

func (c *Cache) Set(id string, data []models.AnyConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[id] = data
}

func (c *Cache) Get(id string) ([]models.AnyConfig, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	res, ok := c.cache[id]
	return res, ok
}

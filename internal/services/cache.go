package services

import (
	"gibraltar/internal/models"
	"sync"
)

type Cache struct {
	mu    sync.RWMutex
	cache map[string][]models.VlessConfig
}

func NewCache() *Cache {
	cacheMap := make(map[string][]models.VlessConfig)
	return &Cache{
		cache: cacheMap,
	}
}

func (c *Cache) Set(id string, data []models.VlessConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[id] = data
}

func (c *Cache) Get(id string) ([]models.VlessConfig, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	res, ok := c.cache[id]
	return res, ok
}

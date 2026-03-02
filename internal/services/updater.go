package services

import (
	"fmt"
	"gibraltar/config"
	"gibraltar/internal/models"
	"gibraltar/internal/parser"
	"log"
	"net/url"
	"strings"
	"time"
)

type DataParser interface {
	ParseConfigs() ([]models.AnyConfig, error)
	ParseSubnets() (map[string]struct{}, error)
	ParseSNIs() (map[string]struct{}, error)
}

type ConfigUpdater struct {
	Cache          *Cache
	Filter         *ConfigFilter
	URLTestService *URLTestService
	DataSource     DataParser
}

func NewConfigUpdater(cache *Cache, filter *ConfigFilter, urlTestService *URLTestService, dataSource DataParser) *ConfigUpdater {
	return &ConfigUpdater{
		Cache:          cache,
		Filter:         filter,
		URLTestService: urlTestService,
		DataSource:     dataSource,
	}
}

func (u *ConfigUpdater) RunTest(configs []models.AnyConfig) {
	start := time.Now()
	defer func() {
		log.Printf("woriking time (test): %s\n", time.Since(start))
	}()
	u.URLTestService.TestConfigs(configs, len(configs)/32)

}

func (u *ConfigUpdater) AddConfigsToCacheFromSource() error {
	configs, err := u.DataSource.ParseConfigs()
	if err != nil {
		return err
	}

	filtered := make([]models.AnyConfig, 0)
	for _, cfg := range configs {
		if err = parseConfig(cfg); err != nil {
			continue
		}
		if ok, _ := u.Filter.IsAvailableConfig(cfg); !ok {
			continue
		}
		filtered = append(filtered, cfg)
	}
	prevConfigs, ok := u.Cache.Get(config.AllKey)
	if !ok {
		u.Cache.Set(config.AllKey, filtered)
		return nil
	}

	prevMap := make(map[string]models.AnyConfig, len(prevConfigs))
	for _, prev := range prevConfigs {
		key, err := getKeyByUrl(prev.GetURL())
		if err != nil {
			continue
		}
		prevMap[key] = prev
	}

	for _, cfg := range filtered {
		key, err := getKeyByUrl(cfg.GetURL())
		if err != nil {
			continue
		}
		if old, ok := prevMap[key]; ok {
			cfg.SetStability(old.GetStability())
		}
		prevMap[key] = cfg
	}

	result := make([]models.AnyConfig, 0, len(prevMap))
	for _, v := range prevMap {
		result = append(result, v)
	}
	u.Cache.Set(config.AllKey, result)

	return nil
}

func parseConfig(config models.AnyConfig) error {

	switch c := config.(type) {
	case *models.VlessConfig:
		return parser.ParseVless(c)
	case *models.TrojanConfig:
		return parser.ParseTrojan(c)
	case *models.ShadowsocksConfig:
		return parser.ParseShadowsocks(c)
	default:
		return fmt.Errorf("unknown scheme")
	}
}

func (u *ConfigUpdater) AddAvailableConfigsToCache() error {
	configs, ok := u.Cache.Get(config.AllKey)
	if !ok {
		if err := u.AddConfigsToCacheFromSource(); err != nil {
			return err
		}
		configs, _ = u.Cache.Get(config.AllKey)
	}

	u.RunTest(configs)

	availableList := deleteDuplicates(filterConfigsByStability(configs))
	if len(availableList) != 0 {
		u.Cache.Set(config.AvailableKey, availableList)
	}
	return nil
}

func filterConfigsByStability(configs []models.AnyConfig) []models.AnyConfig {
	result := make([]models.AnyConfig, 0, 10)
	for _, cfg := range configs {
		if cfg.GetStability() >= config.MinValueForAccept {
			if cfg.GetStability() >= config.MinValueForStable {
				markAsStable(cfg)
			}
			result = append(result, cfg)
		}
	}
	return result
}

func markAsStable(cfg models.AnyConfig) {
	s := cfg.GetURL()
	idx := strings.IndexByte(s, '#')
	if idx == -1 {
		cfg.SetURL(s + "#Стабильный ")
		return
	}
	after := s[idx+1:]
	if strings.HasPrefix(after, "Стабильный | ") || strings.HasPrefix(after, "Stable | ") {
		return
	}
	cfg.SetURL(s[:idx+1] + "Стабильный | " + after)
}

func deleteDuplicates(configs []models.AnyConfig) []models.AnyConfig {
	resultMap := make(map[string]models.AnyConfig, len(configs))
	for _, cfg := range configs {
		u, err := url.Parse(cfg.GetURL())
		if err != nil {
			continue
		}
		base := u.Scheme + "://" + u.Host
		if old, ok := resultMap[base]; !ok {
			resultMap[base] = cfg
		} else if old.GetStability() < cfg.GetStability() {
			resultMap[base] = cfg
		}
	}

	result := make([]models.AnyConfig, 0, len(resultMap))
	for _, cfg := range resultMap {
		result = append(result, cfg)
	}
	return result
}

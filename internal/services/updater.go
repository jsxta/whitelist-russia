package services

import (
	"gibraltar/config"
	"gibraltar/internal/models"
	"log"
	"net/url"
	"strings"
	"time"
)

type DataParser interface {
	ParseConfigs() ([]*models.VlessConfig, error)
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

func (u *ConfigUpdater) RunTest(configs []*models.VlessConfig) {
	start := time.Now()
	defer func() {
		log.Printf("woriking time (test): %s\n", time.Since(start))
	}()
	u.URLTestService.TestConfigs(configs, len(configs)/4)

}

func (u *ConfigUpdater) AddConfigsToCacheFromSource() error {
	configs, err := u.DataSource.ParseConfigs()
	if err != nil {
		return err
	}
	filtered := make([]models.VlessConfig, 0, len(configs)/10)
	for _, config := range configs {
		err = parseVlessURL(config)
		if err != nil {
			continue
		}
		if ok, _ := u.Filter.IsAvailableConfig(config); !ok {
			continue
		}
		filtered = append(filtered, *config)
	}
	prevConfigs, ok := u.Cache.Get(config.AllKey)
	if !ok {
		u.Cache.Set(config.AllKey, filtered)
	} else {
		prevMap := make(map[string]models.VlessConfig, len(prevConfigs))
		for i := 0; i < len(prevConfigs); i++ {
			pre := prevConfigs[i].URL
			link, err := url.Parse(pre)
			if err != nil {
				continue
			}
			key := link.Scheme + "://" + link.User.String() + "@" + link.Host
			prevMap[key] = prevConfigs[i]
		}
		for i := 0; i < len(filtered); i++ {
			pre := filtered[i].URL
			link, err := url.Parse(pre)
			if err != nil {
				continue
			}
			key := link.Scheme + "://" + link.User.String() + "@" + link.Host
			prevMap[key] = filtered[i]
		}
		result := make([]models.VlessConfig, 0, len(prevMap))
		for _, v := range prevMap {
			result = append(result, v)
		}
		u.Cache.Set(config.AllKey, result)

	}

	return nil

}

func (u *ConfigUpdater) AddAvailableConfigsToCache() error {
	configs, ok := u.Cache.Get(config.AllKey)
	if !ok {
		if err := u.AddConfigsToCacheFromSource(); err != nil {
			return err
		}
		configs, _ = u.Cache.Get(config.AllKey)
	}
	pointers := make([]*models.VlessConfig, 0, len(configs))
	for idx := range configs {
		pointers = append(pointers, &configs[idx])
	}
	u.RunTest(pointers)
	availableList := deleteDuplicates(filterConfigsByStability(pointers))
	u.Cache.Set(config.AvailableKey, *availableList)
	return nil
}

func filterConfigsByStability(configs []*models.VlessConfig) *[]models.VlessConfig {
	if configs == nil {
		return nil
	}
	result := make([]models.VlessConfig, 0, 10)
	for idx := range configs {
		if configs[idx].Stability >= config.MinValueForAccept {
			result = append(result, *configs[idx])

		}
		if configs[idx].Stability >= config.MinValueForStable {
			markAsStable(&result[len(result)-1])
		}
	}
	return &result
}

func markAsStable(cfg *models.VlessConfig) {
	if cfg == nil {
		return
	}
	s := cfg.URL
	idx := strings.IndexByte(s, '#')
	if idx == -1 {
		cfg.URL = s + "#Стабильный "
		return
	}
	before := s[:idx+1]
	after := s[idx+1:]

	if strings.HasPrefix(after, "Стабильный | ") || strings.HasPrefix(after, "Stable | ") {
		return
	}

	cfg.URL = before + "Стабильный | " + after
}

func deleteDuplicates(configs *[]models.VlessConfig) *[]models.VlessConfig {

	resultMap := make(map[string]*models.VlessConfig, len(*configs))
	for i := 0; i < len(*configs); i++ {
		u, err := url.Parse((*configs)[i].URL)
		if err != nil {
			continue
		}
		base := u.Scheme + "://" + u.Host
		if data, ok := resultMap[base]; !ok {
			resultMap[base] = &(*configs)[i]
		} else {
			if data.Stability < (*configs)[i].Stability {
				resultMap[base] = &(*configs)[i]
			}

		}

	}
	result := make([]models.VlessConfig, 0, len(resultMap))
	for _, cfg := range resultMap {
		result = append(result, *cfg)
	}
	return &result
}

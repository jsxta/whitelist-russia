package services

import (
	"fmt"
	"gibraltar/internal/models"
)

type ConfigFilter struct {
	AllowedSubnets map[string]struct{}
	AllowedSNIs    map[string]struct{}
}

func NewConfigFilter(allowedSubnets map[string]struct{}, allowedSNIs map[string]struct{}) *ConfigFilter {
	return &ConfigFilter{
		AllowedSubnets: allowedSubnets,
		AllowedSNIs:    allowedSNIs,
	}
}

func (f ConfigFilter) IsAvailableConfig(config *models.VlessConfig) (bool, error) {
	ok, err := f.isIPFromWhitelist(config.Server)
	if !ok {
		return false, err
	}
	return true, nil
}

func (f *ConfigFilter) isIPFromWhitelist(ip string) (bool, error) {
	if len(ip) == 0 {
		return false, fmt.Errorf("empty ip")
	}
	ipSubnet := getSubnet(ip)
	_, ok := f.AllowedSubnets[string(ipSubnet)]
	if !ok {
		return false, fmt.Errorf("ip not exist in whitelist")
	}
	return true, nil
}

func (f *ConfigFilter) isSNIFromWhitelist(sni string) (bool, error) {
	_, ok := f.AllowedSNIs[sni]
	if !ok {
		return false, fmt.Errorf("sni not exist in whitelist")
	}
	return true, nil
}

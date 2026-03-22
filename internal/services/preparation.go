package services

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"gibraltar/internal/models"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type IPListSource interface {
	ParseSubnets() (map[string]struct{}, error)
}

type FileParser struct {
	ConifgPath   string
	IPListSource IPListSource
	SNIsPath     string
}

func NewFileParser(configPath, snisPath string, IPListSource IPListSource) *FileParser {
	return &FileParser{
		ConifgPath:   configPath,
		IPListSource: IPListSource,
		SNIsPath:     snisPath,
	}
}

func (p *FileParser) ParseConfigs() ([]*models.VlessConfig, error) {
	file, err := os.Open(p.ConifgPath)
	if err != nil {
		return nil, err

	}
	defer file.Close()
	configs := make([]*models.VlessConfig, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		line = strings.ReplaceAll(line, "&amp;", "&")
		if line == "" {
			continue
		}
		config := &models.VlessConfig{
			BaseConfig: models.BaseConfig{
				URL: line,
			},
		}

		configs = append(configs, config)
	}
	return configs, nil
}

type FileIPList struct {
	Path string
}

func (s *FileIPList) ParseSubnets() (map[string]struct{}, error) {
	file, err := os.Open(s.Path)
	if err != nil {
		return nil, err
	}
	subnetsMap := make(map[string]struct{})
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		subnetsMap[string(getSubnet(line))] = struct{}{}
	}

	return subnetsMap, nil
}

func (p *FileParser) GetSNIsFromFile() (map[string]struct{}, error) {
	file, err := os.Open(p.SNIsPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	allowedSNIs := make(map[string]struct{})
	for scanner.Scan() {
		line := scanner.Text()
		allowedSNIs[line] = struct{}{}
	}
	return allowedSNIs, nil

}

type UrlParser struct {
	ConfigsURLs  []string
	IPListSource IPListSource
	SNIsURL      string
}

func NewUrlParser(configsURLs []string, snisURL string, IPListSource IPListSource) *UrlParser {
	return &UrlParser{
		ConfigsURLs:  configsURLs,
		IPListSource: IPListSource,
		SNIsURL:      snisURL,
	}
}

func (p *UrlParser) ParseConfigs() ([]models.AnyConfig, error) {
	configs := make(map[string]string)
	for _, source := range p.ConfigsURLs {
		resp, err := http.Get(source)
		if err != nil {
			log.Println(err)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		content := decodeIfBase64(string(body))

		for _, line := range strings.Split(string(content), "\n") {
			line = strings.TrimSpace(line)
			line = strings.ReplaceAll(line, "&amp;", "&")

			if line == "" {
				continue
			}
			u, err := url.Parse(line)
			if err != nil {
				continue
			}
			q := u.Query()
			security := strings.TrimSpace(strings.ToLower(q.Get("security")))
			switch u.Scheme {
			case "vless":
				if security == "none" || security == "" {
					continue
				}
			case "trojan":
				if security == "none" {
					continue
				}
			case "ss":
				if security == "none" {
					continue
				}
			default:
				continue
			}

			fragment := u.Fragment
			link := deleteFragment(line)
			key, _ := getKeyByUrl(link)
			configs[key] = link + "#" + fragment
		}
		resp.Body.Close()
	}
	result := make([]models.AnyConfig, 0, len(configs))
	for _, fullUrl := range configs {
		u, _ := url.Parse(fullUrl)
		var config models.AnyConfig

		switch u.Scheme {
		case "vless":
			c := &models.VlessConfig{BaseConfig: models.BaseConfig{URL: fullUrl}}
			config = c
		case "trojan":
			c := &models.TrojanConfig{BaseConfig: models.BaseConfig{URL: fullUrl}}
			config = c
		case "ss":
			c := &models.ShadowsocksConfig{BaseConfig: models.BaseConfig{URL: fullUrl}}
			config = c
		default:
			continue
		}

		result = append(result, config)
	}
	return result, nil
}

type URLIPList struct {
	URL string
}

func (s *URLIPList) ParseSubnets() (map[string]struct{}, error) {
	resp, err := http.Get(s.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	subnetsMap := make(map[string]struct{})
	for scanner.Scan() {
		line := scanner.Text()
		subnetsMap[string(getSubnet(line))] = struct{}{}
	}

	return subnetsMap, nil
}

func (p *UrlParser) ParseSNIs() (map[string]struct{}, error) {
	resp, err := http.Get(p.SNIsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	allowedSNIs := make(map[string]struct{})
	for scanner.Scan() {
		line := scanner.Text()
		allowedSNIs[line] = struct{}{}
	}
	return allowedSNIs, nil
}

func getSubnet(ip string) []byte {
	num := make([]byte, 0)
	dotCount := 0
	for _, ch := range ip {
		if ch == rune('.') {
			if dotCount == 2 {
				break

			}
			dotCount++

		}
		num = append(num, byte(ch))
	}
	return num
}

func deleteFragment(s string) string {
	for i, ch := range s {
		if ch == '#' {
			return s[:i]
		}
	}
	return s
}

func getKeyByUrl(link string) (string, error) {
	parsedUrl, err := url.Parse(link)
	if err != nil {
		return "", err
	}
	key := parsedUrl.String()
	return key, nil
}

func decodeIfBase64(data string) string {
	if strings.Contains(data, "://") {
		return data
	}

	trimmed := strings.TrimSpace(data)

	padded := trimmed
	if pad := len(padded) % 4; pad != 0 {
		padded += strings.Repeat("=", 4-pad)
	}

	decoded, err := base64.StdEncoding.DecodeString(padded)
	if err != nil {
		decoded, err = base64.URLEncoding.DecodeString(padded)
		if err != nil {
			return data
		}
	}

	if !bytes.Contains(decoded, []byte("://")) {
		return data
	}

	return string(decoded)
}

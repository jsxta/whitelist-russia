package services

import (
	"bufio"
	"gibraltar/internal/models"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type FileParser struct {
	ConifgPath string
	IPListPath string
	SNIsPath   string
}

func NewFileParser(configPath, ipListPath, snisPath string) *FileParser {
	return &FileParser{
		ConifgPath: configPath,
		IPListPath: ipListPath,
		SNIsPath:   snisPath,
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
			URL: line,
		}

		configs = append(configs, config)
	}
	return configs, nil
}

func (p *FileParser) ParseSubnets() (map[string]struct{}, error) {
	file, err := os.Open(p.IPListPath)
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
	ConfigsURLs []string
	IPListURL   string
	SNIsURL     string
}

func NewUrlParser(configsURLs []string, ipListURL, snisURL string) *UrlParser {
	return &UrlParser{
		ConfigsURLs: configsURLs,
		IPListURL:   ipListURL,
		SNIsURL:     snisURL,
	}
}

func (p *UrlParser) ParseConfigs() ([]*models.VlessConfig, error) {
	configs := make(map[string]string)
	for _, source := range p.ConfigsURLs {
		resp, err := http.Get(source)
		if err != nil {
			log.Println(err)
			continue
		}
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
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
			if security == "none" || security == "" {
				continue
			}

			fragment := u.Fragment
			link := deleteFragment(line)
			key, _ := getKeyByUrl(link)
			configs[key] = link + "#" + fragment
		}
	}
	result := make([]*models.VlessConfig, 0, len(configs))
	for _, fullUrl := range configs {
		result = append(result, &models.VlessConfig{
			URL: fullUrl,
		})
	}
	return result, nil
}

func (p *UrlParser) ParseSubnets() (map[string]struct{}, error) {
	resp, err := http.Get(p.IPListURL)
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
	key := parsedUrl.Scheme + "://" + parsedUrl.User.String() + "@" + parsedUrl.Host
	return key, nil
}

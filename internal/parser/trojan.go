package parser

import (
	"errors"
	"gibraltar/internal/models"
	"net/url"
	"strconv"
)

func ParseTrojan(config *models.TrojanConfig) error {
	u, err := url.Parse(config.URL)
	if err != nil {
		return err
	}
	if u.Scheme != "trojan" {
		return errors.New("not trojan url")
	}
	if !validateIP(u.Hostname()) {
		return errors.New("invalid ip")
	}

	port, _ := strconv.Atoi(u.Port())
	q := u.Query()

	config.Password = u.User.Username()
	config.Server = u.Hostname()
	config.Port = port
	config.Security = q.Get("security")
	config.SNI = q.Get("sni")
	config.PublicKey = q.Get("pbk")
	config.SID = q.Get("sid")
	config.Fingerprint = q.Get("fp")
	config.Type = q.Get("type")
	config.Path = q.Get("path")
	config.Host = q.Get("host")
	config.ServiceName = q.Get("serviceName")

	return nil
}

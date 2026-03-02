package parser

import (
	"errors"
	"gibraltar/internal/models"
	"net"
	"net/url"
	"strconv"
)

func ParseVless(config *models.VlessConfig) error {
	u, err := url.Parse(config.URL)
	if err != nil {
		return err
	}
	if u.Scheme != "vless" {
		return errors.New("not vless url")
	}

	if !validateIP(u.Hostname()) {
		return errors.New("invalid ip")
	}

	port, _ := strconv.Atoi(u.Port())
	q := u.Query()

	config.UUID = u.User.Username()
	config.Server = u.Hostname()
	config.Port = port
	config.Security = q.Get("security")
	config.SNI = q.Get("sni")
	config.PublicKey = q.Get("pbk")
	config.SID = q.Get("sid")
	config.Fingerprint = q.Get("fp")
	config.Type = q.Get("type")
	config.SPX = q.Get("spx")
	config.Flow = q.Get("flow")
	config.Path = q.Get("path")
	config.Host = q.Get("host")
	config.ServiceName = q.Get("serviceName")
	config.HeaderType = q.Get("headerType")

	return nil
}

func validateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil
}

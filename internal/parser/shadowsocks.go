package parser

import (
	"encoding/base64"
	"errors"
	"fmt"
	"gibraltar/internal/models"
	"net/url"
	"strconv"
	"strings"
)

func ParseShadowsocks(config *models.ShadowsocksConfig) error {
	u, err := url.Parse(config.URL)
	if err != nil {
		return err
	}
	if u.Scheme != "ss" {
		return errors.New("not shadowsocks url")
	}
	if !validateIP(u.Hostname()) {
		return errors.New("invalid ip")
	}

	port, _ := strconv.Atoi(u.Port())
	config.Server = u.Hostname()
	config.Port = port

	if u.User != nil {
		// SIP002: ss://base64(method:password)@host:port
		if err := decodeShadowsocksUserinfo(u.User.Username(), config); err != nil {
			return err
		}
	} else {
		// Legacy: ss://base64(method:password@host:port)
		decoded, err := base64Decode(strings.TrimPrefix(config.URL, "ss://"))
		if err != nil {
			return fmt.Errorf("failed to decode legacy ss url: %w", err)
		}
		lu, err := url.Parse("ss://" + decoded)
		if err != nil {
			return err
		}
		if err := decodeShadowsocksUserinfo(lu.User.Username(), config); err != nil {
			return err
		}
	}

	if p := u.Query().Get("plugin"); p != "" {
		parts := strings.SplitN(p, ";", 2)
		config.Plugin = parts[0]
		if len(parts) > 1 {
			config.PluginOpts = parts[1]
		}
	}

	return nil
}

func decodeShadowsocksUserinfo(userinfo string, config *models.ShadowsocksConfig) error {
	decoded, err := base64Decode(userinfo)
	if err != nil {
		decoded = userinfo
	}
	parts := strings.SplitN(decoded, ":", 2)
	if len(parts) != 2 {
		return errors.New("invalid ss userinfo, expected method:password")
	}
	config.Method = parts[0]
	config.Password = parts[1]
	return nil
}

func base64Decode(s string) (string, error) {
	s = strings.ReplaceAll(s, "-", "+")
	s = strings.ReplaceAll(s, "_", "/")
	if pad := len(s) % 4; pad != 0 {
		s += strings.Repeat("=", 4-pad)
	}
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

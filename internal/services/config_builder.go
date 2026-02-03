package services

import (
	"errors"
	"gibraltar/internal/models"
	"net"
	"net/url"
	"strconv"
	"strings"
)

func parseVlessURL(config *models.VlessConfig) error {
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

func buildVlessOutbound(v *models.VlessConfig) map[string]any {
	outbound := map[string]any{
		"type":        "vless",
		"tag":         "proxy",
		"server":      v.Server,
		"server_port": v.Port,
		"uuid":        v.UUID,
	}

	if v.Flow != "" {
		outbound["flow"] = v.Flow
	}

	outbound["network"] = mapNetwork(v.Type)

	switch strings.ToLower(v.Security) {
	case "reality":
		outbound["tls"] = map[string]any{
			"enabled":     true,
			"server_name": v.SNI,
			"utls": map[string]any{
				"enabled": true,
			},
			"reality": map[string]any{
				"enabled":    true,
				"public_key": v.PublicKey,
				"short_id":   v.SID,
			},
		}
	case "tls", "ssl":
		outbound["tls"] = map[string]any{
			"enabled":     true,
			"server_name": v.SNI,
		}
	}

	return outbound
}

func mapNetwork(t string) string {
	switch strings.ToLower(t) {
	case "", "tcp", "raw":
		return "tcp"
	case "udp", "quic":
		return "udp"
	default:
		return "tcp"
	}
}

func buildTransport(v *models.VlessConfig) map[string]any {
	switch strings.ToLower(v.Type) {
	case "ws":
		ws := map[string]any{"type": "ws"}
		if v.Path != "" {
			ws["path"] = v.Path
		}
		if v.Host != "" {
			ws["headers"] = map[string]any{"Host": v.Host}
		} else if v.SNI != "" {
			ws["headers"] = map[string]any{"Host": v.SNI}
		}
		return ws
	case "grpc":
		grpc := map[string]any{"type": "grpc"}
		if v.ServiceName != "" {
			grpc["service_name"] = v.ServiceName
		}
		return grpc
	case "http", "h2", "http2", "xhttp":
		httpT := map[string]any{"type": "http"}
		if v.Path != "" {
			httpT["path"] = v.Path
		}
		if v.Host != "" {
			httpT["headers"] = map[string]any{"Host": v.Host}
		} else if v.SNI != "" {
			httpT["headers"] = map[string]any{"Host": v.SNI}
		}
		return httpT
	default:
		return nil
	}
}

func buildSingBoxConfig(outbound map[string]any, port int) map[string]any {
	tag := "proxy"
	if t, ok := outbound["tag"].(string); ok && t != "" {
		tag = t
	}
	return map[string]any{
		"log": map[string]any{"level": "error"},
		"inbounds": []any{
			map[string]any{
				"type":        "socks",
				"tag":         "socks-in",
				"listen":      "127.0.0.1",
				"listen_port": port,
			},
		},
		"outbounds": []any{
			outbound,
			map[string]any{"type": "direct", "tag": "direct"},
		},
		"route": map[string]any{
			"rules": []any{
				map[string]any{"outbound": tag},
			},
		},
	}
}

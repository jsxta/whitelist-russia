package beans

import (
	"gibraltar/internal/models"
	"strings"
)

type VlessBuilder struct {
	*models.VlessConfig
}

func (b *VlessBuilder) BuildOutbound() map[string]any {
	outbound := map[string]any{
		"type":        "vless",
		"tag":         "proxy",
		"server":      b.Server,
		"server_port": b.Port,
		"uuid":        b.UUID,
	}

	if b.Flow != "" {
		outbound["flow"] = b.Flow
	}

	outbound["network"] = mapNetwork(b.Type)

	switch strings.ToLower(b.Security) {
	case "reality":
		outbound["tls"] = map[string]any{
			"enabled":     true,
			"server_name": b.SNI,
			"utls": map[string]any{
				"enabled": true,
			},
			"reality": map[string]any{
				"enabled":    true,
				"public_key": b.PublicKey,
				"short_id":   b.SID,
			},
		}
	case "tls", "ssl":
		outbound["tls"] = map[string]any{
			"enabled":     true,
			"server_name": b.SNI,
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

func buildTransport(netType, path, host, serviceName string) map[string]any {
	switch strings.ToLower(netType) {
	case "ws":
		t := map[string]any{"type": "ws"}
		if path != "" {
			t["path"] = path
		}
		if host != "" {
			t["headers"] = map[string]any{"Host": host}
		}
		return t
	case "grpc":
		t := map[string]any{"type": "grpc"}
		if serviceName != "" {
			t["service_name"] = serviceName
		}
		return t
	case "http":
		t := map[string]any{"type": "http"}
		if path != "" {
			t["path"] = []string{path}
		}
		if host != "" {
			t["host"] = []string{host}
		}
		return t
	case "httpupgrade":
		t := map[string]any{"type": "httpupgrade"}
		if path != "" {
			t["path"] = path
		}
		if host != "" {
			t["host"] = host
		}
		return t
	}
	return nil
}

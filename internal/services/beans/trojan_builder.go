package beans

import (
	"gibraltar/internal/models"
	"strings"
)

type TrojanBuilder struct {
	*models.TrojanConfig
}

func (b *TrojanBuilder) BuildOutbound() map[string]any {
	outbound := map[string]any{
		"type":        "trojan",
		"tag":         "proxy",
		"server":      b.Server,
		"server_port": b.Port,
		"password":    b.Password,
	}

	outbound["network"] = mapNetwork(b.Type)

	if t := buildTransport(b.Type, b.Path, b.Host, b.ServiceName); t != nil {
		outbound["transport"] = t
	}

	switch strings.ToLower(b.Security) {
	case "reality":
		outbound["tls"] = map[string]any{
			"enabled":     true,
			"server_name": b.SNI,
			"utls": map[string]any{
				"enabled":     true,
				"fingerprint": b.Fingerprint,
			},
			"reality": map[string]any{
				"enabled":    true,
				"public_key": b.PublicKey,
				"short_id":   b.SID,
			},
		}
	case "tls", "ssl", "":
		tls := map[string]any{
			"enabled":     true,
			"server_name": b.SNI,
		}
		if b.Fingerprint != "" {
			tls["utls"] = map[string]any{
				"enabled":     true,
				"fingerprint": b.Fingerprint,
			}
		}
		outbound["tls"] = tls
	}

	return outbound
}

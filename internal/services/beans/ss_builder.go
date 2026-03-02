package beans

import "gibraltar/internal/models"

type ShadowsocksBuilder struct {
	*models.ShadowsocksConfig
}

func (b *ShadowsocksBuilder) BuildOutbound() map[string]any {
	outbound := map[string]any{
		"type":        "shadowsocks",
		"tag":         "proxy",
		"server":      b.Server,
		"server_port": b.Port,
		"method":      b.Method,
		"password":    b.Password,
	}
	if b.Plugin != "" {
		outbound["plugin"] = b.Plugin
		if b.PluginOpts != "" {
			outbound["plugin_opts"] = b.PluginOpts
		}
	}
	return outbound
}

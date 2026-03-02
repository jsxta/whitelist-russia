package beans

func BuildSingBoxConfig(outbound map[string]any, port int) map[string]any {
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

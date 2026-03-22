package config

import (
	"os"
	"path/filepath"
	"time"
)

// Stability calculation parameters
const (
	Gain    = 1.5
	Decay   = 0.7
	P       = 1.5
	Q       = 1.3
	MinDrop = 2
	MAX     = 100
)

// Stability criteria
const (
	MinValueForAccept float64 = 5.0
	MinValueForStable float64 = 60.0
)

// Cache Keys
const (
	AvailableKey = "latestResults"
	AllKey       = "currentConfigs"
)

// Directory and files
var (
	ConfigSourceDirectoryPath = filepath.Join(os.Getenv("HOME"), "rjsxrd")
	basePath                  = os.Getenv("APP_DATA_DIR")
	VlessSecureConfigsFile    = filepath.Join(ConfigSourceDirectoryPath, "githubmirror", "bypass", "bypass-all.txt")
	CIDRWhitelistFile         = filepath.Join(basePath, "source", "config", "cidrwhitelist.txt")
	URLsWhitelistFile         = filepath.Join(ConfigSourceDirectoryPath, "source", "config", "whitelist-all.txt")
	VlessSecureConfigsURLs    = []string{"https://raw.githubusercontent.com/whoahaow/rjsxrd/refs/heads/main/githubmirror/bypass/bypass-all.txt", "https://raw.githubusercontent.com/igareck/vpn-configs-for-russia/refs/heads/main/WHITE-CIDR-RU-all.txt", "https://raw.githubusercontent.com/FLEXIY0/matryoshka-vpn/refs/heads/main/configs/russia_whitelist.txt", "https://raw.githubusercontent.com/Epodonios/v2ray-configs/refs/heads/main/All_Configs_Sub.txt", "https://raw.githubusercontent.com/ShatakVPN/ConfigForge-V2Ray/refs/heads/main/configs/ru/vless.txt", "https://raw.githubusercontent.com/Argh94/V2RayAutoConfig/refs/heads/main/configs/Vless.txt", "https://raw.githubusercontent.com/zieng2/wl/refs/heads/main/vless_universal.txt", "https://raw.githubusercontent.com/y9felix/s/refs/heads/main/r", "https://raw.githubusercontent.com/kort0881/vpn-checker-backend/main/checked/RU_Best/ru_white_part1.txt", "https://raw.githubusercontent.com/kort0881/vpn-checker-backend/main/checked/RU_Best/ru_white_part2.txt", "https://raw.githubusercontent.com/kort0881/vpn-checker-backend/main/checked/RU_Best/ru_white_part3.txt", "https://raw.githubusercontent.com/kort0881/vpn-checker-backend/main/checked/RU_Best/ru_white_part4.txt", "https://raw.githubusercontent.com/kort0881/vpn-checker-backend/main/checked/RU_Best/ru_white_part5.txt", "https://raw.githubusercontent.com/ginolrewadsb11/studious-umbrella/main/bobi_vpn.txt", "https://raw.githubusercontent.com/sakha1370/OpenRay/fd98dbbea14ddd5912a93481659caaba565e45d4/output/country/RU.txt", "https://raw.githubusercontent.com/KiryaScript/white-lists/cf8bd3a525d1409539e60cae5430f82b58661f31/githubmirror/26.txt", "https://raw.githubusercontent.com/KiryaScript/white-lists/cf8bd3a525d1409539e60cae5430f82b58661f31/githubmirror/27.txt", "https://raw.githubusercontent.com/KiryaScript/white-lists/cf8bd3a525d1409539e60cae5430f82b58661f31/githubmirror/28.txt", "https://subrostunnel.vercel.app/wl.txt"}
	CIDRWhitelistURL          = "https://raw.githubusercontent.com/whoahaow/rjsxrd/refs/heads/main/source/config/cidrwhitelist.txt"
	URLsWhitelistURL          = "https://raw.githubusercontent.com/whoahaow/rjsxrd/refs/heads/main/source/config/whitelist-all.txt"
)

// Remote git repo
const (
	RemoteRepository = "https://github.com/whoahaow/rjsxrd.git"
	RemoteBranch     = "main"
)

// Sing-box workers
const (
	WorkersCount = 60
)

// Test server (URL)
const (
	TestURL = "http://cp.cloudflare.com/"
)

// Timings
const (
	UpdateCooldown = time.Hour
)

// Header tags
const (
	Tags = "#profile-web-page-url: https://github.com/jsxta/whitelist-russia\n#support-url: https://t.me/whitelistsupport_bot\n"
)

// Request Headers
const (
	UserAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:148.0) Gecko/20100101 Firefox/148.0"
)

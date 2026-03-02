package models

type BaseConfig struct {
	URL         string
	Server      string
	Port        int
	Security    string
	SNI         string
	Fingerprint string
	PublicKey   string
	SID         string
	Type        string
	Path        string
	Host        string
	ServiceName string
	TestResult  int
	Stability   float64
}

func (b *BaseConfig) GetServer() string      { return b.Server }
func (b *BaseConfig) GetPort() int           { return b.Port }
func (b *BaseConfig) GetURL() string         { return b.URL }
func (b *BaseConfig) GetTestResult() int     { return b.TestResult }
func (b *BaseConfig) GetStability() float64  { return b.Stability }
func (b *BaseConfig) GetType() string        { return b.Type }
func (b *BaseConfig) GetSNI() string         { return b.SNI }
func (b *BaseConfig) SetTestResult(v int)    { b.TestResult = v }
func (b *BaseConfig) SetStability(v float64) { b.Stability = v }
func (b *BaseConfig) SetURL(v string)        { b.URL = v }

type VlessConfig struct {
	BaseConfig
	UUID       string
	SPX        string
	Flow       string
	HeaderType string
}

type TrojanConfig struct {
	BaseConfig
	Password string
}

type ShadowsocksConfig struct {
	BaseConfig
	Password   string
	Method     string
	Plugin     string
	PluginOpts string
}

func (s *ShadowsocksConfig) GetPassword() string { return s.Password }
func (s *ShadowsocksConfig) GetMethod() string   { return s.Method }

type AnyConfig interface {
	GetServer() string
	GetPort() int
	GetURL() string
	GetType() string
	GetSNI() string
	GetStability() float64
	GetTestResult() int
	SetTestResult(v int)
	SetStability(v float64)
	SetURL(v string)
}

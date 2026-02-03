package models

type VlessConfig struct {
	URL string

	UUID        string
	Server      string
	Port        int
	Security    string
	SNI         string
	Fingerprint string
	PublicKey   string
	SID         string
	SPX         string
	Type        string
	Flow        string

	Path        string
	Host        string
	ServiceName string
	HeaderType  string

	TestResult int
	Stability  float64
}

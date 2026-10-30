package config

import (
	"os"
	"path"
	"rustdesk-api-server-pro/util"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	DebugMode   bool               `yaml:"debugMode"`
	Db          *DbConfig          `yaml:"db"`
	SignKey     string             `yaml:"signKey"`
	HttpConfig  *HttpConfig        `yaml:"httpConfig"`
	SmtpConfig  *SmtpConfig        `yaml:"smtpConfig"`
	JobsConfig  *JobsConfig        `yaml:"jobsConfig"`
	WebClient   *WebClientConfig   `yaml:"webClient"`
	Ldap        *LdapConfig        `yaml:"ldap"`
	RustdeskWeb *RustdeskWebConfig `yaml:"rustdeskWebConfig"`
}

// RustdeskWebConfig is the config of the embedded RustDesk web client,
// it is served to the client as /web/rustdeskWeb.json.
type RustdeskWebConfig struct {
	// Server is the RustDesk rendezvous server, e.g. "hbbs.example.com:21116".
	Server string `yaml:"server" json:"server"`
	// Relay is the RustDesk relay server, e.g. "hbbr.example.com:21115". Leave it empty to use the rendezvous server.
	Relay string `yaml:"relay" json:"relay"`
	// Key is the base64 encoded public key from the id_ed25519.pub of the RustDesk server.
	Key string `yaml:"key" json:"key"`
}

type WebClientConfig struct {
	Url string `yaml:"url" json:"url"`
}

type LdapConfig struct {
	Enabled            bool   `yaml:"enabled" json:"enabled"`
	Host               string `yaml:"host" json:"host"`
	Port               int    `yaml:"port" json:"port"`
	UseTLS             bool   `yaml:"use_tls" json:"use_tls"`
	StartTLS           bool   `yaml:"start_tls" json:"start_tls"`
	BindDN             string `yaml:"bind_dn" json:"bind_dn"`
	BindPassword       string `yaml:"bind_password" json:"bind_password"`
	BaseDN             string `yaml:"base_dn" json:"base_dn"`
	UserFilter         string `yaml:"user_filter" json:"user_filter"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify" json:"insecure_skip_verify"`
	// ActiveDirectory marks the directory as a Microsoft AD. When the bind DN is
	// not a full distinguished name it is bound as an AD account (UPN or DOMAIN\\user).
	ActiveDirectory bool `yaml:"active_directory" json:"active_directory"`
}

type DbConfig struct {
	Driver   string `yaml:"driver"`
	Dsn      string `yaml:"dsn"`
	TimeZone string `yaml:"timeZone"`
	ShowSql  bool   `yaml:"showSql"`
}

type HttpConfig struct {
	PrintRequestLog bool   `yaml:"printRequestLog"`
	Port            string `yaml:"port"`
	StaticDir       string `yaml:"staticdir"`
}

type SmtpConfig struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	Encryption string `yaml:"encryption"` // none ssl/tls starttls
	From       string `yaml:"from"`
}

type DeviceCheckJob struct {
	Duration int `yaml:"duration"`
}

type JobsConfig struct {
	DeviceCheckJob *DeviceCheckJob `yaml:"deviceCheckJob"`
}

var (
	wd, _    = os.Getwd()
	yamlFile = path.Join(wd, "server.yaml")
)

// YamlFile is the configuration file the defaults of the settings page come from.
func YamlFile() string {
	return yamlFile
}

func GetDefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		DebugMode: false,
		Db: &DbConfig{
			Driver:   "sqlite",
			Dsn:      "./server.db",
			ShowSql:  true,
			TimeZone: "Europe/Rome",
		},
		HttpConfig: &HttpConfig{
			Port:      ":8080",
			StaticDir: "dist",
		},
		SignKey: util.RandomString(32),
		WebClient: &WebClientConfig{
			// the embedded web client, see RustdeskWebConfig for its server/key
			Url: "/web/",
		},
		Ldap: &LdapConfig{
			Enabled:            false,
			Port:               389,
			StartTLS:           true,
			UserFilter:         "(sAMAccountName=%s)",
			InsecureSkipVerify: false,
		},
		RustdeskWeb: &RustdeskWebConfig{},
		JobsConfig: &JobsConfig{
			DeviceCheckJob: &DeviceCheckJob{
				Duration: 30,
			},
		},
	}
}

func GetServerConfig() *ServerConfig {
	cfg := GetDefaultServerConfig()
	bytes, err := os.ReadFile(yamlFile)
	if err != nil {
		WriteServerConfig(cfg)
		return cfg
	}

	err = yaml.Unmarshal(bytes, cfg)
	if err != nil {
		WriteServerConfig(cfg)
		return cfg
	}
	return cfg
}

func WriteServerConfig(cfg *ServerConfig) {
	bytes, _ := yaml.Marshal(cfg)
	_ = os.WriteFile(yamlFile, bytes, 0755)
}

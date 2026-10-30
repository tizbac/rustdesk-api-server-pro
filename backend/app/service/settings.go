package service

import (
	"fmt"
	"rustdesk-api-server-pro/app/model"
	"rustdesk-api-server-pro/config"
	"rustdesk-api-server-pro/util"
	"strconv"

	"xorm.io/xorm"
)

// Keys of the rows stored in the system_settings table. They override the
// server.yaml defaults, so the yaml file only provides the initial values and
// the admin panel is the source of truth once something is configured there.
const (
	KeyRustdeskWebServer   = "rustdesk_web.server"
	KeyRustdeskWebRelay    = "rustdesk_web.relay"
	KeyRustdeskWebKey      = "rustdesk_web.key"
	KeyWebClientUrl        = "web_client.url"
	KeyTimeZone            = "timezone"
	KeyLdapEnabled         = "ldap.enabled"
	KeyLdapHost            = "ldap.host"
	KeyLdapPort            = "ldap.port"
	KeyLdapUseTLS          = "ldap.use_tls"
	KeyLdapStartTLS        = "ldap.start_tls"
	KeyLdapBindDN          = "ldap.bind_dn"
	KeyLdapBindPassword    = "ldap.bind_password"
	KeyLdapBaseDN          = "ldap.base_dn"
	KeyLdapUserFilter      = "ldap.user_filter"
	KeyLdapInsecureTLS     = "ldap.insecure_skip_verify"
	KeyLdapActiveDirectory = "ldap.active_directory"
)

// RuntimeSettings is server.yaml merged with everything saved from the panel.
type RuntimeSettings struct {
	RustdeskWeb config.RustdeskWebConfig `json:"rustdeskWeb"`
	WebClient   config.WebClientConfig   `json:"webClient"`
	TimeZone    string                   `json:"timezone"`
	Ldap        config.LdapConfig        `json:"ldap"`
	// Overrides are the keys saved from the panel, the rest still come from yaml.
	Overrides []string `json:"overrides"`
}

func yamlSettings() RuntimeSettings {
	cfg := config.GetServerConfig()

	s := RuntimeSettings{TimeZone: "Europe/Rome"}

	if cfg.RustdeskWeb != nil {
		s.RustdeskWeb = *cfg.RustdeskWeb
	}
	if cfg.WebClient != nil {
		s.WebClient = *cfg.WebClient
	}
	if cfg.Db != nil && cfg.Db.TimeZone != "" {
		s.TimeZone = cfg.Db.TimeZone
	}
	if cfg.Ldap != nil {
		s.Ldap = *cfg.Ldap
	}

	return s
}

// RuntimeSettingsFromYaml returns the server.yaml defaults, they are shown in
// the panel as the values the settings page started from.
func RuntimeSettingsFromYaml() RuntimeSettings {
	return yamlSettings()
}

// GetRuntimeSettings returns the yaml defaults overlaid with the values saved
// in the database by the settings page.
func GetRuntimeSettings(engine *xorm.Engine) (RuntimeSettings, error) {
	settings := yamlSettings()

	if engine == nil {
		return settings, nil
	}

	rows := []model.SystemSettings{}
	if err := engine.Find(&rows); err != nil {
		return settings, fmt.Errorf("failed to load settings: %w", err)
	}

	for _, row := range rows {
		switch row.Key {
		case KeyRustdeskWebServer:
			settings.RustdeskWeb.Server = row.Value
		case KeyRustdeskWebRelay:
			settings.RustdeskWeb.Relay = row.Value
		case KeyRustdeskWebKey:
			settings.RustdeskWeb.Key = row.Value
		case KeyWebClientUrl:
			settings.WebClient.Url = row.Value
		case KeyTimeZone:
			if row.Value != "" {
				settings.TimeZone = row.Value
			}
		case KeyLdapEnabled:
			settings.Ldap.Enabled = parseBool(row.Value)
		case KeyLdapHost:
			settings.Ldap.Host = row.Value
		case KeyLdapPort:
			if port, err := strconv.Atoi(row.Value); err == nil {
				settings.Ldap.Port = port
			}
		case KeyLdapUseTLS:
			settings.Ldap.UseTLS = parseBool(row.Value)
		case KeyLdapStartTLS:
			settings.Ldap.StartTLS = parseBool(row.Value)
		case KeyLdapBindDN:
			settings.Ldap.BindDN = row.Value
		case KeyLdapBindPassword:
			settings.Ldap.BindPassword = row.Value
		case KeyLdapBaseDN:
			settings.Ldap.BaseDN = row.Value
		case KeyLdapUserFilter:
			settings.Ldap.UserFilter = row.Value
		case KeyLdapInsecureTLS:
			settings.Ldap.InsecureSkipVerify = parseBool(row.Value)
		case KeyLdapActiveDirectory:
			settings.Ldap.ActiveDirectory = parseBool(row.Value)
		default:
			continue
		}

		settings.Overrides = append(settings.Overrides, row.Key)
	}

	return settings, nil
}

// SaveRuntimeSettings stores every value in the database, so the panel becomes
// the source of truth and server.yaml only keeps the defaults.
func SaveRuntimeSettings(engine *xorm.Engine, settings RuntimeSettings) error {
	if engine == nil {
		return fmt.Errorf("database is not available")
	}

	values := map[string]string{
		KeyRustdeskWebServer:   settings.RustdeskWeb.Server,
		KeyRustdeskWebRelay:    settings.RustdeskWeb.Relay,
		KeyRustdeskWebKey:      settings.RustdeskWeb.Key,
		KeyWebClientUrl:        settings.WebClient.Url,
		KeyTimeZone:            settings.TimeZone,
		KeyLdapEnabled:         strconv.FormatBool(settings.Ldap.Enabled),
		KeyLdapHost:            settings.Ldap.Host,
		KeyLdapPort:            strconv.Itoa(settings.Ldap.Port),
		KeyLdapUseTLS:          strconv.FormatBool(settings.Ldap.UseTLS),
		KeyLdapStartTLS:        strconv.FormatBool(settings.Ldap.StartTLS),
		KeyLdapBindDN:          settings.Ldap.BindDN,
		KeyLdapBindPassword:    settings.Ldap.BindPassword,
		KeyLdapBaseDN:          settings.Ldap.BaseDN,
		KeyLdapUserFilter:      settings.Ldap.UserFilter,
		KeyLdapInsecureTLS:     strconv.FormatBool(settings.Ldap.InsecureSkipVerify),
		KeyLdapActiveDirectory: strconv.FormatBool(settings.Ldap.ActiveDirectory),
	}

	for key, value := range values {
		row := model.SystemSettings{Key: key}
		has, err := engine.Get(&row)
		if err != nil {
			return fmt.Errorf("failed to read setting %q: %w", key, err)
		}

		if has {
			if _, err := engine.Update(model.SystemSettings{Value: value}, model.SystemSettings{Key: key}); err != nil {
				return fmt.Errorf("failed to save setting %q: %w", key, err)
			}
			continue
		}

		if _, err := engine.Insert(model.SystemSettings{Key: key, Value: value}); err != nil {
			return fmt.Errorf("failed to create setting %q: %w", key, err)
		}
	}

	return nil
}

// ResetRuntimeSettings deletes the saved values and goes back to server.yaml.
func ResetRuntimeSettings(engine *xorm.Engine) error {
	if engine == nil {
		return fmt.Errorf("database is not available")
	}

	rows := []model.SystemSettings{}
	if err := engine.Find(&rows); err != nil {
		return err
	}

	for _, row := range rows {
		if _, err := engine.Delete(model.SystemSettings{Id: row.Id}); err != nil {
			return err
		}
	}

	return nil
}

// RustdeskWebConfig is what the embedded web client reads from /web/rustdeskWeb.json.
func RustdeskWebConfig(engine *xorm.Engine) (config.RustdeskWebConfig, error) {
	settings, err := GetRuntimeSettings(engine)
	if err != nil {
		return config.RustdeskWebConfig{}, err
	}
	return settings.RustdeskWeb, nil
}

// WebClientAvailable reports whether the RustDesk web client files are on disk.
func WebClientAvailable() bool {
	return util.FileExists(util.AssetDir("web") + "/index.html")
}

func parseBool(value string) bool {
	switch value {
	case "1", "true", "TRUE", "True", "yes", "on":
		return true
	default:
		return false
	}
}

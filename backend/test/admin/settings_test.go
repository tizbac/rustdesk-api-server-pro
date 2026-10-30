package admin_test

import (
	"path/filepath"
	"rustdesk-api-server-pro/app/model"
	"rustdesk-api-server-pro/app/service"
	"rustdesk-api-server-pro/config"
	"rustdesk-api-server-pro/db"
	"testing"

	"xorm.io/xorm"
)

func testSettingsEngine(t *testing.T) *xorm.Engine {
	engine, err := xorm.NewEngine("sqlite", filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatalf("create engine: %v", err)
	}

	if err := engine.Sync(new(model.SystemSettings)); err != nil {
		t.Fatalf("sync system_settings: %v", err)
	}

	return engine
}

func TestRuntimeSettingsOverrideYaml(t *testing.T) {
	engine := testSettingsEngine(t)

	saved := service.RuntimeSettings{
		RustdeskWeb: config.RustdeskWebConfig{
			Server: "hbbs.test.local:21116",
			Relay:  "hbbr.test.local:21115",
			Key:    "pubkey",
		},
		WebClient: config.WebClientConfig{Url: "/remote/"},
		TimeZone:  "Europe/Paris",
		Ldap: config.LdapConfig{
			Enabled:         true,
			Host:            "dc01.test.local",
			Port:            636,
			UseTLS:          true,
			BindDN:          "CN=rustdesk,OU=Service,DC=test,DC=local",
			BindPassword:    "secret",
			BaseDN:          "DC=test,DC=local",
			UserFilter:      "(sAMAccountName=%s)",
			ActiveDirectory: true,
		},
	}

	if err := service.SaveRuntimeSettings(engine, saved); err != nil {
		t.Fatalf("save settings: %v", err)
	}

	merged, err := service.GetRuntimeSettings(engine)
	if err != nil {
		t.Fatalf("load settings: %v", err)
	}

	if merged.RustdeskWeb.Server != "hbbs.test.local:21116" || merged.RustdeskWeb.Relay != "hbbr.test.local:21115" {
		t.Fatalf("rustdesk web not merged: %+v", merged.RustdeskWeb)
	}

	if merged.WebClient.Url != "/remote/" || merged.TimeZone != "Europe/Paris" {
		t.Fatalf("web client or timezone not merged: %+v", merged)
	}

	if !merged.Ldap.Enabled || merged.Ldap.Port != 636 || !merged.Ldap.UseTLS || !merged.Ldap.ActiveDirectory {
		t.Fatalf("ldap not merged: %+v", merged.Ldap)
	}

	if len(merged.Overrides) != 16 {
		t.Fatalf("expected 16 overridden keys, got %d", len(merged.Overrides))
	}

	if err := service.ResetRuntimeSettings(engine); err != nil {
		t.Fatalf("reset settings: %v", err)
	}

	backToYaml, err := service.GetRuntimeSettings(engine)
	if err != nil {
		t.Fatalf("load settings after reset: %v", err)
	}

	if len(backToYaml.Overrides) != 0 || backToYaml.RustdeskWeb.Server != "" || backToYaml.TimeZone != "Europe/Rome" {
		t.Fatalf("settings did not go back to the yaml defaults: %+v", backToYaml)
	}
}

func TestSetTimeZone(t *testing.T) {
	db.DbEngine = testSettingsEngine(t)

	if err := db.SetTimeZone("Not/AZone"); err == nil {
		t.Fatal("an invalid timezone should not be accepted")
	}

	if err := db.SetTimeZone("Europe/Rome"); err != nil {
		t.Fatalf("valid timezone rejected: %v", err)
	}

	if db.DbEngine.DatabaseTZ.String() != "Europe/Rome" {
		t.Fatalf("engine timezone not applied: %s", db.DbEngine.DatabaseTZ.String())
	}
}

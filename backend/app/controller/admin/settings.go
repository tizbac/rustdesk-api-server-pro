package admin

import (
	"rustdesk-api-server-pro/app/model"
	"rustdesk-api-server-pro/app/service"
	"rustdesk-api-server-pro/config"
	"rustdesk-api-server-pro/db"
	"strings"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"xorm.io/xorm"
)

type SettingsController struct {
	basicController
}

func (c *SettingsController) BeforeActivation(b mvc.BeforeActivation) {
	b.Handle("GET", "/settings", "HandleSettings")
	b.Handle("POST", "/settings", "HandleSaveSettings")
	b.Handle("POST", "/settings/reset", "HandleResetSettings")
	b.Handle("POST", "/settings/ldap-test", "HandleLdapTest")
}

func (c *SettingsController) HandleSettings() mvc.Result {
	settings, err := service.GetRuntimeSettings(c.Db)
	if err != nil {
		return c.Error(err.Error(), "failed to load settings")
	}

	return c.Success(iris.Map{
		"settings":   settings,
		"defaults":   service.RuntimeSettingsFromYaml(),
		"timezones":  availableTimeZones(),
		"webClient":  iris.Map{"available": service.WebClientAvailable()},
		"yamlFile":   config.YamlFile(),
		"savedAt":    settingsSavedAt(c.Db),
		"serverTime": time.Now().Format(config.TimeFormat),
	}, "ok")
}

func (c *SettingsController) HandleSaveSettings() mvc.Result {
	var settings service.RuntimeSettings
	if err := c.Ctx.ReadJSON(&settings); err != nil {
		return c.Error(err.Error(), "invalid request body")
	}

	if strings.TrimSpace(settings.TimeZone) == "" {
		return c.Error("", "timezone is required")
	}

	if _, err := time.LoadLocation(settings.TimeZone); err != nil {
		return c.Error(err.Error(), "invalid timezone")
	}

	if err := service.SaveRuntimeSettings(c.Db, settings); err != nil {
		return c.Error(err.Error(), "failed to save settings")
	}

	if err := db.SetTimeZone(settings.TimeZone); err != nil {
		return c.Error(err.Error(), "settings saved but the timezone is not valid")
	}

	return c.Success(settings, "settings saved")
}

// HandleResetSettings removes everything saved in the panel and goes back to
// the values of server.yaml.
func (c *SettingsController) HandleResetSettings() mvc.Result {
	if err := service.ResetRuntimeSettings(c.Db); err != nil {
		return c.Error(err.Error(), "failed to reset settings")
	}

	defaults := service.RuntimeSettingsFromYaml()
	if err := db.SetTimeZone(defaults.TimeZone); err != nil {
		return c.Error(err.Error(), "settings reset but the default timezone is not valid")
	}

	return c.Success(defaults, "settings reset to the server.yaml defaults")
}

// HandleLdapTest connects with the values of the form, they do not have to be
// saved yet, and reports every step that succeeded or failed.
func (c *SettingsController) HandleLdapTest() mvc.Result {
	request := service.LdapTestRequest{}
	if err := c.Ctx.ReadJSON(&request); err != nil {
		return c.Error(err.Error(), "invalid request body")
	}

	return c.Success(service.TestLdap(request), "ok")
}

func availableTimeZones() []string {
	zones := []string{
		"Europe/Rome", "Europe/London", "Europe/Berlin", "Europe/Paris",
		"Europe/Madrid", "Europe/Warsaw", "Europe/Zurich", "Europe/Vienna",
		"Europe/Athens", "UTC", "America/New_York", "America/Los_Angeles",
		"America/Sao_Paulo", "Asia/Tokyo", "Asia/Shanghai", "Asia/Kolkata",
		"Australia/Sydney",
	}

	locations := []string{}
	for _, zone := range zones {
		if _, err := time.LoadLocation(zone); err == nil {
			locations = append(locations, zone)
		}
	}

	return locations
}

func settingsSavedAt(engine *xorm.Engine) string {
	if engine == nil {
		return ""
	}

	row := model.SystemSettings{}
	if ok, err := engine.Get(&row); err != nil || !ok {
		return ""
	}

	return row.UpdatedAt.Format(config.TimeFormat)
}

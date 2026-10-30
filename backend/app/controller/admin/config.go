package admin

import (
	"rustdesk-api-server-pro/app/service"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
)

type ConfigController struct {
	basicController
}

func (c *ConfigController) BeforeActivation(b mvc.BeforeActivation) {
	b.Handle("GET", "/config", "HandleConfig")
}

func (c *ConfigController) HandleConfig() mvc.Result {
	settings, err := service.GetRuntimeSettings(c.Db)
	if err != nil {
		return c.Error(err.Error(), "failed to load settings")
	}

	return c.Success(iris.Map{
		"webClientUrl": settings.WebClient.Url,
		"rustdeskWeb":  settings.RustdeskWeb,
		"timezone":     settings.TimeZone,
		"ldapEnabled":  settings.Ldap.Enabled,
	}, "ok")
}

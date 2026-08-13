package admin

import (
	"rustdesk-api-server-pro/config"

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
	cfg := config.GetServerConfig()
	
	return c.Success(iris.Map{
		"webClientUrl": cfg.WebClient.Url,
	}, "ok")
}
package app

import (
	"path/filepath"
	"rustdesk-api-server-pro/app/service"
	"rustdesk-api-server-pro/config"
	"rustdesk-api-server-pro/db"
	"rustdesk-api-server-pro/util"

	"github.com/kataras/iris/v12"
)

// webClientDir holds the RustDesk web client built by rustdesk/build_web.sh
// (make web or the Docker build), it is served as-is.
const webClientDir = "web"

// SetRustdeskWebRoute serves the embedded RustDesk web client under /web and
// feeds it the server/relay/key values saved in the admin panel.
//
// server.yaml only provides the defaults, the values configured in the panel
// (system_settings table) win, so the json is generated on every request.
//
// The client loads rustdeskWeb.json relative to its <base href="/web/">, so the
// json route has to be registered before the directory handler.
func SetRustdeskWebRoute(app *iris.Application, cfg *config.ServerConfig) {
	app.Get("/web/rustdeskWeb.json", func(ctx iris.Context) {
		webCfg, err := service.RustdeskWebConfig(db.DbEngine)
		if err != nil {
			app.Logger().Error("failed to load the rustdesk web config:", err)
			if cfg.RustdeskWeb != nil {
				webCfg = *cfg.RustdeskWeb
			}
		}

		ctx.Header("Cache-Control", "no-store")
		_ = ctx.JSON(iris.Map{
			"server": webCfg.Server,
			"relay":  webCfg.Relay,
			"key":    webCfg.Key,
		})
	})

	app.Get("/web/status", func(ctx iris.Context) {
		_ = ctx.JSON(iris.Map{
			"available": service.WebClientAvailable(),
		})
	})

	app.Get("/web", func(ctx iris.Context) {
		ctx.ServeFile(filepath.Join(util.AssetDir(webClientDir), "index.html"))
	})

	app.HandleDir("/web/", iris.Dir(util.AssetDir(webClientDir)))
}

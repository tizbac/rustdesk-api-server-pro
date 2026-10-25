package app

import (
	"path/filepath"
	"rustdesk-api-server-pro/config"

	"github.com/kataras/iris/v12"
)

// webClientDir holds the RustDesk web client built by rustdesk/build_web.sh
// (make web or the Docker build), it is served as-is.
const webClientDir = "web"

// SetRustdeskWebRoute serves the embedded RustDesk web client under /web and
// feeds it the server/relay/key values of the main config file.
//
// The client loads rustdeskWeb.json relative to its <base href="/web/">, so the
// json route has to be registered before the directory handler.
func SetRustdeskWebRoute(app *iris.Application, cfg *config.ServerConfig) {
	webCfg := cfg.RustdeskWeb
	if webCfg == nil {
		webCfg = &config.RustdeskWebConfig{}
	}

	app.Get("/web", func(ctx iris.Context) {
		ctx.ServeFile(filepath.Join(assetDir(webClientDir), "index.html"))
	})

	app.Get("/web/rustdeskWeb.json", func(ctx iris.Context) {
		ctx.Header("Cache-Control", "no-store")
		_ = ctx.JSON(iris.Map{
			"server": webCfg.Server,
			"relay":  webCfg.Relay,
			"key":    webCfg.Key,
		})
	})

	app.HandleDir("/web/", iris.Dir(assetDir(webClientDir)))
}

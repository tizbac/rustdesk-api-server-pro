package app

import (
	"fmt"
	"rustdesk-api-server-pro/app/middleware"
	"rustdesk-api-server-pro/app/service"
	"rustdesk-api-server-pro/config"
	"rustdesk-api-server-pro/db"
	"rustdesk-api-server-pro/util"

	"github.com/kataras/iris/v12"
)

func newApp(cfg *config.ServerConfig) (*iris.Application, error) {
	app := iris.Default()

	dbEngine, err := db.NewEngine(cfg.Db)
	if err != nil {
		app.Logger().Fatal("Db Engine create error:", err)
		return nil, err
	}
	app.RegisterDependency(dbEngine, cfg)

	// the panel overrides server.yaml, apply the saved timezone before serving
	settings, err := service.GetRuntimeSettings(dbEngine)
	if err != nil {
		app.Logger().Warn("could not load the saved settings:", err)
	} else if settings.TimeZone != "" {
		if err := db.SetTimeZone(settings.TimeZone); err != nil {
			app.Logger().Warn("could not apply the saved timezone:", err)
		}
	}

	app.OnErrorCode(iris.StatusNotFound, func(context iris.Context) {
		requestInfo := fmt.Sprintf("(404)▶ %s:%s", context.Method(), context.Request().RequestURI)
		body, _ := context.GetBody()
		context.Application().Logger().Info(requestInfo)
		for header, value := range context.Request().Header {
			fmt.Println(header+":", value)
		}
		fmt.Println(string(body))
	})

	app.Use(iris.Compression)
	if cfg.HttpConfig.PrintRequestLog {
		app.Use(middleware.RequestLogger())
	}

	SetRoute(app)

	SetRustdeskWebRoute(app, cfg)

	app.HandleDir("/", iris.Dir(util.AssetDir(cfg.HttpConfig.StaticDir)))

	return app, nil
}

func StartServer() (bool, error) {
	cfg := config.GetServerConfig()

	StartJobs(cfg)

	app, err := newApp(cfg)
	if err != nil {
		return false, err
	}

	err = app.Listen(cfg.HttpConfig.Port, iris.WithoutBodyConsumptionOnUnmarshal)
	if err != nil {
		return false, err
	}

	return true, nil
}

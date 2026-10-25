package app

import (
	"fmt"
	"os"
	"path/filepath"
	"rustdesk-api-server-pro/app/middleware"
	"rustdesk-api-server-pro/config"
	"rustdesk-api-server-pro/db"

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

	app.HandleDir("/", iris.Dir(assetDir(cfg.HttpConfig.StaticDir)))

	return app, nil
}

// assetDir prefers a directory shipped next to the binary, so the server keeps
// serving dist and web when it is started from another working directory
// (e.g. /app/data in the container), and falls back to the given path.
func assetDir(dir string) string {
	if dir == "" || filepath.IsAbs(dir) {
		return dir
	}
	if exe, err := os.Executable(); err == nil {
		next := filepath.Join(filepath.Dir(exe), dir)
		if _, err := os.Stat(next); err == nil {
			return next
		}
	}
	return dir
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

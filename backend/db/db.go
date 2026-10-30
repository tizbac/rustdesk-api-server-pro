package db

import (
	"fmt"
	"rustdesk-api-server-pro/config"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "modernc.org/sqlite"
	"xorm.io/xorm"
)

var (
	DbEngine *xorm.Engine
	engineMu sync.Mutex
)

func NewEngine(cfg *config.DbConfig) (*xorm.Engine, error) {
	if DbEngine != nil {
		return DbEngine, nil
	}

	engineMu.Lock()
	defer engineMu.Unlock()

	if DbEngine != nil {
		return DbEngine, nil
	}

	engine, err := xorm.NewEngine(cfg.Driver, cfg.Dsn)
	if err != nil {
		return nil, err
	}
	location, _ := time.LoadLocation(cfg.TimeZone)
	engine.TZLocation = location
	engine.DatabaseTZ = location
	engine.ShowSQL(cfg.ShowSql)
	engine.SetMaxIdleConns(25)
	engine.SetMaxOpenConns(25)
	DbEngine = engine
	return engine, nil
}

// SetTimeZone changes the timezone used to write and read datetime columns.
// It takes effect for the records created from now on, the rows already stored
// keep the time they were written with.
func SetTimeZone(name string) error {
	if DbEngine == nil {
		return fmt.Errorf("database engine is not initialized")
	}

	location, err := time.LoadLocation(name)
	if err != nil {
		return fmt.Errorf("invalid timezone %q: %w", name, err)
	}

	DbEngine.TZLocation = location
	DbEngine.DatabaseTZ = location
	return nil
}

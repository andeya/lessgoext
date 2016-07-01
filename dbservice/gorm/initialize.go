package gorm

import (
	"os"
	"path/filepath"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/lessgo/lessgo"
	"github.com/lessgo/lessgo/utils"
)

// 注册数据库服务
func initDBService() (dbService *DBService) {
	dbService = &DBService{
		List: map[string]*gorm.DB{},
	}

	defer func() {
		if dbService.Default == nil {
			time.Sleep(2e9)
		}
	}()

	err := dbServiceConfig.LoadDBConfig()
	if err != nil {
		lessgo.Log.Error(err.Error())
	}

	for _, conf := range dbServiceConfig.DBList {
		engine, err := gorm.Open(conf.Driver, conf.Connstring)
		if err != nil {
			lessgo.Log.Error("%v\n", err)
			continue
		}
		logger := newILogger(lessgo.Config.Log.AsyncChan, lessgo.Config.Log.Level, conf.Name)
		logger.BeeLogger.EnableFuncCallDepth(lessgo.Config.Debug)
		engine.SetLogger(logger)
		engine.LogMode(conf.ShowSql)

		engine.DB().SetMaxOpenConns(conf.MaxOpenConns)
		engine.DB().SetMaxIdleConns(conf.MaxIdleConns)

		if conf.Driver == "sqlite3" && !utils.FileExists(conf.Connstring) {
			os.MkdirAll(filepath.Dir(conf.Connstring), 0777)
			f, err := os.Create(conf.Connstring)
			if err != nil {
				lessgo.Log.Error("%v", err)
			} else {
				f.Close()
			}
		}

		dbService.List[conf.Name] = engine
		if dbServiceConfig.DefaultDB == conf.Name {
			dbService.Default = engine
		}
	}
	return
}

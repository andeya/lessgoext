package xorm

import (
	"os"
	"path/filepath"
	"time"

	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"

	"github.com/lessgo/lessgo"
	"github.com/lessgo/lessgo/utils"
)

// 注册数据库服务
func initDBService() (dbService *DBService) {
	dbService = &DBService{
		List: map[string]*xorm.Engine{},
	}

	defer func() {
		if dbService.Default == nil {
			time.Sleep(2e9)
		}
	}()

	err := dbServiceConfig.LoadDBConfig()
	if err != nil {
		lessgo.Log.Error(err.Error())
		return
	}

	for _, conf := range dbServiceConfig.DBList {
		engine, err := xorm.NewEngine(conf.Driver, conf.ConnString)
		if err != nil {
			lessgo.Log.Error("%v\n", err)
			continue
		}
		logger := newILogger(lessgo.Config.Log.AsyncChan, lessgo.Config.Log.Level, conf.Name)
		logger.BeeLogger.EnableFuncCallDepth(lessgo.Config.Debug)

		engine.SetLogger(logger)
		engine.SetMaxOpenConns(conf.MaxOpenConns)
		engine.SetMaxIdleConns(conf.MaxIdleConns)
		engine.SetDisableGlobalCache(conf.DisableCache)
		engine.ShowSQL(conf.ShowSql)
		engine.ShowExecTime(conf.ShowExecTime)

		if (conf.TableFix == "prefix" || conf.TableFix == "suffix") && len(conf.TableSpace) > 0 {
			var impr core.IMapper
			if conf.TableSnake {
				impr = core.SnakeMapper{}
			} else {
				impr = core.SameMapper{}
			}
			if conf.TableFix == "prefix" {
				engine.SetTableMapper(core.NewPrefixMapper(impr, conf.TableSpace))
			} else {
				engine.SetTableMapper(core.NewSuffixMapper(impr, conf.TableSpace))
			}
		}

		if (conf.ColumnFix == "prefix" || conf.ColumnFix == "suffix") && len(conf.ColumnSpace) > 0 {
			var impr core.IMapper
			if conf.ColumnSnake {
				impr = core.SnakeMapper{}
			} else {
				impr = core.SameMapper{}
			}
			if conf.ColumnFix == "prefix" {
				engine.SetTableMapper(core.NewPrefixMapper(impr, conf.ColumnSpace))
			} else {
				engine.SetTableMapper(core.NewSuffixMapper(impr, conf.ColumnSpace))
			}
		}

		if conf.Driver == "sqlite3" && !utils.FileExists(conf.ConnString) {
			os.MkdirAll(filepath.Dir(conf.ConnString), 0777)
			f, err := os.Create(conf.ConnString)
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

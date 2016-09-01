package sqlx

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"

	"github.com/lessgo/lessgo"
	"github.com/lessgo/lessgo/utils"
)

// 注册数据库服务
func initDBService() (dbService *DBService) {
	dbService = &DBService{
		List: map[string]*sqlx.DB{},
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
		db, err := sqlx.Connect(conf.Driver, conf.Connstring)
		if err != nil {
			lessgo.Log.Error("%v\n", err)
			continue
		}

		db.SetMaxOpenConns(conf.MaxOpenConns)
		db.SetMaxIdleConns(conf.MaxIdleConns)

		var strFunc = strings.ToLower
		if conf.ColumnSnake {
			strFunc = utils.SnakeString
		}
		if conf.StructTag == "" {
			conf.StructTag = "db"
		}

		// Create a new mapper which will use the struct field tag "json" instead of "db"
		db.Mapper = reflectx.NewMapperFunc(conf.StructTag, strFunc)

		if conf.Driver == "sqlite3" && !utils.FileExists(conf.Connstring) {
			os.MkdirAll(filepath.Dir(conf.Connstring), 0777)
			f, err := os.Create(conf.Connstring)
			if err != nil {
				lessgo.Log.Error("%v", err)
			} else {
				f.Close()
			}
		}

		dbService.List[conf.Name] = db
		if dbServiceConfig.DefaultDB == conf.Name {
			dbService.Default = db
		}
	}
	return
}

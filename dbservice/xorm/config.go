package xorm

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/henrylee2cn/lessgo"
	confpkg "github.com/henrylee2cn/lessgo/config"
)

type (
	config struct {
		DefaultDB string
		DBList    map[string]DBConfig
	}
	// DataBase connection Config
	DBConfig struct {
		Name         string
		Driver       string // Driver：mssql | odbc(mssql) | mysql | mymysql | postgres | sqlite3 | oci8 | goracle
		Connstring   string
		MaxOpenConns int
		MaxIdleConns int
		TableFix     string // 表命名空间是前缀还是后缀：prefix | suffix
		TableSpace   string // 表命名空间
		TableSnake   bool   // 表名使用snake风格或保持不变
		ColumnFix    string // 列命名空间是前缀还是后缀：prefix | suffix
		ColumnSpace  string // 列命名空间
		ColumnSnake  bool   // 列名使用snake风格或保持不变
		DisableCache bool
		ShowExecTime bool
		ShowSql      bool
	}
)

// 项目固定目录文件名称
const (
	DBCONFIG_FILE     = lessgo.CONFIG_DIR + "/xorm.config"
	DATABASE_DIR      = "database"
	DEFAULTDB_SECTION = "defaultdb"
	LOG_FOLDER        = "logger/"
)

var dbServiceConfig = func() *config {
	return &config{
		DefaultDB: "lessgo",
		DBList: map[string]DBConfig{
			"lessgo": {
				Name:         "lessgo",
				Driver:       "sqlite3",
				Connstring:   DATABASE_DIR + "/sqlite.db",
				MaxOpenConns: 0,
				MaxIdleConns: 0,
				TableFix:     "prefix",
				TableSpace:   "",
				TableSnake:   true,
				ColumnFix:    "prefix",
				ColumnSpace:  "",
				ColumnSnake:  true,
				DisableCache: false,
				ShowExecTime: false,
				ShowSql:      false,
			},
		},
	}
}()

func (this *config) LoadDBConfig() (err error) {
	fname := DBCONFIG_FILE
	iniconf, err := confpkg.NewConfig("ini", fname)
	if err == nil {
		os.Remove(fname)
		sections := iniconf.(*confpkg.IniConfigContainer).Sections()
		if len(sections) > 0 {
			this.DefaultDB = ""
			defDB := this.DBList["lessgo"]
			delete(this.DBList, "lessgo")
			for _, section := range sections {
				dbconfig := defDB
				lessgo.ReadSingleConfig(section, &dbconfig, iniconf)
				if strings.ToLower(section) == DEFAULTDB_SECTION {
					this.DefaultDB = dbconfig.Name
				}
				this.DBList[dbconfig.Name] = dbconfig
			}
			if this.DefaultDB == "" {
				this.DefaultDB = iniconf.DefaultString(sections[0]+"::name", defDB.Name)
			}
		}
	}

	os.MkdirAll(filepath.Dir(fname), 0777)
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	f.Close()
	iniconf, err = confpkg.NewConfig("ini", fname)
	if err != nil {
		return err
	}
	for _, dbconfig := range this.DBList {
		if this.DefaultDB == dbconfig.Name {
			lessgo.WriteSingleConfig(DEFAULTDB_SECTION, &dbconfig, iniconf)
		} else {
			lessgo.WriteSingleConfig(dbconfig.Name, &dbconfig, iniconf)
		}
	}

	return iniconf.SaveConfigFile(fname)
}

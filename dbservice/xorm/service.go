/**
 * 使用xorm数据库服务
 */
package xorm

import (
	"fmt"

	_ "github.com/denisenkom/go-mssqldb" //mssql
	_ "github.com/go-sql-driver/mysql"   //mysql
	_ "github.com/lib/pq"                //postgres
	// _ "github.com/mattn/go-sqlite3"      //sqlite
	// _ "github.com/mattn/go-oci8"    //oracle，需安装pkg-config工具

	"github.com/go-xorm/xorm"
)

/**
 * DBService 数据库服务
 */
type (
	DBService struct {
		Default *xorm.Engine
		List    map[string]*xorm.Engine
	}
)

var dbService = initDBService()

/**
 * 获取默认数据库引擎
 */
func DefaultDB() *xorm.Engine {
	return dbService.Default
}

/**
 * 获取指定数据库引擎
 */
func GetDB(name string) (*xorm.Engine, bool) {
	engine, ok := dbService.List[name]
	return engine, ok
}

/**
 * 获取全部数据库引擎列表
 */
func DBList() map[string]*xorm.Engine {
	return dbService.List
}

/**
 * 设置默认数据库引擎
 */
func SetDefault(name string) error {
	engine, ok := dbService.List[name]
	if !ok {
		return fmt.Errorf("Specified database does not exist: %v.", name)
	}
	dbService.Default = engine
	return nil
}

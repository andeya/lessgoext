/**
 * 使用sqlx数据库服务
 */
package sqlx

import (
	"fmt"

	_ "github.com/denisenkom/go-mssqldb" //mssql
	_ "github.com/go-sql-driver/mysql"   //mysql
	_ "github.com/lib/pq"                //postgres
	// _ "github.com/mattn/go-sqlite3"      //sqlite
	// _ "github.com/mattn/go-oci8"    //oracle，需安装pkg-config工具

	"github.com/jmoiron/sqlx"
)

/**
 * DBService 数据库服务
 */
type DBService struct {
	Default *sqlx.DB            //默认数据库引擎
	List    map[string]*sqlx.DB //数据库引擎列表
}

var dbService = initDBService()

/**
 * 获取默认数据库引擎
 */
func DefaultDB() *sqlx.DB {
	return dbService.Default
}

/**
 * 获取指定数据库引擎
 */
func GetDB(name string) (*sqlx.DB, bool) {
	db, ok := dbService.List[name]
	return db, ok
}

/**
 * 获取全部数据库引擎列表
 */
func DBList() map[string]*sqlx.DB {
	return dbService.List
}

/**
 * 获取默认数据库连接字符串
 */
func DefaultConnstring() string {
	return dbServiceConfig.DBList[dbServiceConfig.DefaultDB].Connstring
}

/**
 * 设置默认数据库引擎
 */
func SetDefault(name string) error {
	db, ok := dbService.List[name]
	if !ok {
		return fmt.Errorf("Specified database does not exist: %v.", name)
	}
	dbService.Default = db
	return nil
}

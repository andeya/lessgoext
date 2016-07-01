package gorm

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql" //github.com/go-sql-driver/mysql
	_ "github.com/jinzhu/gorm/dialects/postgres"
	// _ "github.com/jinzhu/gorm/dialects/sqlite" //github.com/mattn/go-sqlite3
)

type DBService struct {
	Default *gorm.DB            //默认数据库引擎
	List    map[string]*gorm.DB //数据库引擎列表
}

var dbService = initDBService()

/**
 * 获取默认数据库引擎
 */
func DefaultDB() *gorm.DB {
	return dbService.Default
}

/**
 * 获取指定数据库引擎
 */
func GetDB(name string) (*gorm.DB, bool) {
	engine, ok := dbService.List[name]
	return engine, ok
}

/**
 * 获取全部数据库引擎列表
 */
func DBList() map[string]*gorm.DB {
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
func SetDefaultDB(name string) error {
	engine, ok := dbService.List[name]
	if !ok {
		return fmt.Errorf("Specified database does not exist: %v.", name)
	}
	dbService.Default = engine
	return nil
}

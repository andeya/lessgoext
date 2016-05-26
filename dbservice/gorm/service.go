package gorm

import (
	"fmt"

	"github.com/jinzhu/gorm"
	//_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql" //github.com/go-sql-driver/mysql
	_ "github.com/jinzhu/gorm/dialects/postgres"
	//	_ "github.com/jinzhu/gorm/dialects/sqlite" //github.com/mattn/go-sqlite3
)

type DBService struct {
	Default *gorm.DB
	List    map[string]*gorm.DB
}

/**
 * 获取默认数据库引擎
 */
func (d *DBService) DefaultDB() *gorm.DB {
	return d.Default
}

/**
 * 获取指定数据库引擎
 */
func (d *DBService) GetDB(name string) (*gorm.DB, bool) {
	engine, ok := d.List[name]
	return engine, ok
}

/**
 * 获取全部数据库引擎列表
 */
func (d *DBService) DBList() map[string]*gorm.DB {
	return d.List
}

/**
 * 设置默认数据库引擎
 */
func (d *DBService) SetDefaultDB(name string) error {
	engine, ok := d.List[name]
	if !ok {
		return fmt.Errorf("Specified database does not exist: %v.", name)
	}
	d.Default = engine
	return nil
}

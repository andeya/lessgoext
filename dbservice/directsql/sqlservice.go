/*
 功能：动态SQL执行函数供其他包调用单元
 作者：畅雨
*/
package directsql

import (
	"database/sql"
	"errors"
	"github.com/go-xorm/core"
	"github.com/henrylee2cn/lessgo"
)

var notFoundError = func(sqlid string) error {
	return errors.New("错误:未定义的sql:" + sqlid)
}
var notMatchError = func() error {
	return errors.New("错误:调用的语句的sqltype与该函数不匹配！")
}

//查询 根据modelId，sqlId ，mp:map[string]interface{}命名参数,返回*core.Rows
func SelectMap(modelId, sqlId string, mp map[string]interface{}) (*core.Rows, error) {
	//获取Sqlentity,db
	se, db := findSqlEntityAndDB(modelId, sqlId)
	if se == nil {
		return nil, notFoundError(modelId + "/" + sqlId)
	}
	//判断类型不是Select 就返回错误
	if se.Sqltype != ST_SELECT {
		return nil, notMatchError()
	}
	return db.QueryMap(se.Sqlcmds[0].Sql, &mp)
}

//查询 返回 []map[string]interface{}
func SelectMap2(modelId, sqlId string, mp map[string]interface{}) ([]map[string]interface{}, error) {
	rows, err := SelectMap(modelId, sqlId, mp)
	if err != nil {
		return nil, err
	}
	return rows2mapObjects(rows)
}

//执行返回多個結果集的多個查询根据modelId，sqlId ，SQLmp:map[string]interface{}命名参数 返回结果 map[string]*Rows
func MultiSelectMap(modelId, sqlId string, mp map[string]interface{}) (map[string]*core.Rows, error) {
	result := make(map[string]*core.Rows)
	//获取Sqlentity,db
	se, db := findSqlEntityAndDB(modelId, sqlId)
	if se == nil {
		return nil, notFoundError(modelId + "/" + sqlId)
	}
	//判断类型不是MULTISELECT 就返回错误
	if se.Sqltype != ST_MULTISELECT {
		return nil, notMatchError()
	}
	//循環每個sql定義
	for i, sqlcmd := range se.Sqlcmds {
		lessgo.Log.Debug("MultiSelectMap :" + sqlcmd.Sql)
		rows, err := db.QueryMap(sqlcmd.Sql, &mp)
		if err != nil {
			return nil, err
		}
		if len(sqlcmd.Rout) == 0 {
			result["data"+string(i)] = rows
		} else {
			result[sqlcmd.Rout] = rows
		}
	}
	return result, nil
}

//多個查询 返回 map[string][]map[string]interface{}
func MultiSelectMap2(modelId, sqlId string, mp map[string]interface{}) (map[string][]map[string]interface{}, error) {
	multirows, err := MultiSelectMap(modelId, sqlId, mp)
	if err != nil {
		return nil, err
	}
	result := make(map[string][]map[string]interface{})
	for key, rows := range multirows {
		single, err := rows2mapObjects(rows)
		if err != nil {
			return nil, err
		}
		result[key] = single
	}
	return result, nil
}

//执行 UPDATE、DELETE、INSERT，mp 是MAP类型命名参数 返回结果 sql.Result
func ExecMap(modelId, sqlId string, mp map[string]interface{}) (sql.Result, error) {
	//获取Sqlentity,db
	se, db := findSqlEntityAndDB(modelId, sqlId)
	if se == nil {
		return nil, notFoundError(modelId + "/" + sqlId)
	}
	//判断类型不是UPDATE、DELETE、INSERT 就返回错误
	if (se.Sqltype != ST_DELETE) || (se.Sqltype != ST_INSERT) || (se.Sqltype != ST_UPDATE) {
		return nil, notMatchError()
	}
	return db.ExecMap(se.Sqlcmds[0].Sql, &mp)
}

//批量执行 UPDATE、INSERT、mp 是MAP类型命名参数
func BacthExecMap(modelId, sqlId string, sp []map[string]interface{}) error {
	//获取Sqlentity,db
	se, db := findSqlEntityAndDB(modelId, sqlId)
	if se == nil {
		return notFoundError(modelId + "/" + sqlId)
	}
	//判断类型不是BATCHINSERT、BATCHUPDATE 就返回错误
	if (se.Sqltype != ST_BATCHINSERT) || (se.Sqltype != ST_BATCHUPDATE) {
		return notMatchError()
	}
	return transact(db, func(tx *core.Tx) error {
		for _, p := range sp {
			lessgo.Log.Debug("BacthExecMap :" + se.Sqlcmds[0].Sql)
			if _, err := tx.ExecMap(se.Sqlcmds[0].Sql, &p); err != nil {
				return err
			}
		}
		return nil
	})
}

//批量执行 BacthComplex、mp 是MAP类型命名参数,事务中依次执行
func BacthExecComplexMap(modelId, sqlId string, mp map[string][]map[string]interface{}) error {
	//获取Sqlentity,db
	se, db := findSqlEntityAndDB(modelId, sqlId)
	if se == nil {
		return notFoundError(modelId + "/" + sqlId)
	}
	//判断类型不是BATCHCOMPLEX 就返回错误
	if se.Sqltype != ST_BATCHCOMPLEX {
		return notMatchError()
	}
	return transact(db, func(tx *core.Tx) error {
		//循環每個sql定義
		for _, sqlcmd := range se.Sqlcmds {
			//循環其批量參數
			if sp, ok := mp[sqlcmd.Pin]; ok {
				for _, p := range sp {
					lessgo.Log.Debug("BacthExecComplexMap :" + sqlcmd.Sql)
					if _, err := tx.ExecMap(sqlcmd.Sql, &p); err != nil {
						return err
					}
				}
			} else {
				return errors.New("错误：传入的参数与SQL节点定义的sql.pin名称不匹配！")
			}
		}
		return nil
	})
}

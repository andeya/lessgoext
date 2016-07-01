/*
  功能：动态SQL执行引擎
  作者：畅雨
*/
package directsql

import (
	"errors"
	"fmt"
	"github.com/go-xorm/core"
	"github.com/lessgo/lessgo"
)

//根据sqlid获取 *Sqlentity
func (ms *ModelSql) findSqlEntity(sqlid string) *Sqlentity {
	if se, ok := ms.Sqlentitys[sqlid]; ok {
		return se
	}
	return nil
}

//执行普通的单个查询SQL  mp 是MAP类型命名参数 map[string]interface{},返回结果 []map[string][]interface{}
func (ms *ModelSql) selectMap(se *Sqlentity, mp map[string]interface{}) ([]map[string]interface{}, error) {
	rows, err := ms.DB.QueryMap(se.Sqlcmds[0].Sql, &mp)
	if err != nil {
		return nil, err
	}
	return rows2mapObjects(rows)
}

//分頁查詢的返回結果
type PagingSelectResult struct {
	Total int                      `json:"total"`
	Data  []map[string]interface{} `json:"data"`
}

//执行分页查询SQL  mp 是MAP类型命名参数 返回结果 int,[]map[string][]interface{}
func (ms *ModelSql) pagingSelectMap(se *Sqlentity, mp map[string]interface{}) (*PagingSelectResult, error) {
	if len(se.Sqlcmds) != 2 {
		return nil, errors.New("错误：分页查询必须定义2个SQL节点，一个获取总页数另一个用于查询数据！")
	}
	//1.获取总页数，約定該SQL放到第二條，並且只返回一條記錄一個字段
	trows, err := ms.DB.QueryMap(se.Sqlcmds[0].Sql, &mp)
	if err != nil {
		return nil, err
	}
	for trows.Next() {
		var total = make([]int, 1)
		err := trows.ScanSlice(&total)
		if err != nil {
			return nil, err
		}
		if len(total) != 1 {
			return nil, errors.New("错误：获取总页数的SQL执行结果非唯一记录！")
		}
		//2.获取当前页數據，約定該SQL放到第二條
		rows, err := ms.DB.QueryMap(se.Sqlcmds[1].Sql, &mp)
		if err != nil {
			return nil, err
		}
		result, err := rows2mapObjects(rows)
		if err != nil {
			return nil, err
		}
		return &PagingSelectResult{Total: total[0], Data: result}, nil //最終的結果
	}
	return nil, err
}

//执行返回多個結果集的多個查询SQL， mp 是MAP类型命名参数 返回结果 map[string][]map[string][]string
func (ms *ModelSql) multiSelectMap(se *Sqlentity, mp map[string]interface{}) (map[string][]map[string]interface{}, error) {
	result := make(map[string][]map[string]interface{})
	//循環每個sql定義
	for i, sqlcmd := range se.Sqlcmds {
		lessgo.Log.Debug("MultiSelectMap :" + sqlcmd.Sql)
		rows, err := ms.DB.QueryMap(sqlcmd.Sql, &mp)
		if err != nil {
			return nil, err
		}
		single, err := rows2mapObjects(rows)
		if err != nil {
			return nil, err
		}
		if len(sqlcmd.Rout) == 0 {
			result["data"+string(i)] = single
		} else {
			result[sqlcmd.Rout] = single
		}
	}
	return result, nil
}

//执行单个查询SQL返回JSON父子嵌套結果集 mp 是MAP类型命名参数 map[string]interface{},返回结果 []map[string][]interface{}
func (ms *ModelSql) nestedSelectMap(se *Sqlentity, mp map[string]interface{}) ([]map[string]interface{}, error) {
	lessgo.Log.Debug("NestedSelectMap :" + se.Sqlcmds[0].Sql)
	rows, err := ms.DB.QueryMap(se.Sqlcmds[0].Sql, &mp)
	if err != nil {
		return nil, err
	}
	return rows2mapObjects(rows)
}

type Execresult struct {
	LastInsertId int64  `json:"lastinsertdd"`
	RowsAffected int64  `json:"rowsaffected"`
	Info         string `json:"info"`
}

//执行 UPDATE、DELETE、INSERT，mp 是 map[string]interface{}, 返回结果 execresult
func (ms *ModelSql) execMap(se *Sqlentity, mp map[string]interface{}) (*Execresult, error) {
	lessgo.Log.Debug("ExecMap :" + se.Sqlcmds[0].Sql)
	Result, err := ms.DB.ExecMap(se.Sqlcmds[0].Sql, &mp)
	if err != nil {
		return nil, err
	}
	LIId, _ := Result.LastInsertId()
	RAffected, _ := Result.RowsAffected()
	return &Execresult{LastInsertId: LIId, RowsAffected: RAffected, Info: "Exec sql ok!"}, nil
}

//批量执行 UPDATE、INSERT、mp 是MAP类型命名参数
func (ms *ModelSql) bacthExecMap(se *Sqlentity, sp []map[string]interface{}) error {
	return transact(ms.DB, func(tx *core.Tx) error {
		for _, p := range sp {
			lessgo.Log.Debug("BacthExecMap :" + se.Sqlcmds[0].Sql)
			if _, err := tx.ExecMap(se.Sqlcmds[0].Sql, &p); err != nil {
				return err
			}
		}
		return nil
	})
}

//批量执行 BacthComplex、mp 是map[string][]map[string]interface{}参数,事务中依次执行
func (ms *ModelSql) bacthExecComplexMap(se *Sqlentity, mp map[string][]map[string]interface{}) error {
	return transact(ms.DB, func(tx *core.Tx) error {
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

//-----------------------------------------------------------------
//ransaction handler 封装在一个事务中执行多个SQL语句
func transact(db *core.DB, txFunc func(*core.Tx) error) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return
	}
	defer func() {
		if p := recover(); p != nil {
			switch p := p.(type) {
			case error:
				err = p
			default:
				err = fmt.Errorf("%s", p)
			}
		}
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()
	return txFunc(tx)
}

/*
//根据sqlid获取 类型
func (ms *ModelSql) getSqltype(sqlid string) Tsqltype {
	if se, ok := ms.Sqlentitys[sqlid]; ok {
		return se.Sqltype
	}
	return -1
}

//执行普通的单个查询SQL  mp 是MAP类型命名参数 map[string]interface{},返回结果 []map[string][]interface{}
func (ms *ModelSql) selectMap(sqlid string, mp map[string]interface{}) ([]map[string]interface{}, error) {
	if se, ok := ms.Sqlentitys[sqlid]; ok {
		rows, err := ms.DB.QueryMap(se.Sqlcmds[0].Sql, &mp)
		if err != nil {
			return nil, err
		}
		return rows2mapObjects(rows)
	}
	return nil, notFoundErr(sqlid)
}

//分頁查詢的返回結果
type PagingSelectResult struct {
	Total int                      `json:"total"`
	Data  []map[string]interface{} `json:"data"`
}

//执行分页查询SQL  mp 是MAP类型命名参数 返回结果 int,[]map[string][]interface{}
func (ms *ModelSql) pagingSelectMap(sqlid string, mp map[string]interface{}) (*PagingSelectResult, error) {
	if se, ok := ms.Sqlentitys[sqlid]; ok {
		if len(se.Sqlcmds) != 2 {
			return nil, errors.New("错误：分页查询必须定义2个SQL节点，一个获取总页数另一个用于查询数据！")
		}
		//lessgo.Log.Debug("PagingSelectMap -0:" + se.Sqlcmds[0].Sql)
		//lessgo.Log.Debug("PagingSelectMap -1:" + se.Sqlcmds[1].Sql)
		//1.获取总页数，約定該SQL放到第二條，並且只返回一條記錄一個字段
		trows, err := ms.DB.QueryMap(se.Sqlcmds[0].Sql, &mp)
		if err != nil {
			return nil, err
		}
		for trows.Next() {
			var total = make([]int, 1)
			err := trows.ScanSlice(&total)
			if err != nil {
				return nil, err
			}
			if len(total) != 1 {
				return nil, errors.New("错误：获取总页数的SQL执行结果非唯一记录！")
			}
			//2.获取当前页數據，約定該SQL放到第二條
			rows, err := ms.DB.QueryMap(se.Sqlcmds[1].Sql, &mp)
			if err != nil {
				return nil, err
			}
			result, err := rows2mapObjects(rows)
			if err != nil {
				return nil, err
			}
			return &PagingSelectResult{Total: total[0], Data: result}, nil //最終的結果
		}
		return nil, err
	}
	return nil, notFoundErr(sqlid)
}

//执行返回多個結果集的多個查询SQL， mp 是MAP类型命名参数 返回结果 map[string][]map[string][]string
//func (ms *ModelSql) pagingSelectMap(sqlid string, mp map[string]interface{})
func (ms *ModelSql) multiSelectMap(sqlid string, mp map[string]interface{}) (map[string][]map[string]interface{}, error) {
	if se, ok := ms.Sqlentitys[sqlid]; ok {
		result := make(map[string][]map[string]interface{})
		//循環每個sql定義
		for i, sqlcmd := range se.Sqlcmds {
			lessgo.Log.Debug("MultiSelectMap :" + sqlcmd.Sql)
			rows, err := ms.DB.QueryMap(sqlcmd.Sql, &mp)
			if err != nil {
				return nil, err
			}
			single, err := rows2mapObjects(rows)
			if err != nil {
				return nil, err
			}
			if len(sqlcmd.Rout) == 0 {
				result["data"+string(i)] = single
			} else {
				result[sqlcmd.Rout] = single
			}
		}
		return result, nil
	}
	return nil, notFoundErr(sqlid)
}

//执行单个查询SQL返回JSON父子嵌套結果集 mp 是MAP类型命名参数 map[string]interface{},返回结果 []map[string][]interface{}
func (ms *ModelSql) nestedSelectMap(sqlid string, mp map[string]interface{}) ([]map[string]interface{}, error) {
	if se, ok := ms.Sqlentitys[sqlid]; ok {
		lessgo.Log.Debug("NestedSelectMap :" + se.Sqlcmds[0].Sql)
		rows, err := ms.DB.QueryMap(se.Sqlcmds[0].Sql, &mp)
		if err != nil {
			return nil, err
		}
		return rows2mapObjects(rows)
	}
	return nil, notFoundErr(sqlid)
}

type Execresult struct {
	LastInsertId int64  `json:"lastinsertdd"`
	RowsAffected int64  `json:"rowsaffected"`
	Info         string `json:"info"`
}

//执行 UPDATE、DELETE、INSERT，mp 是 map[string]interface{}, 返回结果 execresult
func (ms *ModelSql) execMap(sqlid string, mp map[string]interface{}) (*Execresult, error) {
	if se, ok := ms.Sqlentitys[sqlid]; ok {
		lessgo.Log.Debug("ExecMap :" + se.Sqlcmds[0].Sql)
		Result, err := ms.DB.ExecMap(se.Sqlcmds[0].Sql, &mp)
		if err != nil {
			return nil, err
		}
		LIId, _ := Result.LastInsertId()
		RAffected, _ := Result.RowsAffected()
		return &Execresult{LastInsertId: LIId, RowsAffected: RAffected, Info: "Exec sql ok!"}, nil
	}
	return nil, notFoundErr(sqlid)
}

//批量执行 UPDATE、INSERT、mp 是MAP类型命名参数
func (ms *ModelSql) bacthExecMap(sqlid string, sp []map[string]interface{}) error {
	if se, ok := ms.Sqlentitys[sqlid]; ok {
		return transact(ms.DB, func(tx *core.Tx) error {
			for _, p := range sp {
				lessgo.Log.Debug("BacthExecMap :" + se.Sqlcmds[0].Sql)
				if _, err := tx.ExecMap(se.Sqlcmds[0].Sql, &p); err != nil {
					return err
				}
			}
			return nil
		})
	}
	return notFoundErr(sqlid)
}

//批量执行 BacthComplex、mp 是map[string][]map[string]interface{}参数,事务中依次执行
func (ms *ModelSql) bacthExecComplexMap(sqlid string, mp map[string][]map[string]interface{}) error {
	if se, ok := ms.Sqlentitys[sqlid]; ok {
		return transact(ms.DB, func(tx *core.Tx) error {
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
	return notFoundErr(sqlid)
}

*/

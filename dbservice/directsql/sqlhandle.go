/*
   请求处理handle
*/
package directsql

import (
	"github.com/henrylee2cn/lessgo"
)

//DirectSQL handler 定义
var DirectSQL = lessgo.ApiHandler{
	Desc:   "DirectSQL",
	Method: "GET POST",
	Handler: func(ctx *lessgo.Context) error {
		//1.根据路径获取sqlentity:去掉/bos/
		//件路径拆分成 modelId，sqlId
		modelId, sqlId := trimBeforeSplitRight(ctx.Request().URL.Path, '/', 2)
		//获取ModelSql
		ms := findModelSql(modelId)
		if ms == nil {
			lessgo.Log.Error("错误：未定义的Model文件:" + modelId)
			return ctx.JSONMsg(404, 404, "错误：未定义的Model文件:"+modelId)
		}
		//获取Sqlentity
		se := ms.findSqlEntity(sqlId)
		if se == nil {
			lessgo.Log.Error("错误：Model文件中未定义sql:" + modelId)
			return ctx.JSONMsg(404, 404, "错误：Model文件中未定义sql:"+modelId)
		}

		//2.根据SQL类型分别处理执行并返回结果信息
		switch se.Sqltype {
		case ST_PAGINGSELECT: //分页选择SQL，分頁查詢結果不能用cache
			//2.1 获取POST参数並轉換
			var jsonpara map[string]interface{}
			err := ctx.Bind(&jsonpara) //从Body获取JSON参数
			if err != nil {
				lessgo.Log.Error(err.Error())
				return ctx.JSONMsg(404, 404, err.Error())
			}
			//2.2常規參數處理
			var callback string
			if v, ok := jsonpara["callback"]; ok {
				s, ok := v.(string)
				if ok {
					callback = s
					delete(jsonpara, "callback")
				}
			}
			//2.3 執行並返回結果
			data, err := ms.pagingSelectMap(se, jsonpara)
			if err != nil {
				lessgo.Log.Error(err.Error())
				return ctx.JSONMsg(404, 404, err.Error())
			}
			if len(callback) > 0 {
				return ctx.JSONP(200, callback, *data)
			}
			return ctx.JSON(200, data)

		case ST_SELECT, ST_NESTEDSELECT, ST_INSERTPRO: //一般选择SQL,嵌套选择暂时未实现跟一般选择一样,增强的插入后返回sysid ----OK
			//2.1 获取POST参数並轉換
			var jsonpara map[string]interface{}
			err := ctx.Bind(&jsonpara) //从Body获取JSON参数
			if err != nil {
				lessgo.Log.Error(err.Error())
				return ctx.JSONMsg(404, 404, err.Error())
			}
			//2.2常規參數處理
			var callback string
			if v, ok := jsonpara["callback"]; ok {
				s, ok := v.(string)
				if ok {
					callback = s
					delete(jsonpara, "callback")
				}
			}
			/*該Cache功能暫不實現
			if v, ok := jsonpara["recache"]; ok {
				r, ok := v.(int)
				if ok {
					delete(jsonpara, "recache")
					fmt.Println(r)
				}
			}*/
			//2.3 執行並返回結果
			data, err := ms.selectMap(se, jsonpara)
			if err != nil {
				lessgo.Log.Error(err.Error())
				return ctx.JSONMsg(404, 404, err.Error())
			}
			if len(callback) > 0 {
				return ctx.JSONP(200, callback, data)
			}
			return ctx.JSON(200, data)

		case ST_MULTISELECT: //返回多結果集選擇
			//2.1 获取POST参数並轉換
			var jsonpara map[string]interface{}
			err := ctx.Bind(&jsonpara) //从Body获取JSON参数
			if err != nil {
				lessgo.Log.Error(err.Error())
				return ctx.JSONMsg(404, 404, err.Error())
			}
			//2.2常規參數處理
			var callback string
			if v, ok := jsonpara["callback"]; ok {
				s, ok := v.(string)
				if ok {
					callback = s
					delete(jsonpara, "callback")
				}
			}
			//2.3 執行並返回結果
			data, err := ms.multiSelectMap(se, jsonpara)
			if err != nil {
				lessgo.Log.Error(err.Error())
				return ctx.JSONMsg(404, 404, err.Error())
			}
			if len(callback) > 0 {
				return ctx.JSONP(200, callback, data)
			}
			return ctx.JSON(200, data)

		/*
		 case ST_NESTEDSELECT: //返回嵌套結果集選擇

			return ctx.JSON(200, "data")*/

		case ST_UPDATE, ST_INSERT, ST_DELETE: //更新\插入\删除
			//2.1.获取 Ajax post json 参数
			var jsonpara map[string]interface{}
			err := ctx.Bind(&jsonpara) //从Body获取 json参数
			if err != nil {
				lessgo.Log.Error(err.Error())
				return ctx.JSONMsg(404, 404, err.Error())
			}
			data, err := ms.execMap(se, jsonpara)
			if err != nil {
				lessgo.Log.Error(err.Error())
				return ctx.JSONMsg(404, 404, err.Error())
			}
			return ctx.JSON(200, data)

		case ST_BATCHUPDATE, ST_BATCHINSERT: //批量插入、更新
			//2.1.获取 Ajax post json 参数
			var jsonpara []map[string]interface{}
			err := ctx.Bind(&jsonpara) //从Body获取 json参数
			if err != nil {
				lessgo.Log.Error(err.Error())
				return ctx.JSONMsg(404, 404, err.Error())
			}
			err = ms.bacthExecMap(se, jsonpara)
			if err != nil {
				lessgo.Log.Error(err.Error())
				return ctx.JSONMsg(404, 404, err.Error())
			}
			return ctx.JSONMsg(200, 200, "Exec bacth sql ok!")

		case ST_BATCHCOMPLEX: //批量複合語句
			//2.1.获取 Ajax post json 参数
			var jsonpara map[string][]map[string]interface{}
			err := ctx.Bind(&jsonpara) //从Body获取 json参数
			if err != nil {
				lessgo.Log.Error(err.Error())
				return ctx.JSONMsg(404, 404, err.Error())
			}
			//2.2執行語句
			err = ms.bacthExecComplexMap(se, jsonpara)
			if err != nil {
				lessgo.Log.Error(err.Error())
				return ctx.JSONMsg(404, 404, err.Error())
			}
			return ctx.JSONMsg(200, 200, "Exec bacth complex sql ok!")
		}
		return ctx.JSONMsg(404, 404, "Undefined sqltype!")
	},
}.Reg()

//重新载入全部ModelSql配置文件
var DirectSQLReloadAll = lessgo.ApiHandler{
	Desc:   "DirectSQL ModelSql Reload",
	Method: "GET",
	Handler: func(ctx *lessgo.Context) error {
		ReloadAll()
		return ctx.JSONMsg(200, 200, "Reload all modelsqls file ok!")
	},
}.Reg()

//重新载入单个ModelSql配置文件
var DirectSQLReloadModel = lessgo.ApiHandler{
	Desc:   "DirectSQL ModelSql Reload",
	Method: "GET",
	Handler: func(ctx *lessgo.Context) error {
		//ctx.Request().URL.Path, '/', 2) 去掉 /bom/reload/
		err := ReloadModel(trimBefore(ctx.Request().URL.Path, '/', 3))
		if err != nil {
			return err
		}
		return ctx.JSONMsg(200, 200, "Reload the modelsql file ok!")
	},
}.Reg()

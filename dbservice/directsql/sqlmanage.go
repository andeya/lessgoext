/*
  管理动态SQL,根据配置文件目录从配置文件加载到内存中
  作者：畅雨
*/
package directsql

import (
	"encoding/xml"
	"github.com/fsnotify/fsnotify"
	"github.com/go-xorm/core"
	"github.com/lessgo/lessgo"
	confpkg "github.com/lessgo/lessgo/config"
	lessgoxorm "github.com/lessgo/lessgoext/dbservice/xorm"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

//var modelsqls map[string]*ModelSql

//配置文件配置参数
const MSCONFIGFILE = "config/directsql.config"

// 全部业务SQL路由表,不根据目录分层次，直接放在map sqlmodels中，key=带路径不带扩展名的文件名比如
type ModelSqls struct {
	roots     map[string]string    //需要载入的根目录(可以多个)-短名=实际路径名
	modelsqls map[string]*ModelSql //全部定义模型对象
	extension string               //模型定义文件的扩展名(默认为.msql)
	loadLock  sync.RWMutex
	watcher   *fsnotify.Watcher //监控文件变化的wather
}

//全局所有业务模型对象
var globalmodelSqls = &ModelSqls{
	modelsqls: make(map[string]*ModelSql)}

//sqlmodel 一个配置文件的SQLModel对应的结构
type ModelSql struct {
	Id         string                //root起用映射、不带扩展名的文件名
	DB         *core.DB              //本模块的db引擎 *xorm.Engine.DB()
	Sqlentitys map[string]*Sqlentity //sqlentity key=sqlentity.id
}

//临时转换用，因为 XML 不支持解析到 map，所以先读入到[]然后再根据[]创建map
type tempModelSql struct {
	XMLName    xml.Name     `xml:"model"`
	Id         string       `xml:"id,attr"` //不带扩展名的文件名
	Database   string       `xml:"database,attr"`
	Sqlentitys []*Sqlentity `xml:"sql"`
}

//sqlentity <Select/>等节点对应的结构
type Sqlentity struct {
	XMLName xml.Name `xml:"sql"`
	Id      string   `xml:"id,attr"` //sqlid
	//DB          *xorm.Engine `xml:"-"`       ///执行SQL的数据库引擎移除到本Model中
	Sqltypestr  string    `xml:"type,attr"`
	Sqltype     Tsqltype  `xml:"-"`                //SQL类型
	Idfield     string    `xml:"idfield,attr"`     //SQlType为6=嵌套jsoin树时的ID字段
	Pidfield    string    `xml:"pidfield,attr"`    //SQlType为6=嵌套jsoin树时的ParentID字段
	Sqlcmds     []*Sqlcmd `xml:"cmd"`              // sqlcmd(sqltype为分页查询时的计数SQL放第一个，结果SQL放第二个)
	Transaction bool      `xml:"transaction,attr"` //是否启用事务 1=启用 0=不启用
	//Resultcached bool      `xml:"resultcached,attr"` //是否需要缓存结果 1=缓存 0=不缓存
	//resultcache     inteface{}  //缓存内容
}

//sqlcmd  <Select/>等节点的下级节点<sql />对应结构
type Sqlcmd struct {
	XMLName xml.Name `xml:"cmd"`
	Pin     string   `xml:"in,attr"`   //输入参数标示
	Rout    string   `xml:"out,attr"`  //输出结果标示
	Sql     string   `xml:",chardata"` //SQL
}

//sqlentity 类型
type Tsqltype int

const (
	ST_SELECT       Tsqltype = iota //0=普通查询 ---OK!
	ST_PAGINGSELECT                 //1=分页查询 ---OK!
	ST_NESTEDSELECT                 //2=嵌套jsoin树---------
	ST_MULTISELECT                  //3=多结果集查询---OK!
	ST_DELETE                       //4=删除 ---OK!
	ST_INSERT                       //5=插入 ---OK!
	ST_UPDATE                       //6=更新 ---OK!
	ST_BATCHINSERT                  //7=批量插入(单数据集批量插入)---OK!
	ST_BATCHUPDATE                  //8=批量更新(单数据集批量更新)---OK!
	ST_BATCHCOMPLEX                 //9=批量复合SQL(一般多数据集批量插入或更新)---OK!
	ST_INSERTPRO                    //10=插入升级版，可以在服务端生成key的ID并返回到客户端
)

func init() {
	globalmodelSqls.loadModelSqls()

}

//读入全部模型
func (mss *ModelSqls) loadModelSqls() {
	mss.loadLock.Lock()
	defer mss.loadLock.Unlock()
	//打开directsql的配置文件
	cfg, err := confpkg.NewConfig("ini", MSCONFIGFILE)
	if err != nil {
		lessgo.Log.Error(err.Error())
		return
	}

	//读取ModelSQL文件个根目录
	roots, err := cfg.GetSection("roots")
	if err != nil {
		lessgo.Log.Error(err.Error())
		return
	}
	mss.roots = roots

	//读取扩展名，读取不到就用默认的.msql
	ext := cfg.DefaultString("ext", ".msql")
	mss.extension = ext

	//根据路径遍历加载
	for _, value := range mss.roots {
		lessgo.Log.Debug(value)
		err = filepath.Walk(value, mss.walkFunc)
		if err != nil {
			lessgo.Log.Error(err.Error())
			//return
		}

	}
	//是否监控文件变化
	watch := cfg.DefaultBool("watch", false)
	if watch {
		err := mss.StartWatcher()
		if err != nil {
			lessgo.Log.Error(err.Error())
		}
	}

}

//将带路径文件名处理成 ModelSql的 id 示例： bizmodel\demo.msql  --> biz/demo
func (mss *ModelSqls) filenameToModelId(path string) string {

	key := strings.Replace(path, "\\", "/", -1)
	key = strings.TrimSuffix(key, mss.extension) //去掉扩展名
	for root, value := range mss.roots {
		if strings.HasPrefix(key, value) {
			key = strings.Replace(key, value, root, 1) //处理前缀,将定义的根路径替换为名称
			break
		}
	}
	return key
}

//遍历子目录文件处理函数
func (mss *ModelSqls) walkFunc(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}
	if strings.HasSuffix(path, mss.extension) {
		ms, err := mss.parseModelSql(path)
		if err != nil {
			lessgo.Log.Error(err.Error())
			return nil //单个文件解析出错继续加载其他的文件
			//return err
		}
		//将本文件对应的ModelSql放入到ModelSqls
		mss.modelsqls[mss.filenameToModelId(path)] = ms
		lessgo.Log.Debug("ModelSql file: " + path + " ------> " + mss.filenameToModelId(path) + "  loaded. ")
	}
	return nil
}

//解析单个ModelSQL定义文件
func (mss *ModelSqls) parseModelSql(msqlfile string) (*ModelSql, error) {
	//读取文件
	content, err := ioutil.ReadFile(msqlfile)
	if err != nil {
		return nil, err
	}
	tempresult := tempModelSql{}
	err = xml.Unmarshal(content, &tempresult)
	if err != nil {
		return nil, err
	}
	//
	dbe, ok := lessgoxorm.GetDB(tempresult.Database)
	if ok == false {
		dbe = lessgoxorm.DefaultDB()
		//lessgo.Log.Debug("database:", tempresult.Database)
	}
	//定义一个 ModelSql将 tempModelSql 转换为 ModelSql
	result := &ModelSql{Id: tempresult.Id, DB: dbe.DB(), Sqlentitys: make(map[string]*Sqlentity)}
	//处理一遍：设置数据库访问引擎，设置Sqlentity的类似
	for _, se := range tempresult.Sqlentitys {
		//处理SQL类型
		switch se.Sqltypestr {
		case "pagingselect":
			se.Sqltype = ST_PAGINGSELECT
		case "nestedselect":
			se.Sqltype = ST_NESTEDSELECT
		case "multiselect":
			se.Sqltype = ST_MULTISELECT
		case "select":
			se.Sqltype = ST_SELECT
		case "delete":
			se.Sqltype = ST_DELETE
		case "insert":
			se.Sqltype = ST_INSERT
		case "update":
			se.Sqltype = ST_UPDATE
		case "batchinsert":
			se.Sqltype = ST_BATCHINSERT
		case "batchupdate":
			se.Sqltype = ST_BATCHUPDATE
		case "batchcomplex":
			se.Sqltype = ST_BATCHCOMPLEX
		case "insertpro":
			se.Sqltype = ST_INSERTPRO
		}
		result.Sqlentitys[se.Id] = se
	}

	return result, nil
}

//--------------------------------------------------------------------------

// 获取sqlentity SQL的执行实体
func (mss *ModelSqls) findsqlentity(modelid string, sqlid string) *Sqlentity {
	if sm, ok := mss.modelsqls[modelid]; ok {
		if se, ok := sm.Sqlentitys[sqlid]; ok {
			return se
		}
	}
	return nil
}

// 获取sqlentity SQL的执行实体与DB执行引擎
func (mss *ModelSqls) findsqlentityanddb(modelid string, sqlid string) (*Sqlentity, *core.DB) {
	if sm, ok := mss.modelsqls[modelid]; ok {
		if se, ok := sm.Sqlentitys[sqlid]; ok {
			return se, sm.DB
		}
	}
	return nil, nil
}

// 根据路径加文件名(不带文件扩展名)获取其ModelSql
func (mss *ModelSqls) findmodelsql(modelid string) *ModelSql {
	if sm, ok := mss.modelsqls[modelid]; ok {
		return sm
	}
	return nil
}

//文件内容改变重新载入(新增、修改的都触发)
func (mss *ModelSqls) refreshModelFile(msqlfile string) error {
	mss.loadLock.Lock()
	defer mss.loadLock.Unlock()
	//重新解析
	ms, err := mss.parseModelSql(msqlfile)
	if err != nil {
		lessgo.Log.Error(err.Error())
		return err //单个文件解析出错继续加载其他的文件
		//return err
	}
	//将本文件对应的ModelSql放入到ModelSqls
	mss.modelsqls[mss.filenameToModelId(msqlfile)] = ms
	return nil
}

//文件已经被移除，从内存中删除
func (mss *ModelSqls) removeModelFile(msqlfile string) error {
	mss.loadLock.Lock()
	defer mss.loadLock.Unlock()
	delete(mss.modelsqls, mss.filenameToModelId(msqlfile))
	return nil
}

//文件改名---暂无实现
func (mss *ModelSqls) renameModelFile(msqlfile, newfilename string) error {
	//err := mss.removeModelFile(msqlfile)
	//err = mss.refreshModelFile(newfilename)
	return nil
}

//单元访问文件--------------------------------------------------------------
func findSqlEntityAndDB(modelid string, sqlid string) (*Sqlentity, *core.DB) {
	return globalmodelSqls.findsqlentityanddb(modelid, sqlid)
}

// findSqlEntity 获取sqlentity SQL的执行实体
func findSqlEntity(modelid string, sqlid string) *Sqlentity {
	lessgo.Log.Debug("ModelSqlPath: " + modelid + " ,SqlId: " + sqlid)
	return globalmodelSqls.findsqlentity(modelid, sqlid)
}

//根据ModelSql文件路径获取 ModelSql
func findModelSql(modelid string) *ModelSql {
	lessgo.Log.Debug("ModelSqlPath: " + modelid)
	return globalmodelSqls.findmodelsql(modelid)
}

//重置配置文件全部重新载入,API：/bom/reload  handle调用
func ReloadAll() {
	globalmodelSqls = &ModelSqls{
		modelsqls: make(map[string]*ModelSql)}
	globalmodelSqls.loadModelSqls()
}

//重新载入单个模型文件---未测试
func ReloadModel(msqlfile string) error {
	//已经去掉 "/bom/reload/",需要加上扩展名
	return globalmodelSqls.refreshModelFile(msqlfile + globalmodelSqls.extension)
}

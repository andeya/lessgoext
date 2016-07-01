##直接执行SQL的引擎
    - 通过url直接执行SQL，参数采用POST方式传递  /bos/... sql标示
    - sqlengine--API执行sql的引擎
    - sqlhandle--处理执行请求的handle(API)
    - sqlrouter--SQL路由处理
    - sqlmanage--管理维护SQL定义文件到内存中
    - sqlservice--通过代码调用的接口单元
    - sqlhelper--辅助函数
    - sqlwatcher--SQL配置文件监控自动更新(  属于实现文件新增修改，删除监控，改名不行)
    - 系统中通过代码如何调用：
        directsql/sqlService 单元中的函数

##配置文件：/config/directsql.config 的内容
    ;SQL配置文件扩展名，只能一个。
    ext=.msql 
    ;是否开始监控所有roots目录下的配置文件变化，改变自动处理(增加，删除，修改)
    watch=true
    ;SQL配置文件加载的根目录，可以个多个，定义后自动将真实文件名映射到前边名称
    [roots]
    biz=bizmodel  ; 比如： 系统根目录/bizmodel/plan/main.msql 访问url为 bos/biz/plan/main
    sys=sysmodel  ; 比如： 系统根目录/sysmodel/admin/users.msql 访问url为 bos/sys/admin/users

##全部SQL类型
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
    ST_INSERTPRO                  //10=插入升级版，可以在服务端生成key的ID并返回到客户端

##剩余问题
    - nestedselect-返回嵌套JSON的未实现
    - sqlwatcher中不能监控改名的文件（可以通过重新载入解决）

##参数通过 POST
   //-----------------------以下为传入的JSON参数解-------------------------
    - 简单json参数---ok
     {"id":"001","name":"aaaaa"}
     其中  "callback":"可选参数，不为空时返回JSONP",
        "start":"可选参数，分页查询时需要 开始记录数",
        "limted":"可选参数，分页查询时需要 每页记录数",
        "recache":"可选参数，如果结果配置的为缓存，通过该参数刷新缓存",

    - 简单批量的json---ok
      [{"id":"001","name":"aaaaa"},{"id":"002","name":"bbbbb"},{"id":"003","name":"ccc"}]

    - 复杂批量json---Ok
    {"main":[{"id":"01","name":"aaaaa"},{"id":"002","name":"bbbbb"}],
     "sub1":[{"id":"0111","pid":"01","name":"sub11"},{"id":"0112","pid":"01","name":"sub12"}]
    "sub2":[{"id":"0121","pid":"01","name":"sub21"},{"id":"0122","pid":"01","name":"sub22"}]
    }

##嵌套的MAP
    根據 sql.rows  {id,pid,name...} 
    構造嵌套的MAP
    {Id:'01',
    Name:'xxx',
    Childen:[{
        Id:'0101',
        Pid:'01',
        Name:'xxx',
        Childen:[{
            Id:'010101',
            Pid:'0101',
            Name:'xxx'},
            {
            Id:'010102',
            Pid:'0101',
            Name:'xxx'}]
        }]
       }
      }
    }


## Lessgo项目部署工具

1.下载框架源码
```sh
go get github.com/lessgo/lessgo
go get github.com/lessgo/lessgoext/...
```

2.安装部署工具
```sh
cd %GOPATH%/github.com/lessgo/lessgoext/lessgo
go install
```
(该工具将会自动创建一套Demo，以供学习与开发)

3.创建项目（在项目目录下运行cmd）
```sh
$ lessgo new appname
```

4.以热编译模式运行（在项目目录下运行cmd）
```sh
$ cd appname
$ lessgo run
```

##项目目录结构

```
─Project 项目开发目录
├─config 配置文件目录
│  ├─app.config 系统应用配置文件
│  └─db.config 数据库配置文件
├─common 后端公共目录
│  └─... 如utils等其他
├─middleware 后端公共中间件目录
├─static 前端公共目录 (url: /static)
│  ├─tpl 公共tpl模板目录
│  ├─js 公共js目录 (url: /static/js)
│  ├─css 公共css目录 (url: /static/css)
│  ├─img 公共img目录 (url: /static/img)
│  └─plugin 公共js插件 (url: /static/plugin)
├─uploads 默认上传下载目录
├─router 源码路由配置
│  ├─sys_router.go 系统模块路由文件
│  ├─biz_router.go 业务模块路由文件
├─sys_handler 系统模块后端目录
│  ├─xxx 子模块目录
│  │  ├─example.go example操作
│  │  └─... xxx的子模块目录
│  └─... 其他子模块目录
├─sys_model 系统模块数据模型目录
├─sys_view 系统模块前端目录 (url: /sys)
│  ├─xxx 与sys_handler对应的子模块目录 (url: /sys/xxx)
│  │  ├─example.tpl 相应操作的模板文件
│  │  ├─example2.html 无需绑定操作的静态html文件
│  │  ├─xxx.css css文件(可有多个)
│  │  ├─xxx.js js文件(可有多个)
│  │  └─... xxx的子模块目录
├─biz_handler 业务模块后端目录
│  ├─xxx 子模块目录
│  │  ├─example.go example操作
│  │  └─... xxx的子模块目录
│  └─... 其他子模块目录
├─biz_model 业务模块数据模型目录
├─biz_view 业务模块前端目录 (url: /biz)
│  ├─xxx 与biz_handler对应的子模块目录 (url: /biz/xxx)
│  │  ├─example.tpl 相应操作的模板文件
│  │  ├─example2.html 无需绑定操作的静态html文件
│  │  ├─xxx.css css文件(可有多个)
│  │  ├─xxx.js js文件(可有多个)
│  │  └─... xxx的子模块目录
├─database 默认数据库文件存储目录
├─logger 运行日志输出目录
└─main.go 应用入口文件
```
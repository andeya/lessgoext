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

##项目组织目录

```
─Project 项目开发目录
├─Config 配置文件目录
│  ├─app.config 系统应用配置文件
│  └─db.config 数据库配置文件
├─Common 后端公共目录
│  ├─Middleware 中间件目录
│  └─... 其他
├─Static 前端公共目录 (url: /static)
│  ├─Tpl 公共tpl模板目录
│  ├─Js 公共js目录 (url: /static/js)
│  ├─Css 公共css目录 (url: /static/css)
│  ├─Img 公共img目录 (url: /static/img)
│  └─Plugin 公共js插件 (url: /static/plugin)
├─SystemAPI 系统模块后端目录
│  ├─SysRouter.go 系统模块路由文件
│  ├─Xxx Xxx子模块目录
│  │  ├─ExampleHandle.go Example操作
│  │  ├─ExampleModel.go Example数据模型及模板函数
│  │  └─... Xxx的子模块目录
│  └─... 其他子模块目录
├─SystemView 系统模块前端目录 (url: /system)
│  ├─Xxx Xxx子模块目录 (url: /system/xxx)
│  │  ├─example.tpl ExampleHandle对应的模板文件
│  │  ├─example2.html 无需绑定操作的静态html文件
│  │  ├─xxx.css css文件(可有多个)
│  │  ├─xxx.js js文件(可有多个)
│  │  └─... Xxx的子模块目录
├─BusinessAPI 业务模块后端目录
│  ├─BusRouter.go 业务模块路由文件
│  ├─Xxx Xxx子模块目录
│  │  ├─ExampleHandle.go Example操作
│  │  ├─ExampleModel.go Example数据模型及模板函数
│  │  └─... Xxx的子模块目录
│  └─... 其他子模块目录
├─BusinessView 业务模块前端目录 (url: /business)
│  ├─Xxx Xxx子模块目录 (url: /business/xxx)
│  │  ├─example.tpl ExampleHandle对应的模板文件
│  │  ├─example2.html 无需绑定操作的静态html文件
│  │  ├─xxx.css css文件(可有多个)
│  │  ├─xxx.js js文件(可有多个)
│  │  └─... Xxx的子模块目录
├─Uploads 默认上传下载目录
├─Logger 运行日志输出目录
└─Main.go 应用入口文件
```


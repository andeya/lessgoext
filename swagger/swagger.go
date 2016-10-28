package swagger

import (
	"errors"
	"fmt"
	"net/http"
	"path"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/lessgo/lessgo"
	"github.com/lessgo/lessgoext/myconfig"
)

/*
 * API自动化文档
 * 仅限局域网访问
 */

/**
 * 注册"/apidoc"路由
 * 配置文件config/apidoc_allow.myconfig说明
 * 参数"checkrealip=true"，表示以真实ip为准进行过滤访问
 * 参数"freeaccess=true"，表示允许任意ip访问
 * 参数"freeaccess=false ipprefix="，为空表示不允许任何ip访问
 * 参数"freeaccess=false ipprefix=192;202"，表示仅允许以192或202开头的ip访问
 */
func Reg() {
	lessgo.Root(
		lessgo.Leaf(jsonUrl, swaggerHandle, allowApidoc),
		lessgo.Leaf("/apidoc/*filepath", apidocHandle, allowApidoc),
	)
	if apidocConfig.FreeAccess {
		lessgo.Log.Sys(`Swagger API's URL path is '/apidoc' [free access]`)
	} else if len(apidocConfig.IpPrefix) == 0 {
		lessgo.Log.Sys(`Swagger API's URL path is '/apidoc' [no access]`)
	} else if apidocConfig.CheckRealIp {
		lessgo.Log.Sys(`Swagger API's URL path is '/apidoc' [check real ip for filter]`)
	} else {
		lessgo.Log.Sys(`Swagger API's URL path is '/apidoc' [check direct ip for filter]`)
	}
}

var (
	apidoc     *Swagger
	virtRouter *lessgo.VirtRouter
	rwlock     sync.RWMutex
	jsonUrl    = "/swagger.json"
	scheme     = func() string {
		if lessgo.Config.Listen.EnableTLS {
			return "https"
		} else {
			return "http"
		}
	}()
	swaggerHandle = &lessgo.ApiHandler{
		Desc:   "swagger",
		Method: "GET",
		Handler: func(c *lessgo.Context) error {
			rwlock.RLock()
			canSet := virtRouter != lessgo.RootRouter()
			rwlock.RUnlock()
			if canSet {
				resetApidoc(c.Request().Host)
			} else {
				apidoc.Host = c.Request().Host // 根据请求动态设置host，修复因首次访问为localhost时，其他ip无法使用的bug
			}
			return c.JSON(200, apidoc)
		},
	}
	apidocHandle = &lessgo.ApiHandler{
		Desc:   "apidoc",
		Method: "GET",
		Handler: func(c *lessgo.Context) error {
			http.StripPrefix("/apidoc/", http.FileServer(assetFS())).
				ServeHTTP(c.Response(), c.Request())
			return nil
		},
	}

	// 静态目录路由参数
	staticParam = &Parameter{
		In:          "path",
		Name:        "static",
		Type:        build("*"),
		Description: "any static path or file",
		Required:    true,
		Format:      fmt.Sprintf("%T", "*"),
		Default:     "",
	}

	// 拦截IP前缀的中间件
	allowApidoc = lessgo.ApiMiddleware{
		Name: "allowApidoc",
		Middleware: func(next lessgo.HandlerFunc) lessgo.HandlerFunc {
			return func(c *lessgo.Context) error {
				if apidocConfig.FreeAccess {
					return next(c)
				}

				if len(apidocConfig.IpPrefix) == 0 {
					return c.Failure(http.StatusForbidden, errors.New(`no access`))
				}

				var remoteAddress string
				if apidocConfig.CheckRealIp {
					remoteAddress = c.RealRemoteAddr()
				} else {
					remoteAddress = c.Request().RemoteAddr
				}
				for _, ipPrefix := range apidocConfig.IpPrefix {
					if strings.HasPrefix(remoteAddress, ipPrefix) {
						return next(c)
					}
				}

				return c.Failure(http.StatusForbidden, errors.New(`not allow to access: `+remoteAddress))
			}
		},
	}.Reg()
)

// 配置被允许访问的ip前缀规则
type ApidocAllow struct {
	CheckRealIp bool //是否检查真实IP
	FreeAccess  bool //允许任意IP访问
	IpPrefix    []string
}

var apidocConfig = func() *ApidocAllow {
	conf := &ApidocAllow{
		CheckRealIp: false,
		FreeAccess:  false,
		IpPrefix: []string{
			"[:",
			"::",
			"127.",
			"192.168.",
			"10.",
		},
	}
	err := myconfig.Sync(conf)
	if err != nil {
		lessgo.Log.Error("%s", err.Error())
		return conf
	}

	ipPrefixMap := map[string]bool{}
	for _, ipPrefix := range conf.IpPrefix {
		if len(ipPrefix) > 0 {
			ipPrefixMap[ipPrefix] = true
		}
	}
	conf.IpPrefix = conf.IpPrefix[:0]
	for ipPrefix := range ipPrefixMap {
		conf.IpPrefix = append(conf.IpPrefix, ipPrefix)
	}

	return conf
}()

// 构建api文档Swagger对象
func resetApidoc(host string) {
	rwlock.Lock()
	defer rwlock.Unlock()
	if virtRouter == lessgo.RootRouter() {
		return
	}
	virtRouter = lessgo.RootRouter()
	rootTag := &Tag{
		Name:        virtRouter.Path(),
		Description: tagDesc(virtRouter.Description()),
	}
	apidoc = &Swagger{
		Version: SwaggerVersion,
		Info: &Info{
			Title:          strings.Title(lessgo.Config.AppName) + " API",
			Description:    lessgo.Config.Info.Description,
			ApiVersion:     lessgo.Config.Info.Version,
			Contact:        &Contact{Email: lessgo.Config.Info.Email},
			TermsOfService: lessgo.Config.Info.TermsOfServiceUrl,
			License: &License{
				Name: lessgo.Config.Info.License,
				Url:  lessgo.Config.Info.LicenseUrl,
			},
		},
		Host:     host,
		BasePath: "/",
		Tags:     []*Tag{rootTag},
		Schemes:  []string{scheme},
		Paths:    map[string]map[string]*Opera{},
		// SecurityDefinitions: map[string]map[string]interface{}{},
		// Definitions:         map[string]Definition{},
		// ExternalDocs:        map[string]string{},
	}

	for _, child := range virtRouter.Children {
		if child.Type == lessgo.HANDLER {
			// 添加API操作项
			addpath(child, rootTag)
			continue
		}
		childTag := &Tag{
			Name:        child.Path(),
			Description: tagDesc(child.Description()),
		}
		apidoc.Tags = append(apidoc.Tags, childTag)
		for _, grandson := range child.Children {
			if grandson.Type == lessgo.HANDLER {
				// 添加API操作项
				addpath(grandson, childTag)
				continue
			}
			grandsonTag := &Tag{
				Name:        grandson.Path(),
				Description: tagDesc(grandson.Description()),
			}
			apidoc.Tags = append(apidoc.Tags, grandsonTag)
			for _, vr := range grandson.Progeny() {
				if vr.Type == lessgo.HANDLER {
					// 添加API操作项
					addpath(vr, grandsonTag)
					continue
				}
			}
		}
	}
}

// 添加API操作项
func addpath(vr *lessgo.VirtRouter, tag *Tag) {
	operas := map[string]*Opera{}
	pid := createPath(vr)
	Summary := summary(vr.Description())
	Description := desc(vr)
	for _, method := range vr.Methods() {
		if method == lessgo.CONNECT || method == lessgo.TRACE {
			continue
		}
		if method == lessgo.WS {
			method = lessgo.GET
		}
		o := &Opera{
			Tags:        []string{tag.Name},
			Summary:     Summary,
			Description: Description,
			OperationId: vr.Id,
			Consumes:    CommonMIMETypes,
			Produces:    CommonMIMETypes,
			Responses:   make(map[string]*Resp, 1),
			// Security: []map[string][]string{},
		}

		// 固定参数路由
		for _, param := range vr.Params() {
			p := &Parameter{
				In:          param.In,
				Name:        param.Name,
				Description: param.Desc,
				Required:    param.Required,
				// Items:       &Items{},
				// Schema:      &Schema{},
			}
			typ := build(param.Model)
			switch p.In {
			default:
				switch typ {
				case "file":
					o.Consumes = []string{"multipart/form-data"}
					p.Type = typ

				case "array":
					subtyp, first := slice(param.Model)
					switch subtyp {
					case "object":
						ref := definitions(vr.Path(), p.Name, method, param.Model)
						p.Schema = &Schema{
							Type: typ,
							Items: &Items{
								Ref: "#/definitions/" + ref,
							},
						}

					default:
						p.Type = typ
						p.Items = &Items{
							Type:    subtyp,
							Enum:    param.Model,
							Default: first,
						}
						p.CollectionFormat = "multi"
					}

				case "object":
					ref := definitions(vr.Path(), p.Name, method, param.Model)
					p.Schema = &Schema{
						Type: typ,
						Ref:  "#/definitions/" + ref,
					}

				default:
					p.Type = typ
					p.Format = fmt.Sprintf("%T", param.Model)
					p.Default = param.Model
				}
			}

			o.Parameters = append(o.Parameters, p)
		}

		// 静态目录路由
		if strings.HasSuffix(pid, "/{static}") {
			o.Parameters = append(o.Parameters, staticParam)
		}

		// 响应结果描述
		var resp = new(Resp)
		switch l := len(vr.HTTP200()); l {
		case 0:
			resp.Description = "successful operation"
		case 1:
			ref := definitions(vr.Path(), "http200", method, vr.HTTP200()[0])
			resp.Schema = &Schema{
				Ref:  "#/definitions/" + ref,
				Type: "object",
			}
		default:
			m := make(map[string]lessgo.Result, l)
			for _, ret := range vr.HTTP200() {
				m[fmt.Sprintf("Code == %v", ret.Code)] = ret
			}
			ref := definitions(vr.Path(), "http200", method, m)
			resp.Schema = &Schema{
				Ref:  "#/definitions/" + ref,
				Type: "object",
			}
		}
		o.Responses["200"] = resp

		operas[strings.ToLower(method)] = o
	}
	if _operas, ok := apidoc.Paths[pid]; ok {
		for k, v := range operas {
			_operas[k] = v
		}
	} else {
		apidoc.Paths[pid] = operas
	}
}

func properties(obj interface{}) map[string]*Property {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	for {
		if t.Kind() != reflect.Ptr {
			break
		}
		t = t.Elem()
		v = v.Elem()
	}
	if t.Kind() == reflect.Slice {
		t = t.Elem()
		if v.Len() > 0 {
			v = v.Index(0)
		} else {
			v = reflect.Value{}
		}
	}
	for {
		if t.Kind() != reflect.Ptr {
			break
		}
		t = t.Elem()
	}
	for {
		if v.Kind() != reflect.Ptr {
			break
		}
		v = v.Elem()
	}
	if v == (reflect.Value{}) {
		v = reflect.New(t).Elem()
	}

	ps := map[string]*Property{}
	switch t.Kind() {
	case reflect.Map:
		kvs := v.MapKeys()
		for _, kv := range kvs {
			val := v.MapIndex(kv)
			for {
				if val.Kind() != reflect.Ptr {
					break
				}
				val = val.Elem()
			}
			if val == (reflect.Value{}) {
				val = reflect.New(val.Type()).Elem()
			}
			p := &Property{
				Type:    build(val.Type()),
				Format:  val.Type().Name(),
				Default: val.Interface(),
			}
			ps[kv.String()] = p
		}
		return ps

	case reflect.Struct:
		num := t.NumField()
		for i := 0; i < num; i++ {
			field := t.Field(i)
			p := &Property{
				Type:   build(field.Type),
				Format: field.Type.Name(),
			}
			fv := v.Field(i)
			ft := field.Type
			if fv.Kind() == reflect.Ptr {
				fv = fv.Elem()
				ft = ft.Elem()
			}
			if fv.Interface() == nil {
				fv = reflect.New(ft).Elem()
			}
			p.Default = fv.Interface()
			ps[field.Name] = p
		}
		return ps

	}

	return nil
}

func definitions(upath, pname, method string, format interface{}) (ref string) {
	ref = strings.Replace(path.Join(upath[1:], pname, method), "/", "@", -1)
	def := &Definition{
		Type: "object",
		Xml:  &Xml{Name: ref},
	}
	def.Properties = properties(format)
	if apidoc.Definitions == nil {
		apidoc.Definitions = map[string]*Definition{}
	}
	apidoc.Definitions[ref] = def
	return
}

func createPath(vr *lessgo.VirtRouter) string {
	u := vr.Path()
	a := strings.Split(u, "/*")
	if len(a) > 1 {
		u = path.Join(a[0], "{static}")
	}
	s := strings.Split(u, "/:")
	p := s[0]
	if len(s) == 1 {
		return p
	}
	for _, param := range s[1:] {
		p = path.Join(p, "{"+param+"}")
	}
	return p
}

func tagDesc(desc string) string {
	return strings.TrimSpace(desc)
}

// 操作摘要
func summary(desc string) string {
	return strings.TrimSpace(strings.Split(strings.TrimSpace(desc), "\n")[0])
}

// 操作描述
func desc(vr *lessgo.VirtRouter) string {
	var desc = new(string)
	middlewareDesc(desc, vr)
	var desc2 string
	var idx int
	for i, s := range strings.Split(*desc, "\n\n[路由中间件 ") {
		if i > 0 {
			idx++
			desc2 += "\n\n[路由中间件 " + strconv.Itoa(idx) + s
		}
	}
	return "<pre style=\"line-height:18px;\">" + strings.TrimSpace(vr.Description()) + desc2 + "</pre>"
}

// 递归获取相关中间件描述
func middlewareDesc(desc *string, vr *lessgo.VirtRouter) {
	for i := len(vr.Middlewares) - 1; i >= 0; i-- {
		m := vr.Middlewares[i]
		*desc = "\n\n[路由中间件 ] " + m.Name + ":\n" + m.GetApiMiddleware().Desc + *desc
	}
	if vr.Parent != nil {
		middlewareDesc(desc, vr.Parent)
	}
}

// 获取切片参数值的信息
func slice(value interface{}) (subtyp string, first interface{}) {
	subtyp = fmt.Sprintf("%T", value)
	idx := strings.Index(subtyp, "]")
	subtyp = subtyp[idx+1:]
	if strings.HasPrefix(subtyp, "[]") {
		subtyp = "array"
	} else {
		subtyp = mapping2[subtyp]
		if len(subtyp) == 0 {
			subtyp = "object"
		}
	}
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Len() > 0 {
		first = rv.Index(0).Interface()
	}
	return
}

// 获取参数值类型
func build(value interface{}) string {
	if value == nil {
		return "file"
	}
	rv, ok := value.(reflect.Type)
	if !ok {
		rv = reflect.TypeOf(value)
	}
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	return mapping[rv.Kind()]
}

// github.com/mcuadros/go-jsonschema-generator
var mapping = map[reflect.Kind]string{
	reflect.Bool:    "bool",
	reflect.Int:     "integer",
	reflect.Int8:    "integer",
	reflect.Int16:   "integer",
	reflect.Int32:   "integer",
	reflect.Int64:   "integer",
	reflect.Uint:    "integer",
	reflect.Uint8:   "integer",
	reflect.Uint16:  "integer",
	reflect.Uint32:  "integer",
	reflect.Uint64:  "integer",
	reflect.Float32: "number",
	reflect.Float64: "number",
	reflect.String:  "string",
	reflect.Slice:   "array",
	reflect.Struct:  "object",
	reflect.Map:     "object",
}

var mapping2 = map[string]string{
	"bool":    "bool",
	"int":     "integer",
	"int8":    "integer",
	"int16":   "integer",
	"int32":   "integer",
	"int64":   "integer",
	"uint":    "integer",
	"uint8":   "integer",
	"uint16":  "integer",
	"uint32":  "integer",
	"uint64":  "integer",
	"float32": "number",
	"float64": "number",
	"string":  "string",
}

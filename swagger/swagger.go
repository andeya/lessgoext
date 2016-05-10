package swagger

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/lessgo/lessgo"
	"github.com/lessgo/lessgo/utils"
	"github.com/lessgo/lessgoext/middleware"
	"github.com/lessgo/lessgoext/tools/copyfiles"
)

/*
 * API自动化文档
 * 仅限局域网访问
 */

var (
	apidoc *Swagger
	once   = true
	scheme = func() string {
		if lessgo.AppConfig.Listen.EnableHTTPS {
			return "https"
		} else {
			return "http"
		}
	}()
	jsonUrl       = "/swagger.json"
	dstSwagger    = "./SystemView/Swagger"
	swaggerHandle = &lessgo.ApiHandler{
		Desc:    "swagger",
		Methods: []string{"GET"},
		Handler: func(c lessgo.Context) error {
			if once {
				apidoc.Host = c.Request().Host()
				once = false
			}
			return c.JSON(200, apidoc)
		},
	}
	apidocHandle = &lessgo.ApiHandler{
		Desc:    "apidoc",
		Methods: []string{"GET"},
		Handler: func(c lessgo.Context) error {
			if c.Request().URL().Path() == "/apidoc" {
				return c.Redirect(302, "/apidoc/index.html")
			}
			return c.File(path.Join(dstSwagger, c.P(0)))
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
)

func Init() {
	// 注册路由
	lessgo.Root(
		lessgo.Leaf(jsonUrl, swaggerHandle, middleware.OnlyLANAccessWare),
		lessgo.Leaf("/apidoc*", apidocHandle, middleware.OnlyLANAccessWare),
	)

	lessgo.Logger().Sys(`Swagger API doc has been enabled, please access "/apidoc/index.html".` + "\n")

	// 拷贝swagger文件至当前目录下
	if !utils.FileExists(dstSwagger) {
		CopySwaggerFiles()
	}

	// 构建Swagger对象
	rootTag := &Tag{
		Name:        lessgo.RootRouter().Prefix,
		Description: lessgo.RootRouter().Description(),
	}
	// 生成swagger依赖的json对象
	apidoc = &Swagger{
		Version: SwaggerVersion,
		Info: &Info{
			Title:          lessgo.AppConfig.AppName + " API",
			Description:    lessgo.AppConfig.Info.Description,
			ApiVersion:     lessgo.AppConfig.Info.Version,
			Contact:        &Contact{Email: lessgo.AppConfig.Info.Email},
			TermsOfService: lessgo.AppConfig.Info.TermsOfServiceUrl,
			License: &License{
				Name: lessgo.AppConfig.Info.License,
				Url:  lessgo.AppConfig.Info.LicenseUrl,
			},
		},
		BasePath: "/",
		Tags:     []*Tag{rootTag},
		Schemes:  []string{scheme},
		Paths:    map[string]map[string]*Opera{},
		// SecurityDefinitions: map[string]map[string]interface{}{},
		// Definitions:         map[string]Definition{},
		// ExternalDocs:        map[string]string{},
	}

	for _, child := range lessgo.RootRouter().Children() {
		if child.Type == lessgo.HANDLER {
			addpath(child, rootTag)
			continue
		}
		childTag := &Tag{
			Name:        child.Prefix,
			Description: child.Description(),
		}
		apidoc.Tags = append(apidoc.Tags, childTag)
		for _, grandson := range child.Children() {
			if grandson.Type == lessgo.HANDLER {
				addpath(grandson, childTag)
				continue
			}
			grandsonTag := &Tag{
				Name:        child.Prefix + grandson.Prefix,
				Description: grandson.Description(),
			}
			apidoc.Tags = append(apidoc.Tags, grandsonTag)
			for _, vr := range grandson.Progeny() {
				if vr.Type != lessgo.HANDLER {
					continue
				}
				addpath(vr, grandsonTag)
			}
		}
	}
}

func addpath(vr *lessgo.VirtRouter, tag *Tag) {
	operas := map[string]*Opera{}
	pid := createPath(vr)

	for _, method := range vr.Methods() {
		if method == lessgo.CONNECT || method == lessgo.TRACE {
			continue
		}
		o := &Opera{
			Tags:        []string{tag.Name},
			Summary:     vr.Description(),
			Description: vr.Description(),
			OperationId: vr.Id,
			Produces:    []string{"application/xml", "application/json"},
			Responses: map[string]*Resp{
				"200": {Description: "Successful operation"},
				"400": {Description: "Invalid status value"},
				"404": {Description: "Not found"},
			},
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
			typ := build(param.Format)
			if typ == "object" {
				ref := strings.Replace(vr.Path()[1:]+param.Name, "/", "__", -1)
				p.Schema = &Schema{
					Ref: "#/definitions/" + ref,
				}
				def := &Definition{
					Type: typ,
					Xml:  &Xml{Name: ref},
				}
				def.Properties = properties(param.Format)
				if apidoc.Definitions == nil {
					apidoc.Definitions = map[string]*Definition{}
				}
				apidoc.Definitions[ref] = def
			} else {
				p.Type = typ
				p.Format = fmt.Sprintf("%T", param.Format)
				p.Default = param.Format
			}
			o.Parameters = append(o.Parameters, p)
		}

		// 静态目录路由
		if strings.HasSuffix(pid, "/{static}") {
			o.Parameters = append(o.Parameters, staticParam)
		}

		o.SetConsumes()
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
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	ps := map[string]*Property{}
	if v.Kind() == reflect.Map {
		kvs := v.MapKeys()
		for _, kv := range kvs {
			val := v.MapIndex(kv).Interface()
			p := &Property{
				Type:    build(val),
				Format:  fmt.Sprintf("%T", val),
				Default: val,
			}
			ps[kv.String()] = p
		}
		return ps
	}
	if v.Kind() == reflect.Struct {
		num := v.NumField()
		for i := 0; i < num; i++ {
			val := v.Field(i).Interface()
			p := &Property{
				Type:    build(val),
				Format:  fmt.Sprintf("%T", val),
				Default: val,
			}
			ps[v.Type().Field(i).Name] = p
		}
		return ps
	}
	return nil
}

func createPath(vr *lessgo.VirtRouter) string {
	u := vr.Path()
	if strings.HasSuffix(u, "*") {
		u = strings.TrimSuffix(u, "*")
		u = path.Join(u, "{static}")
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

type FileInfo struct {
	RelPath string
	Size    int64
	IsDir   bool
	Handle  *os.File
}

// 拷贝swagger中所有文件至dstSwagger下
func CopySwaggerFiles() {
	fp := filepath.Clean(filepath.Join(os.Getenv("GOPATH"), `\src\github.com\lessgo\lessgoext\swagger\swagger-ui`))
	copyfiles.CopyFiles(fp, dstSwagger, "", copyFunc)
}

//复制文件操作函数
func copyFunc(srcHandle, dstHandle *os.File) (err error) {
	stat, _ := srcHandle.Stat()
	if stat.Name() == "index.html" {
		b, err := ioutil.ReadAll(srcHandle)
		if err != nil {
			return err
		}
		b = bytes.Replace(b, []byte("{{{JSON_URL}}}"), []byte(jsonUrl), -1)
		_, err = io.Copy(dstHandle, bytes.NewBuffer(b))
		return err
	}
	_, err = io.Copy(dstHandle, srcHandle)
	return err
}

func build(value interface{}) string {
	if value == nil {
		return ""
	}
	rv := reflect.ValueOf(value)
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

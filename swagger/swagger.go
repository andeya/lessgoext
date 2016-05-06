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
)

/*
 * API自动化文档
 * 仅限局域网访问
 */

var (
	apidoc *Swagger
	scheme = func() string {
		if lessgo.AppConfig.Listen.EnableHTTPS {
			return "https"
		} else {
			return "http"
		}
	}()
	jsonUrl       = scheme + "://" + path.Join(lessgo.Lessgo().AppConfig.Info.Host, "swagger.json")
	swaggerHandle = &lessgo.ApiHandler{
		Desc:    "swagger",
		Methods: []string{"GET"},
		Handler: func(c lessgo.Context) error {
			return c.JSON(200, apidoc)
		},
	}
	apidocHandle = &lessgo.ApiHandler{
		Desc:    "apidoc",
		Methods: []string{"GET"},
		Handler: func(c lessgo.Context) error {
			return c.File(path.Join("Swagger", c.P(0)))
		},
	}
)

func Init() {
	// if !lessgo.AppConfig.CrossDomain {
	// 	lessgo.Logger().Warn("If you want to use swagger, please set crossdomain to true.")
	// }
	lessgo.Root(
		lessgo.Leaf("/swagger.json", swaggerHandle, middleware.OnlyLANAccessWare),
		lessgo.Leaf("/apidoc*", apidocHandle, middleware.OnlyLANAccessWare),
	)

	lessgo.Logger().Sys(`Swagger API doc has been enabled, please access "/apidoc/index.html".` + "\n")

	if !utils.FileExists("Swagger") {
		// 拷贝swagger文件至当前目录下
		CopySwaggerFiles()
	}
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
		Host:     lessgo.AppConfig.Info.Host,
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
		tag := &Tag{
			Name:        child.Prefix,
			Description: child.Description(),
		}
		apidoc.Tags = append(apidoc.Tags, tag)
		for _, vr := range child.Progeny() {
			if vr.Type != lessgo.HANDLER {
				continue
			}
			addpath(vr, tag)
		}
	}
}

func addpath(vr *lessgo.VirtRouter, tag *Tag) {
	operas := map[string]*Opera{}
	for _, method := range vr.Methods() {
		if method == lessgo.CONNECT || method == lessgo.TRACE {
			continue
		}
		o := &Opera{
			Tags:        []string{tag.Name},
			Summary:     vr.Description(),
			Description: vr.Description(),
			OperationId: vr.Id,
			// Consumes:    vr.Produces(),
			// Produces:    vr.Produces(),

			// Parameters:  []*Parameter{},
			Responses: map[string]*Resp{
				"200": {Description: "Successful operation"},
				"400": {Description: "Invalid status value"},
				"404": {Description: "Not found"},
			},
			// Security: []map[string][]string{},
		}
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
		operas[strings.ToLower(method)] = o
	}
	pid := createPath(vr)
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
	s := strings.Split(vr.Path(), "/:")
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

// 拷贝swagger文件至当前目录下
func CopySwaggerFiles() {
	files_ch := make(chan *FileInfo, 100)
	fp := filepath.Clean(filepath.Join(os.Getenv("GOPATH"), `\src\github.com\lessgo\lessgoext\swagger\swagger-ui`))
	go walkFiles(fp, "", files_ch) //在一个独立的 goroutine 中遍历文件
	os.MkdirAll("Swagger", os.ModeDir)
	writeFiles("Swagger", files_ch)
}

//遍历目录，将文件信息传入通道
func walkFiles(srcDir, suffix string, c chan<- *FileInfo) {
	suffix = strings.ToUpper(suffix)
	filepath.Walk(srcDir, func(f string, fi os.FileInfo, err error) error { //遍历目录
		if err != nil {
			lessgo.Logger().Error("%v", err)
			return err
		}
		fileInfo := &FileInfo{}
		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) { //匹配文件
			if fh, err := os.OpenFile(f, os.O_RDONLY, os.ModePerm); err != nil {
				lessgo.Logger().Error("%v", err)
				return err
			} else {
				fileInfo.Handle = fh
				fileInfo.RelPath, _ = filepath.Rel(srcDir, f) //相对路径
				fileInfo.Size = fi.Size()
				fileInfo.IsDir = fi.IsDir()
			}
			c <- fileInfo
		}
		return nil
	})
	close(c) //遍历完成，关闭通道
}

//写目标文件
func writeFiles(dstDir string, c <-chan *FileInfo) {
	if err := os.Chdir(dstDir); err != nil { //切换工作路径
		lessgo.Logger().Fatal("%v", err)
	}
	for f := range c {
		if fi, err := os.Stat(f.RelPath); os.IsNotExist(err) { //目标不存在
			if f.IsDir {
				if err := os.MkdirAll(f.RelPath, os.ModeDir); err != nil {
					lessgo.Logger().Error("%v", err)
				}
			} else {
				if err := ioCopy(f.Handle, f.RelPath); err != nil {
					lessgo.Logger().Error("%v", err)
				} else {
					lessgo.Logger().Info("CP: %v", f.RelPath)
				}
			}
		} else if !f.IsDir { //目标存在，而且源不是一个目录
			if fi.IsDir() != f.IsDir { //检查文件名被目录名占用冲突
				lessgo.Logger().Error("%v", "filename conflict:", f.RelPath)
			} else if fi.Size() != f.Size { //源和目标的大小不一致时才重写
				if err := ioCopy(f.Handle, f.RelPath); err != nil {
					lessgo.Logger().Error("%v", err)
				} else {
					lessgo.Logger().Info("CP: %v", f.RelPath)
				}
			}
		}
	}
	os.Chdir("../")
}

//复制文件数据
func ioCopy(srcHandle *os.File, dstPth string) (err error) {
	defer srcHandle.Close()
	dstHandle, err := os.OpenFile(dstPth, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer dstHandle.Close()

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

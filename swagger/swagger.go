package swagger

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/lessgo/lessgo"
)

var (
	apidoc *Swagger
)

func Init() {
	lessgo.DefLessgo.AppConfig.CrossDomain = true
	lessgo.MustRouter("/swagger.json", []string{lessgo.GET}, func(c lessgo.Context) error {
		return c.JSON(200, apidoc)
	})

	files_ch := make(chan *FileInfo, 100)
	fp := filepath.Clean(filepath.Join(os.Getenv("GOPATH"), `\src\github.com\lessgo\lessgoext\swagger\swagger-ui`))
	go walkFiles(fp, "", files_ch) //在一个独立的 goroutine 中遍历文件
	os.MkdirAll("Swagger", os.ModeDir)
	writeFiles("Swagger", files_ch)

	var scheme string
	if lessgo.AppConfig.Listen.EnableHTTPS {
		scheme = "https"
	} else {
		scheme = "http"
	}
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
		Tags: []*Tag{
			{
				Name:        "/:",
				Description: "",
			},
		},
		Schemes: []string{scheme},
		Paths:   map[string]map[string]*Opera{},
		// SecurityDefinitions: map[string]map[string]interface{}{},
		// Definitions:         map[string]Definition{},
		// ExternalDocs:        map[string]string{},
	}
	for _, vr := range lessgo.DefLessgo.VirtRouter.Progeny() {
		if vr.Type() != lessgo.HANDLER {
			continue
		}
		opera := map[string]*Opera{}
		for _, method := range vr.Methods() {
			if method == lessgo.CONNECT || method == lessgo.TRACE {
				continue
			}
			o := &Opera{
				Tags:        []string{"/:"},
				Summary:     vr.Name(),
				Description: vr.Description(),
				OperationId: vr.Id(),
				Consumes:    vr.Produces(),
				Produces:    vr.Produces(),
				// Parameters:  []Parameter{},
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
					Type:        build(param.Format),
					// Items:       &Items{},
					// Schema:      &Schema{},
					Format: fmt.Sprintf("%T", param.Format),
				}
				o.Parameters = append(o.Parameters, p)
			}
			opera[strings.ToLower(method)] = o
		}
		apidoc.Paths[createPath(vr)] = opera
	}
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

//复制文件数据
func ioCopy(srcHandle *os.File, dstPth string) (err error) {
	dstHandle, err := os.OpenFile(dstPth, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer srcHandle.Close()
	defer dstHandle.Close()
	_, err = io.Copy(dstHandle, srcHandle)
	return err
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
		f.Handle.Close()
	}
	os.Chdir("../")
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

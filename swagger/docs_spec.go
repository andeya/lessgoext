// Package swagger struct definition
package swagger

/*
 * 关于swagger的说明
 * 一、数据结构主要用于固定格式的服务器响应结构，适用于多个接口可能返回相同的数据结构，编辑保存后相关所有的引用都会变更。
 * 支持的数据类型说明如下：
 * 1、string:字符串类型
 * 2、array:数组类型，子项只能是支持的数据类型中的一种，不能添加多个
 * 3、object:对象类型，只支持一级属性，不支持嵌套，嵌套可以通过在属性中引入ref类型的对象或自定义数据格式
 * 4、int:短整型
 * 5、long:长整型
 * 6、float:浮点型
 * 7、double:浮点型
 * 8、decimal:精确到比较高的浮点型
 * 9、ref:引用类型，即引用定义好的数据结构
 *
 * 二、参数位置
 *    body：http请求body
 *    cookie：本地cookie
 *    formData：表单参数
 *    header：http请求header
 *    path：http请求url,如getInfo/{userId}
 *    query：http请求拼接，如getInfo?userId={userId}
 * 三、参数类型
 *    自定义：目前仅支持自定义json格式，仅当"参数位置"为“body"有效
 */

// SwaggerVersion show the current swagger version
const SwaggerVersion = "2.0"

type (
	Swagger struct {
		Version             string                            `json:"swagger"`
		Info                *Info                             `json:"info"`
		Host                string                            `json:"host"`
		BasePath            string                            `json:"basePath"`
		Tags                []*Tag                            `json:"tags"`
		Schemes             []string                          `json:"schemes"`
		Paths               map[string]map[string]*Opera      `json:"paths,omitempty"` // {"前缀":{"方法":{...}}}
		SecurityDefinitions map[string]map[string]interface{} `json:"securityDefinitions,omitempty"`
		Definitions         map[string]*Definition            `json:"definitions,omitempty"`
		ExternalDocs        map[string]string                 `json:"externalDocs,omitempty"`
	}
	Info struct {
		Title          string   `json:"title"`
		Description    string   `json:"description"`
		ApiVersion     string   `json:"version"`
		Contact        *Contact `json:"contact"`
		TermsOfService string   `json:"termsOfService"`
		License        *License `json:"license,omitempty"`
	}
	Contact struct {
		Email string `json:"email,omitempty"`
	}
	License struct {
		Name string `json:"name"`
		Url  string `json:"url"`
	}
	Tag struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	Opera struct {
		Tags        []string              `json:"tags"`
		Summary     string                `json:"summary"`
		Description string                `json:"description"`
		OperationId string                `json:"operationId"`
		Consumes    []string              `json:"consumes,omitempty"`
		Produces    []string              `json:"produces,omitempty"`
		Parameters  []*Parameter          `json:"parameters,omitempty"`
		Responses   map[string]*Resp      `json:"responses,omitempty"` // {"httpcode":resp}
		Security    []map[string][]string `json:"security,omitempty"`
	}
	Parameter struct {
		In               string      `json:"in"` // 参数位置
		Name             string      `json:"name"`
		Description      string      `json:"description"`
		Required         bool        `json:"required"`
		Type             string      `json:"type,omitempty"` // "array"|"integer"|"object"
		Items            *Items      `json:"items,omitempty"`
		Schema           *Schema     `json:"schema,omitempty"`
		CollectionFormat string      `json:"collectionFormat,omitempty"` // "multi"
		Format           string      `json:"format,omitempty"`           // "int64"
		Default          interface{} `json:"default,omitempty"`
	}
	Items struct {
		Type    string      `json:"type"` // "string"
		Enum    []string    `json:"enum,omitempty"`
		Default interface{} `json:"default,omitempty"`
	}
	Schema struct {
		Ref                  string            `json:"$ref,omitempty"`
		Type                 string            `json:"type,omitempty"` // "array"|"integer"|"object"
		Items                *Items            `json:"items,omitempty"`
		Description          string            `json:"description,omitempty"`
		AdditionalProperties map[string]string `json:"additionalProperties,omitempty"`
	}
	Resp struct {
		Description string `json:"description"`
	}
	Definition struct {
		Type       string               `json:"type,omitempty"` // "object"
		Properties map[string]*Property `json:"properties,omitempty"`
		Xml        *Xml                 `json:"xml,omitempty"`
	}
	Property struct {
		Type        string      `json:"type,omitempty"`   // "array"|"integer"|"object"
		Format      string      `json:"format,omitempty"` // "int64"
		Description string      `json:"description,omitempty"`
		Enum        []string    `json:"enum,omitempty"`
		Example     interface{} `json:"example,omitempty"`
		Default     interface{} `json:"default,omitempty"`
	}
	Xml struct {
		Name    string `json:"name"`
		Wrapped bool   `json:"wrapped,omitempty"`
	}
)

var CommonMIMETypes = []string{
	"application/json",
	"application/javascript",
	"application/xml",
	"application/x-www-form-urlencoded",
	"application/protobuf",
	"application/msgpack",
	"text/html",
	"text/plain",
	"multipart/form-data",
	"application/octet-stream",
}

// Package swagger struct definition
package swagger

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
		In               string      `json:"in"`
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
		Properties map[string]*Property `json:"Properties,omitempty"`
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

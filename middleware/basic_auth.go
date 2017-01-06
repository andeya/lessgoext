package middleware

import (
	"encoding/base64"

	"github.com/henrylee2cn/lessgo"
)

type (
	// BasicAuthConfig defines the config for HTTP basic auth middleware.
	BasicAuthConfig struct {
		// Validator is a function to validate basic auth credentials.
		Validator BasicAuthValidator
	}

	// BasicAuthValidator defines a function to validate basic auth credentials.
	BasicAuthValidator func(string, string) bool
)

const (
	basic = "Basic"
)

// BasicAuth returns an HTTP basic auth middleware.
//
// For valid credentials it calls the next handler.
// For invalid credentials, it sends "401 - Unauthorized" response.
// For empty or invalid `Authorization` header, it sends "400 - Bad Request" response.
var BasicAuth = lessgo.ApiMiddleware{
	Name:   "BasicAuthWithConfig",
	Desc:   `基本的第三方授权中间件，使用前请先在源码配置处理函数。`,
	Config: nil,
	Middleware: func(confObject interface{}) lessgo.MiddlewareFunc {
		config := BasicAuthConfig{confObject.(BasicAuthValidator)}
		return func(next lessgo.HandlerFunc) lessgo.HandlerFunc {
			return func(c *lessgo.Context) error {
				auth := c.HeaderParam(lessgo.HeaderAuthorization)
				l := len(basic)

				if len(auth) > l+1 && auth[:l] == basic {
					b, err := base64.StdEncoding.DecodeString(auth[l+1:])
					if err != nil {
						return err
					}
					cred := string(b)
					for i := 0; i < len(cred); i++ {
						if cred[i] == ':' {
							// Verify credentials
							if config.Validator(cred[:i], cred[i+1:]) {
								return next(c)
							}
						}
					}
				}
				// Need to return `401` for browsers to pop-up login box.
				c.Response().Header().Set(lessgo.HeaderWWWAuthenticate, basic+" realm=Restricted")
				return lessgo.ErrUnauthorized
			}
		}
	},
}.Reg()

package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/henrylee2cn/lessgo"
)

type (
	// CORSConfig defines the config for CORS middleware.
	CORSConfig struct {
		// AllowOrigin defines a list of origins that may access the resource.
		// Optional. Default value []string{"*"}.
		AllowOrigins []string `json:"allow_origins"`

		// AllowMethods defines a list methods allowed when accessing the resource.
		// This is used in response to a preflight request.
		// Optional. Default value DefaultCORSConfig.AllowMethods.
		AllowMethods []string `json:"allow_methods"`

		// AllowHeaders defines a list of request headers that can be used when
		// making the actual request. This in response to a preflight request.
		// Optional. Default value []string{}.
		AllowHeaders []string `json:"allow_headers"`

		// AllowCredentials indicates whether or not the response to the request
		// can be exposed when the credentials flag is true. When used as part of
		// a response to a preflight request, this indicates whether or not the
		// actual request can be made using credentials.
		// Optional. Default value false.
		AllowCredentials bool `json:"allow_credentials"`

		// ExposeHeaders defines a whitelist headers that clients are allowed to
		// access.
		// Optional. Default value []string{}.
		ExposeHeaders []string `json:"expose_headers"`

		// MaxAge indicates how long (in seconds) the results of a preflight request
		// can be cached.
		// Optional. Default value 0.
		MaxAge int `json:"max_age"`
	}
)

var (
	// DefaultCORSConfig is the default CORS middleware config.
	DefaultCORSConfig = CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{lessgo.GET, lessgo.HEAD, lessgo.PUT, lessgo.PATCH, lessgo.POST, lessgo.DELETE},
	}
)

// CORS returns a Cross-Origin Resource Sharing (CORS) middleware.
// See https://developer.mozilla.org/en/docs/Web/HTTP/Access_control_CORS
var CORS = lessgo.ApiMiddleware{
	Name: "CORS",
	Desc: `a Cross-Origin Resource Sharing (CORS) middleware.
See https://developer.mozilla.org/en/docs/Web/HTTP/Access_control_CORS`,
	Config: DefaultCORSConfig,
	Middleware: func(confObject interface{}) lessgo.MiddlewareFunc {
		config := confObject.(CORSConfig)
		// Defaults
		if len(config.AllowOrigins) == 0 {
			config.AllowOrigins = DefaultCORSConfig.AllowOrigins
		}
		if len(config.AllowMethods) == 0 {
			config.AllowMethods = DefaultCORSConfig.AllowMethods
		}
		allowMethods := strings.Join(config.AllowMethods, ",")
		allowHeaders := strings.Join(config.AllowHeaders, ",")
		exposeHeaders := strings.Join(config.ExposeHeaders, ",")
		maxAge := strconv.Itoa(config.MaxAge)

		return func(next lessgo.HandlerFunc) lessgo.HandlerFunc {
			return func(c *lessgo.Context) error {
				req := c.Request()
				res := c.Response()
				header := res.Header()
				origin := req.Header.Get(lessgo.HeaderOrigin)
				_, originSet := req.Header[lessgo.HeaderOrigin]

				// Check allowed origins
				allowedOrigin := ""
				for _, o := range config.AllowOrigins {
					if o == "*" || o == origin {
						allowedOrigin = o
						break
					}
				}

				// Simple request
				if req.Method != lessgo.OPTIONS {
					header.Add(lessgo.HeaderVary, lessgo.HeaderOrigin)
					if !originSet || allowedOrigin == "" {
						return next(c)
					}
					header.Set(lessgo.HeaderAccessControlAllowOrigin, allowedOrigin)
					if config.AllowCredentials {
						header.Set(lessgo.HeaderAccessControlAllowCredentials, "true")
					}
					if exposeHeaders != "" {
						header.Set(lessgo.HeaderAccessControlExposeHeaders, exposeHeaders)
					}
					return next(c)
				}

				// Preflight request
				header.Add(lessgo.HeaderVary, lessgo.HeaderOrigin)
				header.Add(lessgo.HeaderVary, lessgo.HeaderAccessControlRequestMethod)
				header.Add(lessgo.HeaderVary, lessgo.HeaderAccessControlRequestHeaders)
				if !originSet || allowedOrigin == "" {
					return next(c)
				}
				header.Set(lessgo.HeaderAccessControlAllowOrigin, allowedOrigin)
				header.Set(lessgo.HeaderAccessControlAllowMethods, allowMethods)
				if config.AllowCredentials {
					header.Set(lessgo.HeaderAccessControlAllowCredentials, "true")
				}
				if allowHeaders != "" {
					header.Set(lessgo.HeaderAccessControlAllowHeaders, allowHeaders)
				} else {
					h := req.Header.Get(lessgo.HeaderAccessControlRequestHeaders)
					if h != "" {
						header.Set(lessgo.HeaderAccessControlAllowHeaders, h)
					}
				}
				if config.MaxAge > 0 {
					header.Set(lessgo.HeaderAccessControlMaxAge, maxAge)
				}
				return c.NoContent(http.StatusNoContent)
			}
		}
	},
}.Reg()

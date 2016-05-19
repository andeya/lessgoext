package middleware

import (
	"github.com/lessgo/lessgo"
)

type (
	// MethodOverrideConfig defines the config for method override middleware.
	MethodOverrideConfig struct {
		// Getter is a function that gets overridden method from the request.
		Getter MethodOverrideGetter
	}

	// MethodOverrideGetter is a function that gets overridden method from the request
	// Optional, with default values as `MethodFromHeader(lessgo.HeaderXHTTPMethodOverride)`.
	MethodOverrideGetter func(lessgo.Context) string
)

var (
	// DefaultMethodOverrideConfig is the default method override middleware config.
	DefaultMethodOverrideConfig = MethodOverrideConfig{
		Getter: MethodFromHeader(lessgo.HeaderXHTTPMethodOverride),
	}
)

// MethodOverride returns a method override middleware.
// Method override middleware checks for the overridden method from the request and
// uses it instead of the original method.
//
// For security reasons, only `POST` method can be overridden.
var MethodOverride = lessgo.ApiMiddleware{
	Name:   "MethodOverride",
	Desc:   `Checks for the overridden method from the request and uses it instead of the original method.`,
	Config: DefaultMethodOverrideConfig,
	Middleware: func(confObject interface{}) lessgo.MiddlewareFunc {
		config := confObject.(MethodOverrideConfig)
		// Defaults
		if config.Getter == nil {
			config.Getter = DefaultMethodOverrideConfig.Getter
		}

		return func(next lessgo.HandlerFunc) lessgo.HandlerFunc {
			return func(c lessgo.Context) error {
				req := c.Request()
				if req.Method == lessgo.POST {
					m := config.Getter(c)
					if m != "" {
						req.Method = m
					}
				}
				return next(c)
			}
		}
	},
}.Reg()

// MethodFromHeader is a `MethodOverrideGetter` that gets overridden method from
// the request header.
func MethodFromHeader(header string) MethodOverrideGetter {
	return func(c lessgo.Context) string {
		return c.Request().Header.Get(header)
	}
}

// MethodFromForm is a `MethodOverrideGetter` that gets overridden method from the
// form parameter.
func MethodFromForm(param string) MethodOverrideGetter {
	return func(c lessgo.Context) string {
		return c.FormValue(param)
	}
}

// MethodFromQuery is a `MethodOverrideGetter` that gets overridden method from
// the query parameter.
func MethodFromQuery(param string) MethodOverrideGetter {
	return func(c lessgo.Context) string {
		return c.QueryParam(param)
	}
}

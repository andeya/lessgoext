package middleware

import (
	"github.com/lessgo/lessgo"
)

type (
	// TrailingSlashConfig defines the config for TrailingSlash middleware.
	TrailingSlashConfig struct {
		// RedirectCode is the status code used when redirecting the request.
		// Optional, but when provided the request is redirected using this code.
		RedirectCode int `json:"redirect_code"`
	}
)

// AddTrailingSlash returns a root level (before router) middleware which adds a
// trailing slash to the request `URL#Path`.
var AddTrailingSlash = lessgo.ApiMiddleware{
	Name:   "TrailingSlash",
	Desc:   "a root level (before router) middleware which adds a trailing slash to the request `URL#Path`.",
	Config: TrailingSlashConfig{0},
	Middleware: func(confObject interface{}) lessgo.MiddlewareFunc {
		config := confObject.(TrailingSlashConfig)
		return func(next lessgo.HandlerFunc) lessgo.HandlerFunc {
			return func(c lessgo.Context) error {
				req := c.Request()
				url := req.URL
				path := url.Path
				qs := url.RawQuery
				if path != "/" && path[len(path)-1] != '/' {
					path += "/"
					uri := path
					if qs != "" {
						uri += "?" + qs
					}
					if config.RedirectCode != 0 {
						return c.Redirect(config.RedirectCode, uri)
					}
					req.RequestURI = uri
					url.Path = path
				}
				return next(c)
			}
		}
	},
}.Reg()

// RemoveTrailingSlash returns a root level (before router) middleware which removes
// a trailing slash from the request URI.
//
// Usage `Echo#Pre(RemoveTrailingSlash())`
func RemoveTrailingSlash() lessgo.MiddlewareFunc {
	return RemoveTrailingSlashWithConfig(TrailingSlashConfig{})
}

// RemoveTrailingSlashWithConfig returns a RemoveTrailingSlash middleware from config.
// See `RemoveTrailingSlash()`.
func RemoveTrailingSlashWithConfig(config TrailingSlashConfig) lessgo.MiddlewareFunc {
	return func(next lessgo.HandlerFunc) lessgo.HandlerFunc {
		return func(c lessgo.Context) error {
			req := c.Request()
			url := req.URL
			path := url.Path
			qs := url.RawQuery
			l := len(path) - 1
			if path != "/" && path[l] == '/' {
				path = path[:l]
				uri := path
				if qs != "" {
					uri += "?" + qs
				}
				if config.RedirectCode != 0 {
					return c.Redirect(config.RedirectCode, uri)
				}
				req.RequestURI = uri
				url.Path = path
			}
			return next(c)
		}
	}
}

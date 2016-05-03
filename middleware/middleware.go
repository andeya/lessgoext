package middleware

import (
	"github.com/lessgo/lessgo"
)

var (
	AddTrailingSlashWare = (&lessgo.ApiMiddleware{
		Name:       "TrailingSlash",
		Desc:       "a root level (before router) middleware which adds a trailing slash to the request `URL#Path`.",
		Middleware: AddTrailingSlash,
	}).Init()

	StaticWare = (&lessgo.ApiMiddleware{
		Name:          "Static",
		Desc:          "a static middleware to serves static content from the provided root directory.",
		DefaultConfig: DefaultStaticConfig,
		Middleware:    Static,
	}).Init()

	BasicAuthWare = (&lessgo.ApiMiddleware{
		Name:          "BasicAuth",
		Desc:          "an HTTP basic auth middleware from config.",
		DefaultConfig: DefaultBasicAuthConfig,
		Middleware:    BasicAuth,
	}).Init()

	JWTAuthWare = (&lessgo.ApiMiddleware{
		Name: "JWTAuth",
		Desc: `a JSON Web Token (JWT) auth middleware.
For valid token, it sets the user in context and calls next handler.
For invalid token, it sends "401 - Unauthorized" response.
For empty or invalid 'Authorization' header, it sends "400 - Bad Request".`,
		DefaultConfig: DefaultJWTAuthConfig,
		Middleware:    JWTAuth,
	}).Init()

	GzipWare = (&lessgo.ApiMiddleware{
		Name:          "Gzip",
		Desc:          `a middleware which compresses HTTP response using gzip compression scheme.`,
		DefaultConfig: DefaultGzipConfig,
		Middleware:    Gzip,
	}).Init()

	CORSWare = (&lessgo.ApiMiddleware{
		Name: "CORS",
		Desc: `a Cross-Origin Resource Sharing (CORS) middleware.
See https://developer.mozilla.org/en/docs/Web/HTTP/Access_control_CORS`,
		DefaultConfig: DefaultGzipConfig,
		Middleware:    CORS,
	}).Init()
)

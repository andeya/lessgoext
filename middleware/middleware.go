package middleware

import (
	"github.com/lessgo/lessgo"
)

var (
	AddTrailingSlashWare = lessgo.ApiMiddleware{
		Name:       "TrailingSlash",
		Desc:       "a root level (before router) middleware which adds a trailing slash to the request `URL#Path`.",
		Middleware: AddTrailingSlash,
	}.Reg()

	StaticWare = lessgo.ApiMiddleware{
		Name:          "Static",
		Desc:          "a static middleware to serves static content from the provided root directory.",
		DefaultConfig: DefaultStaticConfig,
		Middleware:    Static,
	}.Reg()

	BasicAuthWare = lessgo.ApiMiddleware{
		Name:          "BasicAuth",
		Desc:          "an HTTP basic auth middleware from config.",
		DefaultConfig: DefaultBasicAuthConfig,
		Middleware:    BasicAuth,
	}.Reg()

	JWTAuthWare = lessgo.ApiMiddleware{
		Name: "JWTAuth",
		Desc: `a JSON Web Token (JWT) auth middleware.
For valid token, it sets the user in context and calls next handler.
For invalid token, it sends "401 - Unauthorized" response.
For empty or invalid 'Authorization' header, it sends "400 - Bad Request".`,
		DefaultConfig: DefaultJWTAuthConfig,
		Middleware:    JWTAuth,
	}.Reg()

	GzipWare = lessgo.ApiMiddleware{
		Name:          "Gzip",
		Desc:          `a middleware which compresses HTTP response using gzip compression scheme.`,
		DefaultConfig: DefaultGzipConfig,
		Middleware:    Gzip,
	}.Reg()

	CORSWare = lessgo.ApiMiddleware{
		Name: "CORS",
		Desc: `a Cross-Origin Resource Sharing (CORS) middleware.
See https://developer.mozilla.org/en/docs/Web/HTTP/Access_control_CORS`,
		DefaultConfig: DefaultGzipConfig,
		Middleware:    CORS,
	}.Reg()

	MethodOverrideWare = lessgo.ApiMiddleware{
		Name:          "MethodOverride",
		Desc:          `Checks for the overridden method from the request and uses it instead of the original method.`,
		DefaultConfig: DefaultMethodOverrideConfig,
		Middleware:    MethodOverride,
	}.Reg()

	SecureWare = lessgo.ApiMiddleware{
		Name:          "Secure",
		Desc:          `Provides protection against cross-site scripting (XSS) attack, content type sniffing, clickjacking, insecure connection and other code injection attacks.`,
		DefaultConfig: DefaultSecureConfig,
		Middleware:    Secure,
	}.Reg()

	BodyLimitWare = lessgo.ApiMiddleware{
		Name: "BodyLimit",
		Desc: `sets the maximum allowed size for a request body, if the size exceeds the configured limit, it sends "413 - Request Entity Too Large" response.
The body limit is determined based on the actually read and not 'Content-Length' request header, which makes it super secure.
Limit can be specified as '4x' or '4xB', where x is one of the multiple from K, M, G, T or P.`,
		DefaultConfig: nil,
		Middleware:    BodyLimit,
	}.Reg()

	OnlyLANAccessWare = lessgo.ApiMiddleware{
		Name:          "OnlyLANAccess",
		Desc:          `Only allow LAN access.`,
		DefaultConfig: nil,
		Middleware:    OnlyLANAccess,
	}.Reg()
)

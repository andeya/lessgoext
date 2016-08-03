package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/lessgo/lessgo"
)

var (
	lanPrefix = []string{
		"::",
		"127.",
		"192.168.",
		"10.",
	}
)

// The IPs which are allowed access.
var AllowIPPrefixes = lessgo.ApiMiddleware{
	Name:   "AllowIPPrefixes",
	Desc:   `The IP Prefixes which are allowed access.`,
	Config: lanPrefix,
	Middleware: func(confObject interface{}) lessgo.MiddlewareFunc {
		ips, _ := confObject.([]string)
		return func(next lessgo.HandlerFunc) lessgo.HandlerFunc {
			return func(c *lessgo.Context) error {
				remoteAddress := c.RealRemoteAddr()
				for i, count := 0, len(ips); i < count; i++ {
					if strings.HasPrefix(remoteAddress, ips[i]) {
						return next(c)
					}
				}

				return c.Failure(http.StatusForbidden, errors.New(`Only allow LAN access, your ip is `+c.RealRemoteAddr()+`.`))
			}
		}
	},
}.Reg()

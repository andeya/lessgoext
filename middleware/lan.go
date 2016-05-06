package middleware

import (
	"bytes"
	"net/http"

	"github.com/lessgo/lessgo"
)

var (
	lanPrefix_1 = []byte("[")
	lanPrefix_2 = []byte("127.")
	lanPrefix_3 = []byte("192.168.")
	lanPrefix_4 = []byte("10.")
)

// Only allow LAN access.
func OnlyLANAccess(configJSON string) lessgo.MiddlewareFunc {
	return func(next lessgo.HandlerFunc) lessgo.HandlerFunc {
		return func(c lessgo.Context) error {
			req := c.Request()
			remoteAddr := req.RemoteAddress()
			if ip := req.Header().Get(lessgo.HeaderXRealIP); ip != "" {
				remoteAddr = ip
			} else if ip = req.Header().Get(lessgo.HeaderXForwardedFor); ip != "" {
				remoteAddr = ip
			}
			remoteAddress := []byte(remoteAddr)
			if bytes.HasPrefix(remoteAddress, lanPrefix_1) ||
				bytes.HasPrefix(remoteAddress, lanPrefix_2) ||
				bytes.HasPrefix(remoteAddress, lanPrefix_3) ||
				bytes.HasPrefix(remoteAddress, lanPrefix_4) {
				return next(c)
			}
			return lessgo.NewHTTPError(http.StatusForbidden, "Only allow LAN access.")
		}
	}
}

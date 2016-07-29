package middleware

import (
	"bytes"
	"net/http"

	"github.com/lessgo/lessgo"
)

var (
	lanPrefix_1 = []byte("::")
	lanPrefix_2 = []byte("127.")
	lanPrefix_3 = []byte("192.168.")
	lanPrefix_4 = []byte("10.")
)

// Only allow LAN access.
var OnlyLANAccess = lessgo.ApiMiddleware{
	Name: "OnlyLANAccess",
	Desc: `Only allow LAN access.`,
	Middleware: func(next lessgo.HandlerFunc) lessgo.HandlerFunc {
		return func(c *lessgo.Context) error {
			remoteAddress := []byte(c.RealRemoteAddr())
			if bytes.HasPrefix(remoteAddress, lanPrefix_1) ||
				bytes.HasPrefix(remoteAddress, lanPrefix_2) ||
				bytes.HasPrefix(remoteAddress, lanPrefix_3) ||
				bytes.HasPrefix(remoteAddress, lanPrefix_4) {
				return next(c)
			}

			s := `<html>
<head><title>403 Forbidden</title></head>
<body bgcolor="white">
<center><h1>403 Forbidden</h1></center>
<hr><center>Only allow LAN access, your ip is ` + c.RealRemoteAddr() + `.</center>
</body>
</html>
`
			return c.HTML(http.StatusForbidden, s)
		}
	},
}.Reg()

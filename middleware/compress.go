package middleware

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/lessgo/lessgo"
	"github.com/lessgo/lessgo/engine"
)

type (
	// GzipConfig defines the config for gzip middleware.
	GzipConfig struct {
		// Level is the gzip level.
		// Optional, with default value as -1.
		Level int
	}

	gzipResponseWriter struct {
		engine.Response
		io.Writer
	}
)

var (
	// DefaultGzipConfig is the default gzip middleware config.
	DefaultGzipConfig = GzipConfig{
		Level: -1,
	}
)

// Gzip returns a middleware which compresses HTTP response using gzip compression
// scheme.
func Gzip(configJSON string) lessgo.MiddlewareFunc {
	config := GzipConfig{}
	json.Unmarshal([]byte(configJSON), &config)
	// Defaults
	if config.Level == 0 {
		config.Level = DefaultGzipConfig.Level
	}

	pool := gzipPool(config)
	scheme := "gzip"

	return func(next lessgo.HandlerFunc) lessgo.HandlerFunc {
		return func(c lessgo.Context) error {
			res := c.Response()
			res.Header().Add(lessgo.HeaderVary, lessgo.HeaderAcceptEncoding)
			if strings.Contains(c.Request().Header().Get(lessgo.HeaderAcceptEncoding), scheme) {
				rw := res.Writer()
				gw := pool.Get().(*gzip.Writer)
				gw.Reset(rw)
				defer func() {
					if res.Size() == 0 {
						// We have to reset response to it's pristine state when
						// nothing is written to body or error is returned.
						// See issue #424, #407.
						res.SetWriter(rw)
						res.Header().Del(lessgo.HeaderContentEncoding)
						gw.Reset(ioutil.Discard)
					}
					gw.Close()
					pool.Put(gw)
				}()
				g := gzipResponseWriter{Response: res, Writer: gw}
				res.Header().Set(lessgo.HeaderContentEncoding, scheme)
				res.SetWriter(g)
			}
			return next(c)
		}
	}
}

func (g gzipResponseWriter) Write(b []byte) (int, error) {
	if g.Header().Get(lessgo.HeaderContentType) == "" {
		g.Header().Set(lessgo.HeaderContentType, http.DetectContentType(b))
	}
	return g.Writer.Write(b)
}

func gzipPool(config GzipConfig) sync.Pool {
	return sync.Pool{
		New: func() interface{} {
			w, _ := gzip.NewWriterLevel(ioutil.Discard, config.Level)
			return w
		},
	}
}

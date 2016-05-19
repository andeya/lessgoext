package middleware

import (
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/lessgo/lessgo"
	"github.com/lessgo/lessgo/utils/bytes"
)

type (
	// BodyLimitConfig defines the config for body limit middleware.
	BodyLimitConfig struct {
		// Limit is the maximum allowed size for a request body, it can be specified
		// as `4x` or `4xB`, where x is one of the multiple from K, M, G, T or P.
		Limit string `json:"limit"`
		limit int64
	}

	limitedReader struct {
		BodyLimitConfig
		reader  io.Reader
		read    int64
		context lessgo.Context
	}
)

// BodyLimit returns a body limit middleware.
//
// BodyLimit middleware sets the maximum allowed size for a request body, if the
// size exceeds the configured limit, it sends "413 - Request Entity Too Large"
// response. The body limit is determined based on the actually read and not `Content-Length`
// request header, which makes it super secure.
// Limit can be specified as `4x` or `4xB`, where x is one of the multiple from K, M,
// G, T or P.

var BodyLimit = lessgo.ApiMiddleware{
	Name: "BodyLimit",
	Desc: `sets the maximum allowed size for a request body, if the size exceeds the configured limit, it sends "413 - Request Entity Too Large" response.
The body limit is determined based on the actually read and not 'Content-Length' request header, which makes it super secure.
Limit can be specified as '4x' or '4xB', where x is one of the multiple from K, M, G, T or P.`,
	Config: BodyLimitConfig{},
	Middleware: func(confObject interface{}) lessgo.MiddlewareFunc {
		config := confObject.(BodyLimitConfig)
		limit, err := bytes.Parse(config.Limit)
		if err != nil {
			panic(fmt.Errorf("invalid body-limit=%s", config.Limit))
		}
		config.limit = limit
		pool := limitedReaderPool(config)

		return func(next lessgo.HandlerFunc) lessgo.HandlerFunc {
			return func(c lessgo.Context) error {
				req := c.Request()
				r := pool.Get().(*limitedReader)
				r.Reset(req.Body, c)
				defer pool.Put(r)
				req.SetBody(r)
				return next(c)
			}
		}
	},
}.Reg()

func (r *limitedReader) Read(b []byte) (n int, err error) {
	n, err = r.reader.Read(b)
	r.read += int64(n)
	if r.read > r.limit {
		s := http.StatusRequestEntityTooLarge
		r.context.String(s, http.StatusText(s))
		return n, lessgo.ErrStatusRequestEntityTooLarge
	}
	return
}

func (r *limitedReader) Reset(reader io.Reader, context lessgo.Context) {
	r.reader = reader
	r.context = context
}

func limitedReaderPool(c BodyLimitConfig) sync.Pool {
	return sync.Pool{
		New: func() interface{} {
			return &limitedReader{BodyLimitConfig: c}
		},
	}
}

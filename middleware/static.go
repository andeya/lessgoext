package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/lessgo/lessgo"
)

type (
	// StaticConfig defines the config for static middleware.
	StaticConfig struct {
		// Root is the directory from where the static content is served.
		// Required.
		Root string `json:"root"`

		// Index is the list of index files to be searched and used when serving
		// a directory.
		// Optional, with default value as []string{"index.html"}.
		Index []string `json:"index"`

		// Browse is a flag to enable/disable directory browsing.
		// Optional, with default value as false.
		Browse bool `json:"browse"`
	}
)

var (
	// DefaultStaticConfig is the default static middleware config.
	DefaultStaticConfig = StaticConfig{
		Index:  []string{"index.html"},
		Browse: true,
	}
)

// Static returns a static middleware to serves static content from the provided
// root directory.
func Static(configJSON string) lessgo.MiddlewareFunc {
	config := StaticConfig{}
	json.Unmarshal([]byte(configJSON), &config)

	// Defaults
	if config.Index == nil {
		config.Index = DefaultStaticConfig.Index
	}

	return func(next lessgo.HandlerFunc) lessgo.HandlerFunc {
		return func(c lessgo.Context) error {
			fs := http.Dir(config.Root)
			p := c.Request().URL.Path
			if strings.Contains(c.Path(), "*") { // If serving from a group, e.g. `/static*`.
				p = c.P(0)
			}
			file := path.Clean(p)
			f, err := fs.Open(file)
			if err != nil {
				return next(c)
			}
			defer f.Close()
			fi, err := f.Stat()
			if err != nil {
				return err
			}

			if fi.IsDir() {
				/* NOTE:
				Not checking the Last-Modified header as it caches the response `304` when
				changing different directories for the same path.
				*/
				d := f

				// Index file
				// TODO: search all files
				file = path.Join(file, config.Index[0])
				f, err = fs.Open(file)
				if err != nil {
					if config.Browse {
						dirs, err := d.Readdir(-1)
						if err != nil {
							return err
						}
						// Create a directory index
						res := c.Response()
						res.Header().Set(lessgo.HeaderContentType, lessgo.MIMETextHTMLCharsetUTF8)

						var list string
						prefix := c.Request().URL.Path
						for _, d := range dirs {
							name := d.Name()
							color := "#212121"
							if d.IsDir() {
								color = "#e91e63"
								name += "/"
							}
							list += fmt.Sprintf("<p><a href=\"%s\" style=\"color: %s;\">%s</a></p>\n", path.Join(prefix, name), color, name)
						}
						_, err = res.Write([]byte(list))
						if err == nil {
							return nil
						}
					}
					return err
				}
				defer f.Close()
				if fi, err = f.Stat(); err != nil { // Index file
					return err
				}
			}
			return c.ServeContent(f, fi.Name(), fi.ModTime())
		}
	}
}

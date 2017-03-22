package static

import (
	"net/http"
	"strings"

	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/labstack/echo"
)

type binaryFileSystem struct {
	fs http.FileSystem
}

func (b *binaryFileSystem) Open(name string) (http.File, error) {
	return b.fs.Open(name)
}

func (b *binaryFileSystem) Exists(prefix string, filepath string) bool {
	if p := strings.TrimPrefix(filepath, prefix); len(p) < len(filepath) {
		if _, err := b.Open(p); err != nil {
			return false
		}
		return true
	}
	return false
}

func Static(urlPrefix string, afs *assetfs.AssetFS, fallbackHandler echo.HandlerFunc) echo.MiddlewareFunc {
	fs := &binaryFileSystem{afs}
	fileserver := http.FileServer(fs)
	if urlPrefix != "" {
		fileserver = http.StripPrefix(urlPrefix, fileserver)
	}

	return func(before echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := before(c)
			if err != nil {
				if c, ok := err.(*echo.HTTPError); !ok || c.Code != http.StatusNotFound {
					return err
				}
			}

			if c.Response().Committed {
				// already sent response
				return nil
			}

			w, r := c.Response(), c.Request()
			if fs.Exists(urlPrefix, r.URL.Path) {
				fileserver.ServeHTTP(w, r)
				return nil
			} else if fallbackHandler != nil {
				return fallbackHandler(c)
			}

			return err
		}
	}
}

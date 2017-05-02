package static

import (
	"net/http"
	"strings"

	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
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

type StaticConfig struct {
	UrlPrefix string
	AssetFS   *assetfs.AssetFS
	Skipper   middleware.Skipper
}

func StaticWithConfig(config StaticConfig) echo.MiddlewareFunc {
	afs := config.AssetFS
	urlPrefix := config.UrlPrefix

	if config.Skipper == nil {
		config.Skipper = middleware.DefaultSkipper
	}

	fs := &binaryFileSystem{afs}
	fileserver := http.FileServer(fs)
	if urlPrefix != "" {
		fileserver = http.StripPrefix(urlPrefix, fileserver)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			w, r := c.Response(), c.Request()
			if fs.Exists(urlPrefix, r.URL.Path) {
				fileserver.ServeHTTP(w, r)
				return nil
			}

			return next(c)
		}
	}
}

func Static(urlPrefix string, afs *assetfs.AssetFS) echo.MiddlewareFunc {
	c := StaticConfig{}
	c.UrlPrefix = urlPrefix
	c.AssetFS = afs
	c.Skipper = middleware.DefaultSkipper

	return StaticWithConfig(c)
}

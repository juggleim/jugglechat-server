package routers

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed webim
var webimFS embed.FS

func LoadWebIM(eng *gin.Engine) {
	subFS, err := fs.Sub(webimFS, "webim")
	if err != nil {
		panic(err)
	}
	httpFS := http.FS(subFS)

	serveIndex := func(ctx *gin.Context) {
		serveIndexFile(ctx, subFS)
	}
	eng.GET("/", serveIndex)
	eng.HEAD("/", serveIndex)

	eng.NoRoute(func(ctx *gin.Context) {
		if ctx.Request.Method != http.MethodGet && ctx.Request.Method != http.MethodHead {
			ctx.Status(http.StatusNotFound)
			return
		}

		filePath := strings.TrimPrefix(ctx.Request.URL.Path, "/")
		if filePath == "" {
			filePath = "index.html"
		}

		if fileExists(subFS, filePath) {
			ctx.FileFromFS(filePath, httpFS)
			return
		}
		serveIndexFile(ctx, subFS)
	})
}

func serveIndexFile(ctx *gin.Context, fileSystem fs.FS) {
	data, err := fs.ReadFile(fileSystem, "index.html")
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", data)
}

func fileExists(fileSystem fs.FS, name string) bool {
	file, err := fileSystem.Open(name)
	if err != nil {
		return false
	}
	defer file.Close()

	stat, err := file.Stat()
	return err == nil && !stat.IsDir()
}

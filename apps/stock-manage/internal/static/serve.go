package static

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Register mounts embedded frontend assets and SPA fallback on non-API routes.
func Register(r *gin.Engine) {
	sub, err := fs.Sub(Dist, "dist")
	if err != nil {
		panic(err)
	}

	fileServer := http.FileServer(http.FS(sub))

	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		path := strings.TrimPrefix(c.Request.URL.Path, "/")
		if path == "" {
			serveIndex(c, fileServer)
			return
		}

		if _, err := fs.Stat(sub, path); err != nil {
			serveIndex(c, fileServer)
			return
		}

		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}

func serveIndex(c *gin.Context, fileServer http.Handler) {
	c.Request.URL.Path = "/"
	fileServer.ServeHTTP(c.Writer, c.Request)
}

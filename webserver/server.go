package webserver

import (
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"pluto/global"
	"pluto/util"
	"strconv"

	"github.com/gin-gonic/gin"
)

func Launch(frontendFS embed.FS) error {
	gin.DefaultWriter = &util.SlogWriter{Level: slog.LevelInfo}
	gin.DefaultErrorWriter = &util.SlogWriter{Level: slog.LevelError}
	g := gin.Default()

	//Backend APIs
	api := g.Group("/api")
	api.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version": global.Version,
		})
	})
	//Frontend files
	frontendFp, _ := fs.Sub(frontendFS, "frontend")
	// 所有请求先匹配 api 路由，如果没匹配到就当作静态资源文件处理
	g.NoRoute(gin.WrapH(http.FileServer(http.FS(frontendFp))))

	initMappingApis(api)
	initSourceApi(api)
	err := g.Run(":" + strconv.Itoa(global.Config.Port))
	return err
}

package webserver

import (
	"log/slog"
	"net/http"
	"path/filepath"
	"pluto/global"
	"pluto/util"
	"strconv"

	"github.com/gin-gonic/gin"
)

func Launch() error {
	gin.DefaultWriter = &util.SlogWriter{Level: slog.LevelInfo}
	gin.DefaultErrorWriter = &util.SlogWriter{Level: slog.LevelError}
	g := gin.Default()

	//Backend APIs
	api := g.Group("/api")
	api.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version": global.Version,
		})
	})
	//Frontend files
	g.NoRoute(func(c *gin.Context) {
		// 获取请求路径
		path := c.Request.URL.Path

		// 如果是文件（如 /css/app.css、/js/index.js），尝试返回文件
		if filepath.Ext(path) != "" {
			c.File("frontend" + path)
			return
		}

		// 不是文件 → 返回前端首页（Vue/React/HTML 项目）
		c.File("frontend/index.html")
	})

	initMappingApis(api)
	initSourceApi(api)
	err := g.Run(":" + strconv.Itoa(global.Config.Port))
	return err
}

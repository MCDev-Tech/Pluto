package webserver

import (
	"net/http"
	"os"
	"pluto/mapping"
	"pluto/util"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func handleSourceGet(c *gin.Context) {
	mcVersion, mappingType, class := c.Query("version"), c.Query("type"), c.Query("class")
	if mcVersion == "" || mappingType == "" || class == "" {
		c.String(http.StatusBadRequest, "Missing query parameter(s)")
		return
	}

	class = strings.TrimSpace(class)
	class = strings.Trim(class, "/")
	class = strings.ReplaceAll(class, ".", "/")

	targetPath, err := mapping.GenerateSourceForClass(mcVersion, mappingType, class)
	if err != nil {
		c.String(http.StatusInternalServerError, "Generate single class source failed: "+err.Error())
		return
	}

	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		c.String(http.StatusNotFound, "")
		return
	}

	content, err := os.ReadFile(targetPath)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to read file")
		return
	}
	c.Data(http.StatusOK, "text/plain; charset=utf-8", content)
}

func initSourceApi(g *gin.RouterGroup) {
	g.GET("/source/decompile", RateLimiterMiddleware(10*time.Second, 2), func(c *gin.Context) {
		mcVersion, mappingType := c.Query("version"), c.Query("type")
		if mcVersion == "" || mappingType == "" {
			c.String(http.StatusBadRequest, "Missing query parameter(s)")
			return
		}
		if mapping.IsAvailable(mcVersion, mappingType) {
			c.String(http.StatusOK, "Decompiled")
			return
		}
		if mapping.IsPending(mcVersion, mappingType) {
			c.String(http.StatusForbidden, "This task is pending")
			return
		}
		util.Execute(func() error {
			_, err := mapping.GenerateSource(mcVersion, mappingType)
			return err
		})
		c.String(http.StatusAccepted, "Started decompiling, please wait")
	})
	g.GET("/source/get", RateLimiterMiddleware(2*time.Second, 5), handleSourceGet)
}

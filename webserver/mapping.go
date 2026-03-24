package webserver

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pluto/mapping"
	"time"
)

func initMappingApis(g *gin.RouterGroup) {
	g.GET("/mapping/search", RateLimiterMiddleware(2*time.Second, 5), func(c *gin.Context) {
		mcVersion, mappingType, keyword, translate := c.Query("version"), c.Query("type"), c.Query("keyword"), c.Query("translate")
		if mcVersion == "" || mappingType == "" || keyword == "" {
			c.String(http.StatusBadRequest, "Missing query parameter(s)")
			return
		}
		if len(keyword) <= 2 {
			c.String(http.StatusBadRequest, "Keyword must contain at least three characters")
			return
		}
		mappings, err := mapping.LoadMapping(mcVersion, mappingType)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		results := mappings.Search(keyword, 20)
		if translate != "" {
			mappings, err := mapping.LoadMapping(mcVersion, translate)
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			mappings.AppendTranslate(&results)
		}
		c.JSON(http.StatusOK, results)
	})
}

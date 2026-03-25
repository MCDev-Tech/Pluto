package webserver

import (
	"net/http"
	"pluto/mapping"
	"pluto/mapping/java"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func initMappingApis(g *gin.RouterGroup) {
	g.GET("/mapping/search", RateLimiterMiddleware(2*time.Second, 5), func(c *gin.Context) {
		mcVersion, mappingType, keyword, filterRaw, translate := c.Query("version"), c.Query("type"), c.Query("keyword"), c.Query("filter"), c.Query("translate")
		if mcVersion == "" || mappingType == "" || keyword == "" {
			c.String(http.StatusBadRequest, "Missing query parameter(s)")
			return
		}
		if len(keyword) <= 2 {
			c.String(http.StatusBadRequest, "Keyword must contain at least three characters")
			return
		}
		filter, err := strconv.ParseInt(filterRaw, 10, 0)
		if err != nil {
			c.String(http.StatusBadRequest, "Wrong Filter params, should be integer!")
			return
		}
		mappings, err := mapping.LoadMapping(mcVersion, mappingType)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		results := mappings.Search(keyword, 50, &java.Filter{Key: filter})
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

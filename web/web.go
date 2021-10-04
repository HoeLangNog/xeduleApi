package web

import (
	"github.com/gin-gonic/gin"
	"os"
)

var router *gin.Engine

func init() {

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Cache-Control", "max-age=300, public")
	})

	locations := r.Group("/locations")
	groups := r.Group("/groups")
	registerGroupsEndpoints(groups)
	registerLocations(locations)
	RegisterOldEndpoints(r)
	router = r
}

func Start() {
	router.Run(os.Getenv("address"))
}

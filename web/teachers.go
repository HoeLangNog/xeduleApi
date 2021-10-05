package web

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"jaapie/xscheduleapi/xschedule"
)

func registerTeachersEndpoints(r *gin.RouterGroup) {
	r.GET("", handleGetAllTeachers)
}

func handleGetAllTeachers(c *gin.Context) {
	teachers := xschedule.GetAllTeachers()

	e := json.NewEncoder(c.Writer)

	err := e.Encode(teachers)

	if err != nil {
		c.AbortWithStatusJSON(500, map[string]string{
			"error": "Failed while encoding json",
		})
	}
}

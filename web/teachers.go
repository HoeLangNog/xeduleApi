package web

import (
	"github.com/gin-gonic/gin"
	"jaapie/xscheduleapi/xschedule"
)

func registerTeachersEndpoints(r *gin.RouterGroup) {
	r.GET("", handleGetAllTeachers)
}

func handleGetAllTeachers(c *gin.Context) {

	xschedule.GetAllTeachers()
}

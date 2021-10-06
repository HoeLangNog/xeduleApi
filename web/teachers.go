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

	err := e.Encode(translateTeachers(teachers))

	if err != nil {
		c.AbortWithStatusJSON(500, map[string]string{
			"error": "Failed while encoding json",
		})
	}
}

type webTeacher struct {
	Code      string  `json:"code"`
	Id        string  `json:"id"`
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
}

func translateTeachers(teachers []*xschedule.Teacher) []*webTeacher {
	var newTeachers []*webTeacher
	for _, teacher := range teachers {
		fi, l, f := xschedule.GetTeacherName(teacher.Code)
		if f {

			firstName := &fi
			lastName := &l

			if *firstName == "none" {
				firstName = nil
			}
			if *lastName == "none" {
				lastName = nil
			}

			newTeachers = append(newTeachers, &webTeacher{
				Code:      teacher.Code,
				Id:        teacher.Id,
				FirstName: firstName,
				LastName:  lastName,
			})
		} else {
			newTeachers = append(newTeachers, &webTeacher{
				Code: teacher.Code,
				Id:   teacher.Id,
			})
		}
	}
	return newTeachers
}

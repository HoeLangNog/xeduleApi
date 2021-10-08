package web

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"jaapie/xscheduleapi/xschedule"
	"net/http"
	"strconv"
	"time"
)

func registerGroupsEndpoints(r *gin.RouterGroup) {

	r.GET("", handleGetAllGroups)

	r.GET("/:groupCode", handleGetGroup)
	r.GET("/:groupCode/schedule", handleGetGroupSchedule)
}

func handleGetAllGroups(c *gin.Context) {
	groups := xschedule.GetAllGroups()

	_, v := c.GetQuery("visible")

	e := json.NewEncoder(c.Writer)

	if v {
		var newGroups []*webGroupResponse
		for _, g := range translateGroups(groups...) {
			if g.Visible {
				newGroups = append(newGroups, g)
			}
		}

		err := e.Encode(newGroups)
		if err != nil {
			c.AbortWithStatusJSON(500, map[string]string{
				"error": "Failed encoding to json",
			})
			return
		}

	} else {
		err := e.Encode(translateGroups(groups...))
		if err != nil {
			c.AbortWithStatusJSON(500, map[string]string{
				"error": "Failed encoding to json",
			})
			return
		}
	}

}

func handleGetGroup(c *gin.Context) {

	groupCode := c.Param("groupCode")

	group := xschedule.GetGroup(groupCode)

	if group == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, map[string]string{
			"error": "Did not find group",
		})
	}

	e := json.NewEncoder(c.Writer)

	err := e.Encode(group)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed encoding json",
		})
		return
	}
}

func handleGetGroupSchedule(c *gin.Context) {
	yearString := c.Query("year")
	weekString := c.Query("week")

	nowYear, nowWeek := time.Now().ISOWeek()

	year, err := strconv.Atoi(yearString)
	if err != nil {
		year = nowYear
	}

	week, err := strconv.Atoi(weekString)
	if err != nil {
		week = nowWeek
	}
	groupCode := c.Param("groupCode")

	group := xschedule.GetGroup(groupCode)

	if group == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, map[string]string{
			"error": "Did not find group",
		})
		return
	}

	schedule := xschedule.GetSchedule(&xschedule.TimeSelector{
		Id:   group.Id,
		Year: year,
		Week: week,
		Orus: group.Orus[len(group.Orus)-1],
	})

	if len(schedule) == 0 {
		c.JSON(200, []string{})
		return
	}

	res := schedule[0]

	e := json.NewEncoder(c.Writer)
	err = e.Encode(translateSchedule(res.Apps))

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed encoding json",
		})
		return
	}
}

type webGroupResponse struct {
	Code    string `json:"code"`
	Id      string `json:"id"`
	Visible bool   `json:"visible"`
}

func translateGroups(groups ...*xschedule.Group) []*webGroupResponse {
	var responses []*webGroupResponse

	orIds := xschedule.OrganizationIds()
	for _, group := range groups {
		visible := false

		for _, orus := range group.Orus {

			for _, id := range orIds {
				if id == orus {
					visible = true
					break
				}
			}

		}
		responses = append(responses, &webGroupResponse{
			Code:    group.Code,
			Id:      group.Id,
			Visible: visible,
		})
	}
	return responses
}

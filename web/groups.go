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
	r.GET("/:groupCode/unixoftoday", handleGetGroupUnix)
}

func handleGetGroupUnix(c *gin.Context) {
	groupCode := c.Param("groupCode")

	group := xschedule.GetGroup(groupCode)
	year, week := time.Now().ISOWeek()
	res := xschedule.GetSchedule(&xschedule.TimeSelector{
		Id:   group.Id,
		Year: year,
		Week: week,
		Orus: group.Orus[len(group.Orus)-1],
	})

	if len(res) == 0 {
		c.AbortWithStatus(404)
		return
	}

	scheduleResponse := res[0]

	a := &oldUnixOfTodayResponse{}

	if len(scheduleResponse.Apps) == 0 {
		a.Start = 0
		a.Last = 0
	} else {
		smallestBegin := int64(545645645645678976)
		largestEnd := int64(-1)
		for _, res := range scheduleResponse.Apps {
			sTime, eTime := res.GetDates()
			if sTime.Day() == time.Now().Day() {
				if smallestBegin > sTime.Unix() {
					smallestBegin = sTime.Unix()
				}
				if largestEnd < eTime.Unix() {
					largestEnd = eTime.Unix()
				}
			}
		}
		a.Start = smallestBegin
		a.Last = largestEnd

	}

	e := json.NewEncoder(c.Writer)

	err := e.Encode(a)

	if err != nil {
		c.AbortWithStatusJSON(500, map[string]string{
			"error": "Failed to encode json",
		})
		return
	}
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
	_, today := c.GetQuery("today")

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

	if today {
		apps := res.Apps
		res.Apps = []*xschedule.TimeSlot{}

		time2 := time.Now()

		for _, app := range apps {
			sTime, _ := app.GetDates()

			if sTime.Day() == time2.Day() {
				res.Apps = append(res.Apps, app)
			}
		}
	}

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

package web

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"jaapie/xscheduleapi/xschedule"
	"sort"
	"strconv"
	"time"
)

func RegisterOldEndpoints(r *gin.Engine) {

	r.GET("/unixoftoday", handleUnixOfToday)
	r.GET("/today", handleTodayEndpoint)
}

type oldUnixOfTodayResponse struct {
	Start int64 `json:"start"`
	Last  int64 `json:"last"`
}

func handleUnixOfToday(c *gin.Context) {
	group := xschedule.GetGroup("TTB4-SSD2C")
	year, week := time.Now().ISOWeek()
	res := xschedule.GetSchedule(&xschedule.TimeSelector{
		Id:   group.Id,
		Year: year,
		Week: week,
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

type oldTodayResponse struct {
	Index    int    `json:"index"`
	Id       string `json:"id"`
	Title    string `json:"title"`
	Date     string `json:"date"`
	Day      int    `json:"day"`
	Time     int64  `json:"time"`
	TimeEnd  int64  `json:"timeEnd"`
	Facility string `json:"facility"`
	Docent   string `json:"docent"`
	Group    string `json:"group"`
}

func handleTodayEndpoint(c *gin.Context) {
	group := xschedule.GetGroup("TTB4-SSD2C")

	year, week := time.Now().ISOWeek()
	res := xschedule.GetSchedule(&xschedule.TimeSelector{
		Id:   group.Id,
		Year: year,
		Week: week,
	})

	if len(res) == 0 {
		c.AbortWithStatus(404)
		return
	}

	scheduleResponse := res[0]

	actualResponse := []*oldTodayResponse{}
	i := 0

	sort.SliceStable(scheduleResponse.Apps, func(i, j int) bool {
		s, _ := scheduleResponse.Apps[i].GetDates()
		s2, _ := scheduleResponse.Apps[j].GetDates()
		return s.Unix() < s2.Unix()
	})
	for _, app := range scheduleResponse.Apps {
		if a, e := app.GetDates(); a.Day() == time.Now().Day() {
			fmt.Println(app.Attributes)

			locationCode := ""
			teacherCode := ""
			classCode := ""
			for _, a1 := range app.Attributes {
				if locationCode == "" {
					location := xschedule.GetLocationById(strconv.Itoa(a1))
					if location != nil {
						locationCode = location.Code
						continue
					}
				}

				if teacherCode == "" {
					teacher := xschedule.GetTeacherById(strconv.Itoa(a1))
					if teacher != nil {
						teacherCode = teacher.Code
						continue
					}
				}

				if classCode == "" {
					class := xschedule.GetGroupById(strconv.Itoa(a1))
					if class != nil {
						classCode = class.Code
						continue
					}
				}
			}

			actualResponse = append(actualResponse, &oldTodayResponse{
				Index:    i,
				Id:       app.Id,
				Title:    app.Name,
				Date:     strconv.Itoa(a.Day()) + "-" + strconv.Itoa(int(a.Month())),
				Day:      int(a.Weekday()),
				Time:     a.Unix(),
				TimeEnd:  e.Unix(),
				Facility: locationCode,
				Docent:   teacherCode,
				Group:    classCode,
			})
			i++
		}
	}

	e := json.NewEncoder(c.Writer)
	err := e.Encode(actualResponse)

	if err != nil {
		c.AbortWithStatusJSON(500, map[string]string{
			"error": "Failed encoding json",
		})
	}
}

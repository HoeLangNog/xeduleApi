package web

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"jaapie/xscheduleapi/xschedule"
	"net/http"
	"strconv"
	"time"
)

func registerLocations(r *gin.RouterGroup) {
	r.Use(func(c *gin.Context) {
		c.Header("Cache-Control", "private max-age=900")
	})
	r.GET("", handleGetAllLocations)
	r.GET("/available/:prefix/", getAvailableLocations)
	r.GET("/:id/schedule", handleGetLocationSchedule)
}

func handleGetAllLocations(c *gin.Context) {

	e := json.NewEncoder(c.Writer)
	err := e.Encode(translateLocations(xschedule.GetAllLocations()))
	if err != nil {
		fmt.Println(err)
		c.AbortWithStatusJSON(500, map[string]string{
			"error": "Couldn't encode to json",
		})
	}
}

func getAvailableLocations(c *gin.Context) {
	prefix := c.Param("prefix")
	sTimeUnix := c.Query("starttime")
	eTimeUnix := c.Query("endtime")
	sTimeU, err := strconv.ParseInt(sTimeUnix, 10, 64)
	eTimeU, err := strconv.ParseInt(eTimeUnix, 10, 64)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, map[string]string{
			"error": "Bad query timestamps",
		})
	}

	sTime := time.Unix(sTimeU, 0)
	eTime := time.Unix(eTimeU, 0)

	availableLocations := translateLocations(xschedule.GetAvailableLocations(prefix, sTime, eTime))

	e := json.NewEncoder(c.Writer)
	err = e.Encode(availableLocations)

	if err != nil {
		c.AbortWithStatusJSON(500, map[string]string{
			"error": "Failed to encode to json",
		})
	}
}

func handleGetLocationSchedule(c *gin.Context) {
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

	responses := xschedule.GetSchedule(&xschedule.TimeSelector{
		Id:   c.Param("id"),
		Year: year,
		Week: week,
		Orus: 15,
	})

	if len(responses) == 0 {

		return
	}

	newSlots := translateSchedule(responses[0].Apps)
	e := json.NewEncoder(c.Writer)
	err = e.Encode(newSlots)

	if err != nil {
		c.AbortWithStatusJSON(500, map[string]string{
			"error": "Failed to encode to json",
		})
	}
}

type webLocation struct {
	Code    string `json:"code"`
	Id      string `json:"id"`
	Visible bool   `json:"visible"`
}

func translateLocations(locations []*xschedule.Location) []*webLocation {
	var newLocs []*webLocation

	orIds := xschedule.OrganizationIds()
	for _, loc := range locations {
		visible := false
		for _, orus := range loc.Orus {
			for _, id := range orIds {
				if id == orus {
					visible = true
					break
				}
			}

		}
		newLocs = append(newLocs, &webLocation{
			loc.Code,
			loc.Id,
			visible,
		})
	}
	return newLocs
}

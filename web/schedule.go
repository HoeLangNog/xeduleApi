package web

import (
	"jaapie/xscheduleapi/xschedule"
	"strconv"
)

type webTimeSlot struct {
	Name      string      `json:"name"`
	Summary   string      `json:"summary"`
	Attention string      `json:"attention"`
	StartTime int64       `json:"start_time"`
	EndTime   int64       `json:"end_time"`
	Group     string      `json:"group"`
	Teacher   *webTeacher `json:"teacher"`
	Location  string      `json:"location"`
}

func translateSchedule(responses []*xschedule.TimeSlot) []*webTimeSlot {

	var newSlots []*webTimeSlot
	for _, response := range responses {
		locationCode := ""
		var teacher *webTeacher
		classCode := ""
		for _, a1 := range response.Attributes {
			if locationCode == "" {
				location := xschedule.GetLocationById(strconv.Itoa(a1))
				if location != nil {
					locationCode = location.Code
					continue
				}
			}

			if teacher == nil {
				teacher1 := xschedule.GetTeacherById(strconv.Itoa(a1))
				if teacher1 != nil {
					teacher = translateTeachers([]*xschedule.Teacher{teacher1})[0]
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
		sTime, eTime := response.GetDates()
		newSlots = append(newSlots, &webTimeSlot{
			Name:      response.Name,
			Summary:   response.Summary,
			Attention: response.Attention,
			StartTime: sTime.Unix(),
			EndTime:   eTime.Unix(),
			Group:     classCode,
			Teacher:   teacher,
			Location:  locationCode,
		})
	}
	return newSlots
}

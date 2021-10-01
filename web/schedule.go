package web

import (
	"jaapie/xscheduleapi/xschedule"
)

type webTimeSlot struct {
	Name string `json:"name"`
	Summary string `json:"summary"`
	Attention string `json:"attention"`
	StartTime int64 `json:"start_time"`
	EndTime int64 `json:"end_time"`
}

func translateSchedule(responses []*xschedule.TimeSlot) []*webTimeSlot {

	var newSlots []*webTimeSlot
	for _, response := range responses {

		sTime, eTime := response.GetDates()
		newSlots = append(newSlots, &webTimeSlot{
			Name: response.Name,
			Summary: response.Summary,
			Attention: response.Attention,
			StartTime: sTime.Unix(),
			EndTime: eTime.Unix(),
		})
	}
	return newSlots
}